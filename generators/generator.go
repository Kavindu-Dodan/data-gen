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
	config conf.InputConfig
	input  input
	shChan chan struct{}
}

func newGenerator(cfg conf.InputConfig, in input) *Generator {
	return &Generator{
		config: cfg,
		input:  in,
		shChan: make(chan struct{}),
	}
}

func (g *Generator) Start(delay time.Duration) (<-chan []byte, <-chan error) {
	dChan := make(chan []byte)
	errChan := make(chan error)

	go func() {
		duration, err := time.ParseDuration(g.config.Batching)
		if err != nil {
			errChan <- fmt.Errorf("failed to parse batching duration: %s", err)
			return
		}

		var buf bytes.Buffer
		lastBatch := time.Now()

		for {
			select {
			case <-time.After(delay):
				got, err := g.input.Get()
				if err != nil {
					errChan <- err
				}

				// check for batching
				if time.Since(lastBatch) > duration {
					lastBatch = time.Now()
					dChan <- buf.Bytes()
					buf.Reset()
					g.input.ResetBatch()
				} else {
					buf.Write(got)
				}
			case <-g.shChan:
				slog.Info("Shutting down Generator")
				return
			}
		}
	}()

	return dChan, errChan
}

func (g *Generator) Stop() {
	close(g.shChan)
}
