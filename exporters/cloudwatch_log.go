package exporters

import (
	"context"
	"fmt"
	"os"
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
	LogGroupName  string `yaml:"log_group"`
	LogStreamName string `yaml:"log_stream"`
}

func newCloudWatchLogExporter(ctx context.Context, c *conf.Config) (*CloudWatchExporter, error) {
	var cfg cwLogCfg
	err := c.Output.Conf.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	// load env variable overrides if any
	if v := os.Getenv(conf.EnvOutLogGroup); v != "" {
		cfg.LogGroupName = v
	}
	if v := os.Getenv(conf.EnvOutLogStream); v != "" {
		cfg.LogStreamName = v
	}

	if cfg.LogGroupName == "" || cfg.LogStreamName == "" {
		return nil, fmt.Errorf("cloudwatch log group and/or stream name must be specified for output type %s", c.Output.Type)
	}

	loadedAwsConfig, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(c.AWSCfg.Profile), config.WithRegion(c.AWSCfg.Region))
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

func (ce CloudWatchExporter) send(data *[]byte) error {
	record := cloudwatchlogs.PutLogEventsInput{
		LogGroupName:  aws.String(ce.cfg.LogGroupName),
		LogStreamName: aws.String(ce.cfg.LogStreamName),
		LogEvents: []types.InputLogEvent{
			{
				Message:   aws.String(string(*data)),
				Timestamp: aws.Int64(time.Now().UnixMilli()),
			},
		},
	}

	_, err := ce.cloudwatchClient.PutLogEvents(context.Background(), &record)
	if err != nil {
		return fmt.Errorf("unable to write to cloudwatch log group  %s: %w", ce.cfg.LogGroupName, err)
	}

	return nil
}

func (ce CloudWatchExporter) stop() {
	close(ce.shChan)
}
