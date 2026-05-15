package logger

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"
)

type RequestMetric struct {
	ID         string    `json:"id"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	Duration   int64     `json:"duration_ms"`
	StatusCode int       `json:"status_code"`
	ClientIP   string    `json:"client_ip"`
	Timestamp  time.Time `json:"timestamp"`
	UserAgent  string    `json:"user_agent"`
	LicenseID  string    `json:"license_id,omitempty"`
	IsPaid     bool      `json:"is_paid"`
}

type MetricsLogger struct {
	mu      sync.Mutex
	file    *os.File
	metrics []RequestMetric
}

func NewMetricsLogger(filepath string) (*MetricsLogger, error) {
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &MetricsLogger{file: f}, nil
}

func (m *MetricsLogger) LogRequest(ctx context.Context, metric RequestMetric) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics = append(m.metrics, metric)
	data, _ := json.Marshal(metric)
	_, err := m.file.Write(append(data, '\n'))
	return err
}

func (m *MetricsLogger) Close() error {
	if m.file != nil {
		return m.file.Close()
	}
	return nil
}

func (m *MetricsLogger) GetMetrics() []RequestMetric {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]RequestMetric, len(m.metrics))
	copy(result, m.metrics)
	return result
}