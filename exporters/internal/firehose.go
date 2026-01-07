package internal

import (
	"context"
	"fmt"
	"os"

	"data-gen/conf"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/aws/aws-sdk-go-v2/service/firehose/types"
)

// FirehoseExporter sends generated data to AWS Kinesis Data Firehose.
type FirehoseExporter struct {
	cfg    firehoseCfg
	client *firehose.Client
}

// firehoseCfg specifies the Firehose delivery stream name.
type firehoseCfg struct {
	StreamName string `yaml:"stream_name"`
}

func NewFirehoseExporter(ctx context.Context, c *conf.Config) (*FirehoseExporter, error) {
	var cfg firehoseCfg
	err := c.Output.Conf.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	// load env variable overrides if any
	if v := os.Getenv(conf.EnvOutStreamName); v != "" {
		cfg.StreamName = v
	}

	loadedAwsConfig, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(c.Profile), config.WithRegion(c.Region))
	if err != nil {
		return nil, fmt.Errorf("failed to load default aws config: %w", err)
	}

	fhClient := firehose.New(firehose.Options{
		Credentials: loadedAwsConfig.Credentials,
		Region:      c.Region,
	})

	return &FirehoseExporter{
		cfg:    cfg,
		client: fhClient,
	}, nil
}

func (f *FirehoseExporter) Send(data *[]byte) error {
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
