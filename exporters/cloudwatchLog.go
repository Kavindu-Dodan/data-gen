package exporters

import (
	"context"
	"data-gen/conf"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

type CloudWatchExporter struct {
	logGroup  string
	logStream string

	cloudwatchClient *cloudwatchlogs.Client
	shChan           chan struct{}
}

func NewCloudWatchLogExporter(ctx context.Context, awsConfig conf.AWSCfg) (*CloudWatchExporter, error) {
	if awsConfig.CloudwatchLogGroup == "" || awsConfig.CloudwatchLogStreamName == "" {
		return nil, fmt.Errorf("cloudwatch log group and/or stream name must be specified for output type CLOUDWATCH")
	}

	loadedAwsConfig, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(awsConfig.Profile), config.WithRegion(awsConfig.Region))
	if err != nil {
		return nil, fmt.Errorf("failed to load default aws config: %w", err)
	}

	cloudwatchClient := cloudwatchlogs.NewFromConfig(loadedAwsConfig)

	return &CloudWatchExporter{
		logGroup:         awsConfig.CloudwatchLogGroup,
		logStream:        awsConfig.CloudwatchLogStreamName,
		cloudwatchClient: cloudwatchClient,
		shChan:           make(chan struct{}),
	}, nil
}

func (ce CloudWatchExporter) Start(data <-chan []byte, errChan chan error) {
	go func() {
		for {
			select {
			case d := <-data:
				record := cloudwatchlogs.PutLogEventsInput{
					LogGroupName:  aws.String(ce.logGroup),
					LogStreamName: aws.String(ce.logStream),
					LogEvents: []types.InputLogEvent{
						{
							Message:   aws.String(string(d)),
							Timestamp: aws.Int64(time.Now().UnixMilli()),
						},
					},
				}

				_, err := ce.cloudwatchClient.PutLogEvents(context.Background(), &record)
				if err != nil {
					errChan <- fmt.Errorf("unable to write to cloudwatch log group  %s: %w", ce.logGroup, err)
					return
				}
			case <-ce.shChan:
				slog.Info("shutting down cloudwatch exporter")
				return
			}

		}
	}()

}

func (ce CloudWatchExporter) Stop() {
	close(ce.shChan)
}
