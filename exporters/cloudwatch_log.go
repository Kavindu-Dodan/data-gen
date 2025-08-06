package exporters

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"data-gen/conf"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

type CloudWatchExporter struct {
	cfg              cwLogCfg
	cloudwatchClient *cloudwatchlogs.Client
	shChan           chan struct{}
}

type cwLogCfg struct {
	LogGroupName  string `yaml:"logGroup"`
	LogStreamName string `yaml:"logStream"`
}

func newCloudWatchLogExporter(ctx context.Context, conf *conf.Config) (*CloudWatchExporter, error) {
	var cfg cwLogCfg
	err := conf.Output.Conf.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	if cfg.LogGroupName == "" || cfg.LogStreamName == "" {
		return nil, fmt.Errorf("cloudwatch log group and/or stream name must be specified for output type %s", conf.Output.Type)
	}

	loadedAwsConfig, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(conf.AWSCfg.Profile), config.WithRegion(conf.AWSCfg.Region))
	if err != nil {
		return nil, fmt.Errorf("failed to load default aws config: %w", err)
	}

	cloudwatchClient := cloudwatchlogs.NewFromConfig(loadedAwsConfig)

	return &CloudWatchExporter{
		cfg:              cfg,
		cloudwatchClient: cloudwatchClient,
		shChan:           make(chan struct{}),
	}, nil
}

func (ce CloudWatchExporter) Start(data <-chan *[]byte) <-chan error {
	errChan := make(chan error)

	go func() {
		for {
			select {
			case d := <-data:
				record := cloudwatchlogs.PutLogEventsInput{
					LogGroupName:  aws.String(ce.cfg.LogGroupName),
					LogStreamName: aws.String(ce.cfg.LogStreamName),
					LogEvents: []types.InputLogEvent{
						{
							Message:   aws.String(string(*d)),
							Timestamp: aws.Int64(time.Now().UnixMilli()),
						},
					},
				}

				_, err := ce.cloudwatchClient.PutLogEvents(context.Background(), &record)
				if err != nil {
					errChan <- fmt.Errorf("unable to write to cloudwatch log group  %s: %w", ce.cfg.LogGroupName, err)
					return
				}
			case <-ce.shChan:
				slog.Info("shutting down cloudwatch exporter")
				return
			}

		}
	}()

	return errChan
}

func (ce CloudWatchExporter) Stop() {
	close(ce.shChan)
}
