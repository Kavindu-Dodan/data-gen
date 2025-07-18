package exporters

import (
	"context"
	"data-gen/conf"
	"fmt"
)

type Export interface {
	Start(c <-chan []byte, errChan chan error)
	Stop()
}

func ExporterFor(ctx context.Context, config *conf.Config) (Export, error) {
	switch config.Output.Type {
	case "FILE":
		return newFileExporter(config)
	case "S3":
		return newS3BucketExporter(ctx, config)
	case "FIREHOSE":
		return newFirehoseExporter(ctx, config)
	case "CLOUDWATCH_LOG":
		return newCloudWatchLogExporter(ctx, config)
	}

	return nil, fmt.Errorf("unknown output type: %s", config.Output.Type)
}
