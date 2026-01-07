package exporters

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"data-gen/conf"
	"data-gen/exporters/internal"
)

type output interface {
	Send(*[]byte) error
}

func ExporterFor(ctx context.Context, cfg *conf.Config) (*Exporter, error) {
	var exporter output
	var err error

	switch cfg.Output.Type {
	case "FILE":
		exporter, err = internal.NewFileExporter(cfg)
		if err != nil {
			return nil, err
		}
	case "S3":
		exporter, err = internal.NewS3BucketExporter(ctx, cfg)
		if err != nil {
			return nil, err
		}
	case "FIREHOSE":
		exporter, err = internal.NewFirehoseExporter(ctx, cfg)
		if err != nil {
			return nil, err
		}

	case "CLOUDWATCH_LOG":
		exporter, err = internal.NewCloudWatchLogExporter(ctx, cfg)
		if err != nil {
			return nil, err
		}
	case "EVENTHUB":
		exporter, err = internal.NewEventHubExporter(ctx, cfg)
		if err != nil {
			return nil, err
		}
	case "DEBUG":
		exporter, err = internal.NewDebugExporter(cfg)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown output type: %s", cfg.Output.Type)
	}

	return newExporter(cfg, exporter), nil
}

// Exporter manages the lifecycle of sending generated data to configured outputs.
type Exporter struct {
	cfg     *conf.Config
	output  output
	errChan chan error
	shChan  chan struct{}
}

func newExporter(cfg *conf.Config, output output) *Exporter {
	return &Exporter{
		cfg:     cfg,
		output:  output,
		errChan: make(chan error, 2),
		shChan:  make(chan struct{}),
	}
}

func (e *Exporter) Start(data <-chan *[]byte) <-chan error {
	go func() {
		for {
			select {
			case d := <-data:
				err := e.output.Send(d)
				if err != nil {
					e.errChan <- err
				}
			case <-e.shChan:
				slog.Info("Shutting down exporter")
				return
			}
		}
	}()

	return e.errChan
}

func (e *Exporter) Stop() {
	// close error channel immediately to avoid blocking on error sends
	close(e.errChan)

	// todo: configurable shutdown wait time through config
	<-time.After(time.Second * 2)
	close(e.shChan)
}
