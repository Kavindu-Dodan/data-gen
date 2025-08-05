package exporters

import (
	"context"
	"fmt"
	"log/slog"

	"data-gen/conf"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/aws/aws-sdk-go-v2/service/firehose/types"
)

type FirehoseExporter struct {
	cfg    firehoseCfg
	client *firehose.Client
	shChan chan struct{}
}

type firehoseCfg struct {
	StreamName string `yaml:"stream_name"`
}

func newFirehoseExporter(ctx context.Context, conf *conf.Config) (*FirehoseExporter, error) {
	var cfg firehoseCfg
	err := conf.Output.Conf.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	loadedAwsConfig, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(conf.AWSCfg.Profile), config.WithRegion(conf.AWSCfg.Region))
	if err != nil {
		return nil, fmt.Errorf("failed to load default aws config: %w", err)
	}

	fhClient := firehose.New(firehose.Options{
		Credentials: loadedAwsConfig.Credentials,
		Region:      conf.AWSCfg.Region,
	})

	return &FirehoseExporter{
		cfg:    cfg,
		client: fhClient,
		shChan: make(chan struct{}),
	}, nil
}

func (f *FirehoseExporter) Start(c <-chan []byte) <-chan error {
	errChan := make(chan error)

	go func() {
		for {
			select {
			case d := <-c:
				input := firehose.PutRecordInput{
					DeliveryStreamName: &f.cfg.StreamName,
					Record:             &types.Record{Data: d},
				}

				_, err := f.client.PutRecord(context.Background(), &input)
				if err != nil {
					errChan <- fmt.Errorf("unable to write to firehose stream  %s: %w", f.cfg.StreamName, err)
					return
				}
			case <-f.shChan:
				slog.Info("shutting down firehose exporter")
				return
			}
		}
	}()

	return errChan
}

func (f *FirehoseExporter) Stop() {
	close(f.shChan)
}
