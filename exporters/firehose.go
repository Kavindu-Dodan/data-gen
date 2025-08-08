package exporters

import (
	"context"
	"fmt"
	"os"

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

func newFirehoseExporter(ctx context.Context, c *conf.Config) (*FirehoseExporter, error) {
	var cfg firehoseCfg
	err := c.Output.Conf.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	// load env variable overrides if any
	if v := os.Getenv(conf.EnvOutStreamName); v != "" {
		cfg.StreamName = v
	}

	loadedAwsConfig, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(c.AWSCfg.Profile), config.WithRegion(c.AWSCfg.Region))
	if err != nil {
		return nil, fmt.Errorf("failed to load default aws config: %w", err)
	}

	fhClient := firehose.New(firehose.Options{
		Credentials: loadedAwsConfig.Credentials,
		Region:      c.AWSCfg.Region,
	})

	return &FirehoseExporter{
		cfg:    cfg,
		client: fhClient,
		shChan: make(chan struct{}),
	}, nil
}

func (f *FirehoseExporter) send(data *[]byte) error {
	input := firehose.PutRecordInput{
		DeliveryStreamName: &f.cfg.StreamName,
		Record:             &types.Record{Data: *data},
	}

	_, err := f.client.PutRecord(context.Background(), &input)
	if err != nil {
		return fmt.Errorf("unable to write to firehose stream  %s: %w", f.cfg.StreamName, err)
	}

	return nil
}

func (f *FirehoseExporter) stop() {
	close(f.shChan)
}
