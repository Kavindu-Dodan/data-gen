package generators

import (
	"data-gen/conf"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"time"
)

type MetricGenerator struct {
	region string
	shChan chan struct{}
}

func NewMetricGenerator(cfg *conf.Cfg) *MetricGenerator {
	return &MetricGenerator{
		region: cfg.Region,
		shChan: make(chan struct{}),
	}
}

func (m *MetricGenerator) Start(delay time.Duration, errChan chan<- error) <-chan []byte {
	c := make(chan []byte, 2)

	go func() {
		for {
			select {
			case <-time.After(delay):
				entry, err := m.makeNewMetricsEntry()
				if err != nil {
					errChan <- fmt.Errorf("error generating metrics: %s", err)
					return
				}
				c <- entry
			case _ = <-m.shChan:
				slog.Info("shutting down log generator")
				return
			}
		}
	}()

	return c
}

func (m *MetricGenerator) Stop() {
	close(m.shChan)
}

func (m *MetricGenerator) makeNewMetricsEntry() ([]byte, error) {
	t := time.Now().Unix()

	gen := metricStruct{
		MetricStreamName: "AWSMetrics",
		AccountId:        "111111111111",
		Region:           m.region,
		Namespace:        "AWS/EC2",
		MetricName:       "DiskWriteOps",
		Dimensions: dimensions{
			InstanceId: "i-12345678901",
		},
		Timestamp: t,
		Value: value{
			Count: rand.Intn(100),
			Sum:   rand.Intn(100),
			Max:   rand.Intn(100),
			Min:   rand.Intn(100),
		},
		Unit: "Seconds",
	}

	marshal, err := json.Marshal(gen)
	if err != nil {
		return nil, err
	}

	return append(marshal, []byte("\n")...), nil
}

type metricStruct struct {
	MetricStreamName string     `json:"metric_stream_name"`
	AccountId        string     `json:"account_id"`
	Region           string     `json:"region"`
	Namespace        string     `json:"namespace"`
	MetricName       string     `json:"metric_name"`
	Dimensions       dimensions `json:"dimensions"`
	Timestamp        int64      `json:"timestamp"`
	Value            value      `json:"value"`
	Unit             string     `json:"unit"`
}

type value struct {
	Count int `json:"count"`
	Sum   int `json:"sum"`
	Max   int `json:"max"`
	Min   int `json:"min"`
}

type dimensions struct {
	InstanceId string `json:"InstanceId"`
}
