package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"data-gen/conf"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

const (
	defaultStreamCount = 1
	maxEventsPerPut    = 10_000  // CloudWatch PutLogEvents event count limit
	maxBytesPerPut     = 1_000_000 // Stay under CloudWatch's 1,048,576 byte limit (leave room for overhead)
	perEventOverhead   = 26      // CloudWatch adds 26 bytes overhead per event
)

// CloudWatchExporter sends generated data to AWS CloudWatch Logs.
// When stream_count > 1, it auto-creates multiple log streams and
// round-robins batches across them for higher throughput.
type CloudWatchExporter struct {
	cfg              cwLogCfg
	cloudwatchClient *cloudwatchlogs.Client
	streams          []string
	nextStream       atomic.Uint64
}

type streamChunk struct {
	stream string
	events []types.InputLogEvent
}

// cwLogCfg specifies the CloudWatch log group, stream prefix, and parallelism.
type cwLogCfg struct {
	LogGroupName  string `yaml:"log_group"`
	LogStreamName string `yaml:"log_stream"`
	StreamCount   int    `yaml:"stream_count"`
}

func NewCloudWatchLogExporter(ctx context.Context, c *conf.Config) (*CloudWatchExporter, error) {
	var cfg cwLogCfg
	err := c.Output.Conf.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	if v := os.Getenv(conf.EnvOutLogGroup); v != "" {
		cfg.LogGroupName = v
	}
	if v := os.Getenv(conf.EnvOutLogStream); v != "" {
		cfg.LogStreamName = v
	}

	if cfg.LogGroupName == "" || cfg.LogStreamName == "" {
		return nil, fmt.Errorf("cloudwatch log group and/or stream name must be specified for output type %s", c.Output.Type)
	}

	if cfg.StreamCount <= 0 {
		cfg.StreamCount = defaultStreamCount
	}

	loadedAwsConfig, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(c.Profile), config.WithRegion(c.Region))
	if err != nil {
		return nil, fmt.Errorf("failed to load default aws config: %w", err)
	}

	client := cloudwatchlogs.NewFromConfig(loadedAwsConfig)

	streams, err := ensureStreams(ctx, client, cfg)
	if err != nil {
		return nil, err
	}

	return &CloudWatchExporter{
		cfg:              cfg,
		cloudwatchClient: client,
		streams:          streams,
	}, nil
}

// ensureStreams creates the log group (if needed) and the required log streams.
// For stream_count=1 the stream name is used as-is; for N>1 streams are
// named "<log_stream>-0", "<log_stream>-1", etc.
func ensureStreams(ctx context.Context, client *cloudwatchlogs.Client, cfg cwLogCfg) ([]string, error) {
	_, err := client.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(cfg.LogGroupName),
	})
	if err != nil && !isAlreadyExists(err) {
		return nil, fmt.Errorf("failed to create log group %s: %w", cfg.LogGroupName, err)
	}

	streams := make([]string, cfg.StreamCount)
	for i := 0; i < cfg.StreamCount; i++ {
		name := cfg.LogStreamName
		if cfg.StreamCount > 1 {
			name = fmt.Sprintf("%s-%d", cfg.LogStreamName, i)
		}
		streams[i] = name

		_, err := client.CreateLogStream(ctx, &cloudwatchlogs.CreateLogStreamInput{
			LogGroupName:  aws.String(cfg.LogGroupName),
			LogStreamName: aws.String(name),
		})
		if err != nil && !isAlreadyExists(err) {
			return nil, fmt.Errorf("failed to create log stream %s: %w", name, err)
		}
	}

	slog.Info("CloudWatch streams ready",
		slog.String("log_group", cfg.LogGroupName),
		slog.Int("stream_count", cfg.StreamCount),
		slog.Any("streams", streams),
	)

	return streams, nil
}

func isAlreadyExists(err error) bool {
	var alreadyExists *types.ResourceAlreadyExistsException
	return errors.As(err, &alreadyExists)
}

func (ce *CloudWatchExporter) Send(data *[]byte) error {
	now := time.Now().UnixMilli()

	lines := strings.Split(strings.TrimRight(string(*data), "\n"), "\n")
	logEvents := make([]types.InputLogEvent, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		logEvents = append(logEvents, types.InputLogEvent{
			Message:   aws.String(line),
			Timestamp: aws.Int64(now),
		})
	}

	if len(logEvents) == 0 {
		return nil
	}

	// Chunk by both event count (10,000) and payload size (~1MB),
	// then round-robin chunks across streams.
	var chunk []types.InputLogEvent
	chunkBytes := 0
	var chunks []streamChunk

	for _, event := range logEvents {
		eventSize := len(*event.Message) + perEventOverhead

		if len(chunk) >= maxEventsPerPut || (chunkBytes+eventSize > maxBytesPerPut && len(chunk) > 0) {
			eventsCopy := append([]types.InputLogEvent(nil), chunk...)
			chunks = append(chunks, streamChunk{stream: ce.nextStreamName(), events: eventsCopy})
			chunk = chunk[:0]
			chunkBytes = 0
		}

		chunk = append(chunk, event)
		chunkBytes += eventSize
	}

	if len(chunk) > 0 {
		eventsCopy := append([]types.InputLogEvent(nil), chunk...)
		chunks = append(chunks, streamChunk{stream: ce.nextStreamName(), events: eventsCopy})
	}

	return ce.putChunksParallel(chunks)
}

func (ce *CloudWatchExporter) nextStreamName() string {
	idx := ce.nextStream.Add(1) - 1
	return ce.streams[idx%uint64(len(ce.streams))]
}

func (ce *CloudWatchExporter) putChunk(stream string, events []types.InputLogEvent) error {
	_, err := ce.cloudwatchClient.PutLogEvents(context.Background(), &cloudwatchlogs.PutLogEventsInput{
		LogGroupName:  aws.String(ce.cfg.LogGroupName),
		LogStreamName: aws.String(stream),
		LogEvents:     events,
	})
	if err != nil {
		return fmt.Errorf("unable to write to cloudwatch log stream %s: %w", stream, err)
	}
	return nil
}

// putChunksParallel sends chunks concurrently across streams while preserving
// per-stream ordering (chunks targeting the same stream are sent sequentially).
func (ce *CloudWatchExporter) putChunksParallel(chunks []streamChunk) error {
	if len(chunks) == 0 {
		return nil
	}

	byStream := make(map[string][][]types.InputLogEvent)
	for _, c := range chunks {
		byStream[c.stream] = append(byStream[c.stream], c.events)
	}

	errCh := make(chan error, len(byStream))
	var wg sync.WaitGroup
	wg.Add(len(byStream))

	for stream, streamChunks := range byStream {
		go func(stream string, streamChunks [][]types.InputLogEvent) {
			defer wg.Done()
			for _, evs := range streamChunks {
				if err := ce.putChunk(stream, evs); err != nil {
					errCh <- err
					return
				}
			}
		}(stream, streamChunks)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}
