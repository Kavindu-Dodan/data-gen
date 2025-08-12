package generators

import (
	"encoding/json"
	"math/rand"
	"time"
)

type MetricGenerator struct {
	buf    trackedBuffer
	shChan chan struct{}
}

func NewMetricGenerator() *MetricGenerator {
	return &MetricGenerator{
		buf:    newTrackedBuffer(),
		shChan: make(chan struct{}),
	}
}

func (m *MetricGenerator) Generate() (int64, error) {
	entry, err := m.makeNewMetricsEntry()
	if err != nil {
		return 0, err
	}
	err = m.buf.write(entry)
	return m.buf.size(), err
}

func (m *MetricGenerator) GetAndReset() []byte {
	return m.buf.getAndReset()
}

func (m *MetricGenerator) makeNewMetricsEntry() ([]byte, error) {
	t := time.Now().UnixMilli()

	gen := metricStruct{
		MetricStreamName: "AWSMetrics",
		AccountId:        "111111111111",
		Region:           "us-east-1",
		Namespace:        "AWS/EC2",
		MetricName:       "DiskWriteOps",
		Dimensions: dimensions{
			InstanceId: "i-12345678901",
		},
		Timestamp: t,
		Value: value{
			Count: rand.Intn(100) + 1,
			Sum:   rand.Intn(100) + 1,
			Max:   rand.Intn(100) + 1,
			Min:   rand.Intn(100) + 1,
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
