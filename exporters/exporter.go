package exporters

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"data-gen/conf"
	"data-gen/exporters/internal"
	"data-gen/internal/runtime"
)

const (
	defaultShutdownWait = 2 * time.Second
)

type output interface {
	Send(*[]byte) error
}

func ExporterFor(ctx context.Context, cfg *conf.Config, runtime runtime.Runtime) (*Exporter, error) {
	var exporter output
	var err error

	switch cfg.Output.Type {
	case conf.OutputFile:
		exporter, err = internal.NewFileExporter(cfg)
		if err != nil {
			return nil, err
		}
	case conf.OutputS3:
		exporter, err = internal.NewS3BucketExporter(ctx, cfg)
		if err != nil {
			return nil, err
		}
	case conf.OutputFirehose:
		exporter, err = internal.NewFirehoseExporter(ctx, cfg)
		if err != nil {
			return nil, err
		}

	case conf.OutputCWLogs:
		exporter, err = internal.NewCloudWatchLogExporter(ctx, cfg)
		if err != nil {
			return nil, err
		}
	case conf.OutputEventHub:
		exporter, err = internal.NewEventHubExporter(ctx, cfg)
		if err != nil {
			return nil, err
		}
	case conf.OutputDebug:
		exporter, err = internal.NewDebugExporter(cfg)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown output type: %s", cfg.Output.Type)
	}

	return newExporter(cfg, runtime, exporter), nil
}

// Exporter manages the lifecycle of sending generated data to configured outputs.
type Exporter struct {
	cfg     *conf.Config
	runtime runtime.Runtime
	output  output
	errChan chan error
	shChan  chan struct{}

	sending sync.Mutex
}

func newExporter(cfg *conf.Config, rt runtime.Runtime, output output) *Exporter {
	return &Exporter{
		cfg:     cfg,
		runtime: rt,
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
		// nolint: staticcheck
		e.sending.Unlock()
		return
	}

	slog.Info("Shutting down exporter")
	time.Sleep(defaultShutdownWait)
}
