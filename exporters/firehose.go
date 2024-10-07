package exporters

import (
	"context"
	"data-gen/conf"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/aws/aws-sdk-go-v2/service/firehose/types"
)

type FirehoseExporter struct {
	streamName string
	client     *firehose.Client
	shChan     chan struct{}
}

func NewFirehoseExporter(ctx context.Context, awsConfig conf.AWSCfg) (*FirehoseExporter, error) {
	loadedAwsConfig, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(awsConfig.Profile), config.WithRegion(awsConfig.Region))
	if err != nil {
		return nil, fmt.Errorf("failed to load default aws config: %w", err)
	}

	fhClient := firehose.New(firehose.Options{
		Credentials: loadedAwsConfig.Credentials,
		Region:      awsConfig.Region,
	})

	return &FirehoseExporter{
		streamName: awsConfig.FirehoseStreamName,
		client:     fhClient,
		shChan:     make(chan struct{}),
	}, nil
}

func (f *FirehoseExporter) Start(c <-chan []byte, errChan chan error) {
	go func() {
		for {
			select {
			case d := <-c:
				input := firehose.PutRecordInput{
					DeliveryStreamName: &f.streamName,
					Record:             &types.Record{Data: d},
				}

				_, err := f.client.PutRecord(context.Background(), &input)
				if err != nil {
					errChan <- fmt.Errorf("unable to write to firehose stream  %s: %w", f.streamName, err)
					return
				}
			case <-f.shChan:
				slog.Info("shutting down firehose exporter")
				return
			}
		}
	}()
}

func (f *FirehoseExporter) Stop() {
	close(f.shChan)
}
