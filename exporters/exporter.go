package exporters

import (
	"context"
	"data-gen/conf"
	"fmt"
	"log/slog"
)

type output interface {
	send(*[]byte) error
}

func ExporterFor(ctx context.Context, config *conf.Config) (*Exporter, error) {
	switch config.Output.Type {
	case "FILE":
		exporter, err := newFileExporter(config)
		if err != nil {
			return nil, err
		}
		return newExporter(exporter), nil
	case "S3":
		exporter, err := newS3BucketExporter(ctx, config)
		if err != nil {
			return nil, err
		}
		return newExporter(exporter), nil
	case "FIREHOSE":
		exporter, err := newFirehoseExporter(ctx, config)
		if err != nil {
			return nil, err
		}

		return newExporter(exporter), nil
	case "CLOUDWATCH_LOG":
		exporter, err := newCloudWatchLogExporter(ctx, config)
		if err != nil {
			return nil, err
		}
		return newExporter(exporter), nil
	}

	return nil, fmt.Errorf("unknown output type: %s", config.Output.Type)
}

type Exporter struct {
	output     output
	inputClose <-chan interface{}
	shChan     chan struct{}
}

func newExporter(output output) *Exporter {
	return &Exporter{
		output: output,
		shChan: make(chan struct{})}
}

func (e *Exporter) Start(data <-chan *[]byte) <-chan error {
	errChan := make(chan error)
	go func() {
		for {
			select {
			case d := <-data:
				err := e.output.send(d)
				if err != nil {
					errChan <- err
				}
			case <-e.shChan:
				slog.Info("Shutting down exporter")
				return
			}
		}
	}()
	return errChan
}

func (e *Exporter) Stop() {
	close(e.shChan)
}
