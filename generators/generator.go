package generators

import (
	"data-gen/generators/internal"
	"fmt"
	"log/slog"
	"time"

	"data-gen/conf"
)

const (
	logs              = "LOGS"
	metrics           = "METRICS"
	alb               = "ALB"
	nlb               = "NLB"
	vpc               = "VPC"
	waf               = "WAF"
	cloudTrail        = "CLOUDTRAIL"
	azureResourceLogs = "AZURE_RESOURCE_LOGS"
)

type input interface {
	// Generate and accumulate data and must return the accumulated size of the total generated data
	Generate() (int64, error)
	// GetAndReset returns the accumulated data and resets the buffer
	GetAndReset() []byte
}

func GeneratorFor(config *conf.Config) (*Generator, error) {
	switch config.Input.Type {
	case logs:
		return newGenerator(config.Input, internal.NewLogGenerator())
	case metrics:
		return newGenerator(config.Input, internal.NewMetricGenerator())
	case alb:
		return newGenerator(config.Input, internal.NewALBGen())
	case nlb:
		return newGenerator(config.Input, internal.NewNLBGen())
	case vpc:
		return newGenerator(config.Input, internal.NewVPCGen())
	case waf:
		return newGenerator(config.Input, internal.NewWAFGen())
	case cloudTrail:
		return newGenerator(config.Input, internal.NewCloudTrailGen())
	case azureResourceLogs:
		return newGenerator(config.Input, internal.NewAzureResourceLogGen(config.Input))
	}

	return nil, fmt.Errorf("unknown generator type: %s", config.Input.Type)
}

type parsedDurations struct {
	delay            time.Duration
	batchingDuration time.Duration
	maxDuration      time.Duration
}

// Generator orchestrates data generation with batching, timing, and lifecycle management.
type Generator struct {
	config          conf.InputConfig
	parsedDurations parsedDurations
	input           input
	dataChan        chan *[]byte
	errChan         chan error
	inputComplete   chan struct{}
	shChan          chan struct{}
}

func newGenerator(cfg conf.InputConfig, in input) (*Generator, error) {
	// validate configs
	delay, err := time.ParseDuration(cfg.Delay)
	if err != nil {
		return nil, fmt.Errorf("error parsing delay: %s, please provide value in acceptable string format like `5s`", err.Error())
	}

	batchingDuration, err := time.ParseDuration(cfg.Batching)
	if err != nil {
		return nil, fmt.Errorf("failed to parse batching duration: %s", err)
	}

	maxDuration, err := time.ParseDuration(cfg.MaxRunTime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse max runtime: %s", err)
	}

	// Avoid spamming loop if all timing and terminal conditions are zero
	if batchingDuration == 0 && maxDuration == 0 && cfg.MaxBatchSize == 0 && cfg.MaxBatchElements == 0 && cfg.MaxDataPoints == 0 {
		return nil, fmt.Errorf("invalid configuration: no batching or terminal conditions specified," +
			" please configure at least one of batching duration, max batch size, max batch elements, max data points, or max runtime")
	}

	return &Generator{
		config: cfg,
		parsedDurations: parsedDurations{
			delay:            delay,
			batchingDuration: batchingDuration,
			maxDuration:      maxDuration,
		},
		input:         in,
		dataChan:      make(chan *[]byte, 2),
		errChan:       make(chan error, 2),
		inputComplete: make(chan struct{}),
		shChan:        make(chan struct{}),
	}, nil
}

func (g *Generator) Start() (data <-chan *[]byte, inputClose <-chan struct{}, error <-chan error) {
	go g.runGenerator()

	return g.dataChan, g.inputComplete, g.errChan
}

func (g *Generator) Stop() {
	// generate stops immediately
	close(g.shChan)
	close(g.errChan)
}

// runGenerator manages the data generation loop, handling timing, batching, and shutdown conditions.
// Contains blocking calls hence should be run in a separate goroutine.
func (g *Generator) runGenerator() {
	totalDataPoints := int64(0)
	currentBatchDataPoints := int64(0)
	start := time.Now()
	lastBatch := time.Now()

	for {
		select {
		case <-time.After(g.parsedDurations.delay):
			// update with latest data
			currentBatchByteSize, err := g.input.Generate()
			if err != nil {
				g.errChan <- err
				return
			}
			totalDataPoints++
			currentBatchDataPoints++

			since := time.Since(lastBatch)

			if g.shouldEmit(since, currentBatchByteSize, currentBatchDataPoints, totalDataPoints) {
				b := g.input.GetAndReset()
				g.dataChan <- &b
				slog.Debug("Emitted payload", slog.Int64("dataPoints", currentBatchDataPoints))

				// if batching duration is not elapsed, pause
				if since < g.parsedDurations.batchingDuration {
					select {
					case <-time.After(g.parsedDurations.batchingDuration - since):
					case <-g.shChan:
						slog.Info("Shutting down Generator")
						return
					}
				}

				// update last batch time
				lastBatch = time.Now()
				currentBatchDataPoints = 0
			}
		case <-g.shChan:
			slog.Info("Shutting down Generator")
			return
		}

		if g.isDone(start, totalDataPoints) {
			// notify input close and exit
			close(g.inputComplete)
			return
		}
	}
}

// shouldEmit checks if conditions are met to emit generated data.
// Checks are done for,
// - batching duration (iff defined)
// - batch size (iff defined)
// - batch count (iff defined)
// - max data points (iff defined)
func (g *Generator) shouldEmit(sinceLastBatch time.Duration, batchSizeBytes int64, batchCount int64, dataPointSum int64) bool {
	if g.parsedDurations.batchingDuration > 0 && sinceLastBatch >= g.parsedDurations.batchingDuration {
		return true
	}

	if g.config.MaxBatchSize > 0 && batchSizeBytes > g.config.MaxBatchSize {
		return true
	}

	if g.config.MaxBatchElements > 0 && batchCount >= g.config.MaxBatchElements {
		return true
	}

	if g.config.MaxDataPoints > 0 && dataPointSum >= g.config.MaxDataPoints {
		return true
	}

	return false
}

// isDone checks if the generator should stop generating.
// Checks are done for,
// - max data points (iff defined)
// - max runtime (iff defined)
func (g *Generator) isDone(startTimestamp time.Time, dataPointSum int64) bool {
	// check for data point limit
	if g.config.MaxDataPoints > 0 && dataPointSum >= g.config.MaxDataPoints {
		slog.Info(fmt.Sprintf("Generator shutting down after %d data points", g.config.MaxDataPoints))
		return true
	}

	// check for max runtime
	if g.parsedDurations.maxDuration > 0 && time.Since(startTimestamp) >= g.parsedDurations.maxDuration {
		slog.Info(fmt.Sprintf("Generator shutting down after max runtime of %s", g.parsedDurations.maxDuration))
		return true
	}

	return false
}
