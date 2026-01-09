package runtime

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMetricsImpl_RecordStart(t *testing.T) {
	m := &MetricsImpl{}
	ts := time.Now()
	m.RecordStart(ts)

	if m.StartTime != ts {
		t.Errorf("expected StartTime %v, got %v", ts, m.StartTime)
	}
}

func TestMetricsImpl_RecordEnd(t *testing.T) {
	m := &MetricsImpl{}
	ts := time.Now()
	m.RecordEnd(ts)

	if m.EndTime != ts {
		t.Errorf("expected EndTime %v, got %v", ts, m.EndTime)
	}
}

func TestMetricsImpl_BatchEmitCount(t *testing.T) {
	m := &MetricsImpl{}
	m.BatchEmitCount(5)
	m.BatchEmitCount(3)

	if m.BatchCount != 8 {
		t.Errorf("expected BatchCount 8, got %d", m.BatchCount)
	}
}

func TestMetricsImpl_ElementsSentCount(t *testing.T) {
	m := &MetricsImpl{}
	m.ElementsSentCount(10)
	m.ElementsSentCount(20)

	if m.ElementCount != 30 {
		t.Errorf("expected ElementCount 30, got %d", m.ElementCount)
	}
}

func TestMetricsImpl_BytesSentCount(t *testing.T) {
	m := &MetricsImpl{}
	m.BytesSentCount(100)
	m.BytesSentCount(50)

	if m.BytesCount != 150 {
		t.Errorf("expected BytesCount 150, got %d", m.BytesCount)
	}
}

func TestMetricsImpl_ToJSON(t *testing.T) {
	tStartStr := "2026-01-01T00:00:00Z"
	tStart, err := time.Parse(time.RFC3339, tStartStr)
	if err != nil {
		t.Fatalf("failed to parse time: %v", err)
	}

	tEndStr := "2026-01-01T01:00:00Z"
	tEnd, err := time.Parse(time.RFC3339, tEndStr)
	if err != nil {
		t.Fatalf("failed to parse time: %v", err)
	}

	m := &MetricsImpl{
		StartTime:    tStart,
		EndTime:      tEnd,
		BatchCount:   5,
		ElementCount: 100,
		BytesCount:   500,
	}

	data, err := m.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON returned error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if result["totalBatches"].(float64) != 5 {
		t.Errorf("expected batchCount 5, got %v", result["totalBatches"])
	}
	if result["totalElements"].(float64) != 100 {
		t.Errorf("expected elementCount 100, got %v", result["totalElements"])
	}
	if result["totalBytes"].(float64) != 500 {
		t.Errorf("expected bytesCount 500, got %v", result["totalBytes"])
	}
	if result["startTime"].(string) != tStartStr {
		t.Errorf("expected startTime %s, got %v", tStartStr, result["startTime"])
	}
	if result["endTime"].(string) != tEndStr {
		t.Errorf("expected endTime %s, got %v", tEndStr, result["endTime"])
	}
}
