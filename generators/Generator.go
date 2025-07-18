package generators

import (
	"data-gen/conf"
	"fmt"
	"time"
)

const (
	logs    = "LOGS"
	metrics = "METRICS"
)

type Generator interface {
	Start(delay time.Duration, errChan chan<- error) <-chan []byte
	Stop()
}

func GeneratorFor(config *conf.Config) (Generator, error) {
	switch config.Input.Type {
	case logs:
		return newLogGenerator(), nil
	case metrics:
		return newMetricGenerator(), nil
	}

	return nil, fmt.Errorf("unknown generator type: %s", config.Input.Type)
}
