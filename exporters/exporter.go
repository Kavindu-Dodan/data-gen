package exporters

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"data-gen/conf"
	"data-gen/exporters/internal"
)

const (
	file          = "FILE"
	s3            = "S3"
	firehose      = "FIREHOSE"
	cloudwatchLog = "CLOUDWATCH_LOG"
	eventhub      = "EVENTHUB"
	debug         = "DEBUG"

	defaultShutdownWait = 2 * time.Second
)

type output interface {
	Send(*[]byte) error
}

func ExporterFor(ctx context.Context, cfg *conf.Config) (*Exporter, error) {
	var exporter output
	var err error

	switch cfg.Output.Type {
	case file:
		exporter, err = internal.NewFileExporter(cfg)
		if err != nil {
			return nil, err
		}
	case s3:
		exporter, err = internal.NewS3BucketExporter(ctx, cfg)
		if err != nil {
			return nil, err
		}
	case firehose:
		exporter, err = internal.NewFirehoseExporter(ctx, cfg)
		if err != nil {
			return nil, err
		}

	case cloudwatchLog:
		exporter, err = internal.NewCloudWatchLogExporter(ctx, cfg)
		if err != nil {
			return nil, err
		}
	case eventhub:
		exporter, err = internal.NewEventHubExporter(ctx, cfg)
		if err != nil {
			return nil, err
		}
	case debug:
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

	sending sync.Mutex
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
				e.sending.Lock()
				err := e.output.Send(d)
				if err != nil {
					e.errChan <- err
				}
				e.sending.Unlock()
			case <-e.shChan:
				slog.Info("Shutting down exporter")
				return
			}
		}
	}()

	return e.errChan
}

func (e *Exporter) Stop() {
	close(e.errChan)
	close(e.shChan)

	if e.cfg.Output.WaitForCompletion {
		slog.Info("Waiting for final exports to complete")
		e.sending.Lock()
		e.sending.Unlock()
		return
	}

	slog.Info("Shutting down exporter")
	time.Sleep(defaultShutdownWait)
}
