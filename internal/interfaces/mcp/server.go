package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/vpomo/industrial-mcp/internal/application/command"
	"github.com/vpomo/industrial-mcp/internal/application/query"
	"github.com/vpomo/industrial-mcp/internal/infrastructure/repository"
	"github.com/vpomo/industrial-mcp/pkg/license"
	"github.com/vpomo/industrial-mcp/pkg/logger"
	"github.com/vpomo/industrial-mcp/pkg/x402"
)

type MCPServer struct {
	httpServer *http.Server
	cfg        *Config
	readTagH   *query.ReadTagHandler
	writeTagH  *command.WriteTagHandler
	subTagH    *command.SubscribeTagHandler
	logger     *logger.Logger
	license    *license.Validator
	x402       *x402.Handler
	metrics    *repository.MemoryMetricsRepository
	mu         sync.RWMutex
}

type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      interface{}     `json:"id"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   *MCPError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *MCPError) Error() string {
	return e.Message
}

func NewMCPServer(
	cfg *Config,
	tagService *TagServiceWrapper,
	readTagH *query.ReadTagHandler,
	writeTagH *command.WriteTagHandler,
	subTagH *command.SubscribeTagHandler,
	metricsRepo *repository.MemoryMetricsRepository,
) *MCPServer {
	return &MCPServer{
		cfg:       cfg,
		readTagH:  readTagH,
		writeTagH: writeTagH,
		subTagH:   subTagH,
		logger:    logger.New(cfg.LogLevel),
		metrics:   metricsRepo,
	}
}

func (s *MCPServer) HandleRequest(ctx context.Context, req MCPRequest) (*MCPResponse, error) {
	start := time.Now()
	reqID := uuid.New().String()

	s.logger.Info("request received", "request_id", reqID, "method", req.Method)

	if s.license != nil && s.license.IsEnabled() {
		if err := s.license.Validate(); err != nil {
			s.logger.Error("license validation failed", "request_id", reqID, "error", err)
			return &MCPResponse{
				JSONRPC: "2.0",
				Error:   &MCPError{Code: -32001, Message: "license invalid"},
				ID:      req.ID,
			}, nil
		}
	}

	if s.x402 != nil && s.x402.IsEnabled() {
		if err := s.x402.ValidatePayment(ctx, req.Params); err != nil {
			return &MCPResponse{
				JSONRPC: "2.0",
				Error:   &MCPError{Code: -32002, Message: err.Error()},
				ID:      req.ID,
			}, nil
		}
	}

	var resp interface{}
	var err error

	switch req.Method {
	case "read_tag":
		var q query.ReadTagQuery
		if uerr := json.Unmarshal(req.Params, &q); uerr != nil {
			err = uerr
		} else {
			resp, err = s.readTagH.Handle(ctx, q)
		}
	case "write_tag":
		var c command.WriteTagCommand
		if uerr := json.Unmarshal(req.Params, &c); uerr != nil {
			err = uerr
		} else {
			resp, err = s.writeTagH.Handle(ctx, c)
		}
	case "subscribe_tag":
		var c command.SubscribeTagCommand
		if uerr := json.Unmarshal(req.Params, &c); uerr != nil {
			err = uerr
		} else {
			resp, err = s.subTagH.Handle(ctx, c)
		}
	default:
		err = ErrMethodNotFound
	}

	duration := time.Since(start).Milliseconds()

	if s.metrics != nil {
		s.metrics.SaveRequest(ctx, repository.RequestMetrics{
			ID:        reqID,
			Method:    req.Method,
			Params:    string(req.Params),
			Duration:  duration,
			Timestamp: time.Now().Unix(),
		})
	}

	if err != nil {
		s.logger.Error("request failed", "request_id", reqID, "error", err, "duration_ms", duration)

		code := -32000
		if errors.Is(err, ErrMethodNotFound) {
			code = -32601
		}

		return &MCPResponse{
			JSONRPC: "2.0",
			Error:   &MCPError{Code: code, Message: err.Error()},
			ID:      req.ID,
		}, nil
	}

	s.logger.Info("request completed", "request_id", reqID, "duration_ms", duration)

	return &MCPResponse{
		JSONRPC: "2.0",
		Result:  resp,
		ID:      req.ID,
	}, nil
}

func (s *MCPServer) Start(ctx context.Context) error {
	s.httpServer = &http.Server{
		Addr:    s.cfg.ListenAddr,
		Handler: s,
	}
	return s.httpServer.ListenAndServe()
}

func (s *MCPServer) SetLicenseValidator(lv *license.Validator) {
	s.license = lv
}

func (s *MCPServer) SetX402Handler(xh *x402.Handler) {
	s.x402 = xh
}

var ErrMethodNotFound = errors.New("method not found")
