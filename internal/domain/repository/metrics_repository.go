package repository

import (
	"context"
	"time"
)

type RequestMetrics struct {
	ID         string    `json:"id"`
	Method     string    `json:"method"`
	Params     string    `json:"params"`
	Duration   int64     `json:"duration_ms"`
	Timestamp  time.Time `json:"timestamp"`
	StatusCode int       `json:"status_code"`
	ClientAddr string    `json:"client_addr"`
}

type MetricsRepository interface {
	SaveRequest(ctx context.Context, metrics RequestMetrics) error
	ListRequests(ctx context.Context, limit int) ([]RequestMetrics, error)
}