package generators

import (
	"bytes"
	"fmt"
	"log/slog"
	"time"

	"data-gen/conf"
)

const (
	logs    = "LOGS"
	metrics = "METRICS"
	alb     = "ALB"
	vpc     = "VPC"
)

type input interface {
	Get() ([]byte, error)
	ResetBatch()
}

func GeneratorFor(config *conf.Config) (*Generator, error) {
	switch config.Input.Type {
	case logs:
		return newGenerator(config.Input, NewLogGenerator()), nil
	case metrics:
		return newGenerator(config.Input, NewMetricGenerator()), nil
	case alb:
		return newGenerator(config.Input, &ALBGen{}), nil
	case vpc:
		return newGenerator(config.Input, newVPCGen()), nil
	}

	return nil, fmt.Errorf("unknown generator type: %s", config.Input.Type)
}

type Generator struct {
	config     conf.InputConfig
	input      input
	buffer     bytes.Buffer
	dataChan   chan *[]byte
	errChan    chan error
	inputClose chan struct{}
	shChan     chan struct{}

	dataPoints int64
}

func newGenerator(cfg conf.InputConfig, in input) *Generator {
	return &Generator{
		config:     cfg,
		input:      in,
		dataChan:   make(chan *[]byte, 2),
		errChan:    make(chan error, 2),
		inputClose: make(chan struct{}),
		shChan:     make(chan struct{}),
	}
}

func (g *Generator) Start(delay time.Duration) (data <-chan *[]byte, inputClose <-chan struct{}, error <-chan error) {
	go func() {
		batchingDuration, err := time.ParseDuration(g.config.Batching)
		if err != nil {
			g.errChan <- fmt.Errorf("failed to parse batching duration: %s", err)
			return
		}

		// validate for duration & batching to avoid spamming
		if delay == 0 && batchingDuration == 0 && g.config.MaxSize == 0 {
			g.errChan <- fmt.Errorf("batching & max size must be set when data delay is set to zero")
			return
		}

		buf := newTrackedBuffer()
		lastBatch := time.Now()

		for {
			select {
			case <-time.After(delay):
				// update with latest data
				got, err := g.input.Get()
				if err != nil {
					g.errChan <- err
					return
				}
				g.dataPoints += 1

				_, err = buf.write(got)
				if err != nil {
					g.errChan <- err
				}

				// check for batching,  max size to emit or max data points
				since := time.Since(lastBatch)
				if since > batchingDuration || (g.config.MaxSize != 0 && buf.size() > int64(g.config.MaxSize)) || (g.config.MaxDataPoints > 0 && g.dataPoints >= g.config.MaxDataPoints) {
					b := buf.getAndRest()
					g.dataChan <- &b
					g.input.ResetBatch()

					// if batching duration is not elapsed, pause
					if since < batchingDuration {
						select {
						case <-time.After(batchingDuration - since):
						case <-g.shChan:
							slog.Info("Shutting down Generator")
							return
						}
					}

					// update last batch time
					lastBatch = time.Now()
				}
			case <-g.shChan:
				slog.Info("Shutting down Generator")
				return
			}

			// check for data point limit
			if g.config.MaxDataPoints > 0 && g.dataPoints >= g.config.MaxDataPoints {
				// notify input close and exit
				close(g.inputClose)
				slog.Info(fmt.Sprintf("Generator shutting down after %d points", g.config.MaxDataPoints))
				return
			}
		}
	}()

	return g.dataChan, g.inputClose, g.errChan
}

func (g *Generator) Stop() {
	close(g.shChan)
}

type trackedBuffer struct {
	buf bytes.Buffer
	len int64
}

func newTrackedBuffer() trackedBuffer {
	return trackedBuffer{
		buf: bytes.Buffer{},
	}
}

func (t *trackedBuffer) write(bytes []byte) (int, error) {
	t.len += int64(len(bytes))
	return t.buf.Write(bytes)
}

func (t *trackedBuffer) getAndRest() []byte {
	b := make([]byte, t.buf.Len())
	copy(b, t.buf.Bytes())
	t.buf.Reset()
	t.len = 0
	return b
}

func (t *trackedBuffer) size() int64 {
	return t.len
}
