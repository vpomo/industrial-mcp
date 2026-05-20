package mcp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vpomo/mcp_mqtt_opcua/internal/application/command"
	"github.com/vpomo/mcp_mqtt_opcua/internal/application/query"
	"github.com/vpomo/mcp_mqtt_opcua/internal/domain/service"
	infrarepo "github.com/vpomo/mcp_mqtt_opcua/internal/infrastructure/repository"
)

func TestMCPServerReadWriteTag(t *testing.T) {
	tagRepo := infrarepo.NewMemoryTagRepository()
	metricsRepo, _ := infrarepo.NewMemoryMetricsRepository("")
	tagService := service.NewTagService(tagRepo)

	readTagH := query.NewReadTagHandler(tagService)
	writeTagH := command.NewWriteTagHandler(tagRepo, nil)

	cfg := &Config{ListenAddr: "127.0.0.1:0", LogLevel: "debug"}
	server := NewMCPServer(cfg, nil, readTagH, writeTagH, nil, metricsRepo)

	writer := httptest.NewRecorder()
	body := `{"jsonrpc":"2.0","method":"write_tag","params":{"tag_name":"test_sensor","value":42},"id":1}`
	req, _ := http.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.ServeHTTP(writer, req)

	if writer.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", writer.Code)
	}

	var writeResp MCPResponse
	json.Unmarshal(writer.Body.Bytes(), &writeResp)
	if writeResp.Error != nil {
		t.Errorf("write error: %s", writeResp.Error.Message)
	}

	writer = httptest.NewRecorder()
	body = `{"jsonrpc":"2.0","method":"read_tag","params":{"tag_name":"test_sensor"},"id":2}`
	req, _ = http.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.ServeHTTP(writer, req)

	var readResp MCPResponse
	json.Unmarshal(writer.Body.Bytes(), &readResp)
	if readResp.Error != nil {
		t.Errorf("read error: %s", readResp.Error.Message)
	}
}

func TestMCPServerHealthEndpoint(t *testing.T) {
	metricsRepo, _ := infrarepo.NewMemoryMetricsRepository("")
	cfg := &Config{ListenAddr: "127.0.0.1:0", LogLevel: "debug"}
	server := NewMCPServer(cfg, nil, nil, nil, nil, metricsRepo)

	writer := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	server.ServeHTTP(writer, req)

	if writer.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", writer.Code)
	}
	if !strings.Contains(writer.Body.String(), "ok") {
		t.Errorf("expected 'ok' in response, got %s", writer.Body.String())
	}
}

func TestMCPServerInvalidJSON(t *testing.T) {
	metricsRepo, _ := infrarepo.NewMemoryMetricsRepository("")
	cfg := &Config{ListenAddr: "127.0.0.1:0", LogLevel: "debug"}
	server := NewMCPServer(cfg, nil, nil, nil, nil, metricsRepo)

	writer := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	server.ServeHTTP(writer, req)

	if writer.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", writer.Code)
	}
}

func TestMCPServerMethodNotFound(t *testing.T) {
	metricsRepo, _ := infrarepo.NewMemoryMetricsRepository("")
	tagService := service.NewTagService(infrarepo.NewMemoryTagRepository())
	readTagH := query.NewReadTagHandler(tagService)

	cfg := &Config{ListenAddr: "127.0.0.1:0", LogLevel: "debug"}
	server := NewMCPServer(cfg, nil, readTagH, nil, nil, metricsRepo)

	writer := httptest.NewRecorder()
	body := `{"jsonrpc":"2.0","method":"unknown_method","params":{},"id":1}`
	req, _ := http.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.ServeHTTP(writer, req)

	var resp MCPResponse
	json.Unmarshal(writer.Body.Bytes(), &resp)
	if resp.Error == nil {
		t.Fatal("expected error response")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("expected code -32601, got %d", resp.Error.Code)
	}
}

func TestMCPServerWriteAndReadTag(t *testing.T) {
	tagRepo := infrarepo.NewMemoryTagRepository()
	metricsRepo, _ := infrarepo.NewMemoryMetricsRepository("")
	tagService := service.NewTagService(tagRepo)

	readTagH := query.NewReadTagHandler(tagService)
	writeTagH := command.NewWriteTagHandler(tagRepo, nil)

	cfg := &Config{ListenAddr: "127.0.0.1:0", LogLevel: "debug"}
	server := NewMCPServer(cfg, nil, readTagH, writeTagH, nil, metricsRepo)

	writer := httptest.NewRecorder()
	body := `{"jsonrpc":"2.0","method":"write_tag","params":{"tag_name":"test_sensor","value":42},"id":1}`
	req, _ := http.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.ServeHTTP(writer, req)

	if writer.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", writer.Code)
	}

	writer = httptest.NewRecorder()
	body = `{"jsonrpc":"2.0","method":"read_tag","params":{"tag_name":"test_sensor"},"id":2}`
	req, _ = http.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.ServeHTTP(writer, req)

	if writer.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", writer.Code)
	}
}
