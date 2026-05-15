package repository

import (
	"context"
	"encoding/json"
	"os"
	"sync"
)

type RequestMetrics struct {
	ID         string `json:"id"`
	Method     string `json:"method"`
	Params     string `json:"params"`
	Duration   int64  `json:"duration_ms"`
	Timestamp  int64  `json:"timestamp"`
	StatusCode int    `json:"status_code"`
	ClientAddr string `json:"client_addr"`
}

type MetricsRepository interface {
	SaveRequest(ctx context.Context, metrics RequestMetrics) error
	ListRequests(ctx context.Context, limit int) ([]RequestMetrics, error)
}

type MemoryMetricsRepository struct {
	mu      sync.RWMutex
	metrics []RequestMetrics
	file    *os.File
}

func NewMemoryMetricsRepository(filepath string) (*MemoryMetricsRepository, error) {
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &MemoryMetricsRepository{file: f}, nil
}

func (r *MemoryMetricsRepository) SaveRequest(ctx context.Context, metrics RequestMetrics) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metrics = append(r.metrics, metrics)
	data, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	_, err = r.file.Write(append(data, '\n'))
	return err
}

func (r *MemoryMetricsRepository) ListRequests(ctx context.Context, limit int) ([]RequestMetrics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if limit <= 0 || limit > len(r.metrics) {
		limit = len(r.metrics)
	}
	return r.metrics[len(r.metrics)-limit:], nil
}

func (r *MemoryMetricsRepository) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}