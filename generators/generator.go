package generators

import (
	"bytes"
	"fmt"
	"log/slog"
	"time"

	"data-gen/conf"
)

const (
	logs       = "LOGS"
	metrics    = "METRICS"
	alb        = "ALB"
	vpc        = "VPC"
	CloudTrail = "CLOUDTRAIL"
)

type input interface {
	// Generate and accumulate data and return the size of the data generated
	Generate() (int64, error)
	// GetAndReset returns the accumulated data and resets the buffer
	GetAndReset() []byte
}

func GeneratorFor(config *conf.Config) (*Generator, error) {
	switch config.Input.Type {
	case logs:
		return newGenerator(config.Input, NewLogGenerator()), nil
	case metrics:
		return newGenerator(config.Input, NewMetricGenerator()), nil
	case alb:
		return newGenerator(config.Input, NewALBGen()), nil
	case vpc:
		return newGenerator(config.Input, newVPCGen()), nil
	case CloudTrail:
		return newGenerator(config.Input, newCloudTrailGen()), nil
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

		dataPoints := int64(0)
		currentPayload := int64(0)
		lastBatch := time.Now()

		for {
			select {
			case <-time.After(delay):
				// update with latest data
				currentSize, err := g.input.Generate()
				if err != nil {
					g.errChan <- err
					return
				}
				dataPoints += 1
				currentPayload += 1

				since := time.Since(lastBatch)

				// check for following before emitting
				// - batching duration
				// - max size (iff defined)
				// - max data points (iff defined)
				if since > batchingDuration || (g.config.MaxSize != 0 && currentSize > int64(g.config.MaxSize)) || (g.config.MaxDataPoints > 0 && dataPoints >= g.config.MaxDataPoints) {
					b := g.input.GetAndReset()
					g.dataChan <- &b

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
					slog.Debug("Emitted payload", slog.Int64("dataPoints", currentPayload))
					currentPayload = 0
				}
			case <-g.shChan:
				slog.Info("Shutting down Generator")
				return
			}

			// check for data point limit
			if g.config.MaxDataPoints > 0 && dataPoints >= g.config.MaxDataPoints {
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

// trackedBuffer is a helper to track the size of the buffer
type trackedBuffer struct {
	buf bytes.Buffer
	len int64
}

func newTrackedBuffer() trackedBuffer {
	return trackedBuffer{
		buf: bytes.Buffer{},
	}
}

func (t *trackedBuffer) write(bytes []byte) error {
	t.len += int64(len(bytes))
	_, err := t.buf.Write(bytes)
	return err
}

func (t *trackedBuffer) getAndReset() []byte {
	b := make([]byte, t.buf.Len())
	copy(b, t.buf.Bytes())
	t.buf.Reset()
	t.len = 0
	return b
}

func (t *trackedBuffer) size() int64 {
	return t.len
}
