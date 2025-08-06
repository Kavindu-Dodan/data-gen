package generators

import (
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
}

func GeneratorFor(config *conf.Config) (*Generator, error) {
	switch config.Input.Type {
	case logs:
		return newGenerator(NewLogGenerator()), nil
	case metrics:
		return newGenerator(NewMetricGenerator()), nil
	case alb:
		return newGenerator(&ALBGen{}), nil
	case vpc:
		return newGenerator(&VPCGen{}), nil
	}

	return nil, fmt.Errorf("unknown generator type: %s", config.Input.Type)
}

type Generator struct {
	in     input
	shChan chan struct{}
}

func newGenerator(in input) *Generator {
	return &Generator{
		in:     in,
		shChan: make(chan struct{}),
	}
}

func (g *Generator) Start(delay time.Duration) (<-chan []byte, <-chan error) {
	dChan := make(chan []byte)
	errChan := make(chan error)

	go func() {
		for {
			select {
			case <-time.After(delay):
				get, err := g.in.Get()
				if err != nil {
					errChan <- err
				}
				dChan <- get
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
