package runtime

import (
	"encoding/json"
	"sync"
	"time"
)

// Metrics exposes metrics recording capabilities
type Metrics interface {
	RecordStart(timestamp time.Time)
	RecordEnd(timestamp time.Time)
	BatchEmitCount(count int64)
	ElementsSentCount(count int64)
	BytesSentCount(count int64)
	ToJSON() ([]byte, error)
}

type MetricsImpl struct {
	StartTime    time.Time `json:"startTime"`
	EndTime      time.Time `json:"endTime"`
	BatchCount   int64     `json:"totalBatches"`
	ElementCount int64     `json:"totalElements"`
	BytesCount   int64     `json:"totalBytes"`

	lock sync.Mutex
}

func newMetricsImpl() Metrics {
	return &MetricsImpl{}
}

func (m *MetricsImpl) RecordStart(timestamp time.Time) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.StartTime = timestamp
}

func (m *MetricsImpl) RecordEnd(timestamp time.Time) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.EndTime = timestamp
}

func (m *MetricsImpl) BatchEmitCount(count int64) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.BatchCount += count
}

func (m *MetricsImpl) ElementsSentCount(count int64) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.ElementCount += count
}

func (m *MetricsImpl) BytesSentCount(count int64) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.BytesCount += count
}

func (m *MetricsImpl) ToJSON() ([]byte, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	return json.Marshal(m)
}
