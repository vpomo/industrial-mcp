# План разработки MCP-сервера (MQTT + OPC UA) на Go 1.26

## Содержание
1. [Общая архитектура](#1-общая-архитектура)
2. [Структура проекта](#2-структура-проекта)
3. [Пошаговый план разработки](#3-пошаговый-план-разработки)
   - 3.1 [Инициализация проекта и Docker](#31-инициализация-проекта-и-docker)
   - 3.2 [Domain Layer (Доменный слой)](#32-domain-layer-доменный-слой)
   - 3.3 [Application Layer (Слой приложения)](#33-application-layer-слой-приложения)
   - 3.4 [Infrastructure Layer (Инфраструктурный слой)](#34-infrastructure-layer-инфраструктурный-слой)
   - 3.5 [Interfaces Layer (Слой интерфейсов)](#35-interfaces-layer-слой-интерфейсов)
   - 3.6 [License Module (Модуль лицензирования)](#36-license-module-модуль-лицензирования)
   - 3.7 [Payment x402 Module](#37-payment-x402-module)
   - 3.8 [Logger Module (Модуль логирования)](#38-logger-module-модуль-логирования)
4. [Конфигурация через ENV](#4-конфигурация-через-env)
5. [Сборка и запуск](#5-сборка-и-запуск)

---

## 1. Общая архитектура

```
┌─────────────────────────────────────────────────────────────────┐
│                      LiteDDD Architecture                        │
├─────────────────────────────────────────────────────────────────┤
│  Interfaces Layer (Ports)                                       │
│  ┌─────────────────┐  ┌─────────────────┐                       │
│  │   MCP Server    │  │   x402 Payment  │  ← Входы в систему    │
│  │   (JSON-RPC)    │  │   Handler       │                       │
│  └────────┬────────┘  └────────┬────────┘                       │
│           │                    │                                │
│  Application Layer (Use Cases)                                   │
│  ┌────────┴────────┐  ┌────────┴────────┐                      │
│  │ ReadTagCommand  │  │ WriteTagCommand  │  ← Координация         │
│  │ ReadTagQuery   │  │ SubscribeCommand │                       │
│  └────────┬────────┘  └────────┬────────┘                      │
│           │                    │                                │
│  Domain Layer (Core)                                           │
│  ┌────────┴────────┐  ┌────────┴────────┐                      │
│  │      Tag        │  │  Repository    │  ← Бизнес-логика       │
│  │   Entity        │  │  Interfaces    │                       │
│  └─────────────────┘  └─────────────────┘                      │
│                                                                  │
│  Infrastructure Layer (Adapters)                               │
│  ┌──────────┐ ┌──────────┐ ┌──────────────┐ ┌───────────┐      │
│  │ MQTT    │ │ OPC UA   │ │ MemoryCache  │ │ License   │      │
│  │ Client  │ │ Client   │ │ Repository   │ │ RSA Check │      │
│  └──────────┘ └──────────┘ └──────────────┘ └───────────┘      │
│                                                                  │
│  Shared Modules                                                │
│  ┌───────────┐ ┌─────────────┐ ┌──────────────┐                │
│  │  Logger   │ │  License    │ │   x402      │                │
│  │ (ZeroLog) │ │ (RSA Keys)  │ │   Proto     │                │
│  └───────────┘ └─────────────┘ └──────────────┘                │
└─────────────────────────────────────────────────────────────────┘
```

---

## 2. Структура проекта

```
/home/user/vpomo/industrial-mcp/
├── cmd/
│   └── server/
│       └── main.go                 # Точка входа
├── internal/
│   ├── domain/                     # Доменный слой (LiteDDD)
│   │   ├── entity/
│   │   │   ├── tag.go             # Сущность Tag
│   │   │   └── license.go          # Сущность License
│   │   ├── repository/             # Интерфейсы репозиториев
│   │   │   ├── tag_repository.go
│   │   │   └── metrics_repository.go
│   │   ├── service/                # Доменные сервисы
│   │   │   └── tag_service.go
│   │   └── events/                 # Доменные события
│   │       └── tag_events.go
│   ├── application/                # Слой приложения (Use Cases)
│   │   ├── command/               # Команды (CUD операции)
│   │   │   ├── write_tag.go
│   │   │   └── subscribe_tag.go
│   │   └── query/                  # Запросы (Read операции)
│   │       ├── read_tag.go
│   │       └── list_tags.go
│   ├── infrastructure/             # Инфраструктурный слой
│   │   ├── mqtt/
│   │   │   └── client.go          # MQTT publisher/subscriber
│   │   ├── opcua/
│   │   │   └── client.go         # OPC UA client
│   │   ├── repository/
│   │   │   ├── memory_tag_repo.go # In-memory реализация
│   │   │   └── memory_metrics_repo.go
│   │   └── license/
│   │       └── rsa_validator.go  # RSA-валидация лицензий
│   └── interfaces/                 # Слой интерфейсов (Ports)
│       ├── mcp/
│       │   ├── server.go         # MCP JSON-RPC сервер
│       │   ├── handler.go        # Обработчики tools
│       │   └── middleware.go     # Логирование запросов
│       └── x402/
│           └── handler.go        # x402 payment handler
├── pkg/
│   ├── license/                    # Оффлайн проверка лицензии
│   │   ├── crypto.go              # RSA шифрование/проверка
│   │   ├── hardware.go            # Сбор хардварной информации
│   │   └── validator.go          # Валидация лицензии
│   ├── logger/                    # Логирование
│   │   ├── zap.go                 # ZeroLog обёртка
│   │   └── metrics.go             # Запись метрик запросов
│   └── x402/                      # x402 протокол
│       ├── client.go              # x402 JSON-RPC клиент
│       └── payment.go             # Типы платежей
├── configs/
│   └── config.yaml                # Конфигурация
├── Dockerfile                     # Docker образ
├── docker-compose.yml             # Docker Compose
├── go.mod                          # Go modules
├── go.sum
└── Makefile                       # Makefile для сборки
```

---

## 3. Пошаговый план разработки

### 3.1 Инициализация проекта и Docker

#### 3.1.1 Инициализация Go модуля
```bash
go mod init github.com/vpomo/industrial-mcp
```

#### 3.1.2 Создание go.mod с зависимостями
```go
// go.mod
module github.com/vpomo/industrial-mcp

go 1.26

require (
    github.com/Azure/go-ansiterm v0.0.0-20210617225207-d1cf6083db52
    github.com/docker/docker v27.5.0+incompatible
    github.com/google/uuid v1.6.0
    github.com/rs/zerolog v1.33.0
    github.com/samber/lo v1.47.0
    go.uber.org/multierr v1.11.0
    golang.org/x/crypto v0.27.0
)
```

#### 3.1.3 Dockerfile
```dockerfile
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /server /app/server
COPY configs /app/configs

ENV GIN_MODE=release
EXPOSE 8080

CMD ["/app/server"]
```

#### 3.1.4 docker-compose.yml
```yaml
version: '3.8'
services:
  mcp-server:
    build: .
    container_name: industrial-mcp
    ports:
      - "8080:8080"
    env_file:
      - .env
    volumes:
      - ./license:/app/license:ro
      - ./logs:/app/logs
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

#### 3.1.5 .env.example
```bash
# Сервер
APP_HOST=0.0.0.0
APP_PORT=8080
LOG_LEVEL=info

# MQTT
MQTT_BROKER_URL=tcp://localhost:1883
MQTT_CLIENT_ID=mcp_server
MQTT_USERNAME=
MQTT_PASSWORD=
MQTT_TOPIC_PREFIX=mcp/

# OPC UA
OPCUA_ENDPOINT=opc.tcp://localhost:4840
OPCUA_SECURITY_MODE=None
OPCUA_CERT_FILE=
OPCUA_KEY_FILE=

# Лицензирование
LICENSE_ENABLED=true
LICENSE_PUBLIC_KEY_PATH=/app/license/public.pem
LICENSE_CHECK_INTERVAL=3600

# x402 Payment
X402_ENABLED=true
X402_PAYMENT_ADDRESS=0x...
X402_WALLET_PRIVATE_KEY=

# Статистика
METRICS_ENABLED=true
METRICS_FILE=/app/logs/requests.jsonl
```

---

### 3.2 Domain Layer (Доменный слой)

#### 3.2.1 Сущность Tag (`internal/domain/entity/tag.go`)
```go
package entity

import (
    "time"
    "github.com/google/uuid"
)

type Tag struct {
    id        string
    name      string
    value     interface{}
    timestamp time.Time
    quality   Quality
}

type Quality int

const (
    QualityGood       Quality = 0
    QualityUncertain Quality = 1
    QualityBad       Quality = 2
)

func NewTag(name string, value interface{}) (*Tag, error) {
    if name == "" {
        return nil, errors.New("tag name is required")
    }
    return &Tag{
        id:        uuid.New().String(),
        name:      name,
        value:     value,
        timestamp: time.Now(),
        quality:   QualityGood,
    }, nil
}

func (t *Tag) ID() string         { return t.id }
func (t *Tag) Name() string       { return t.name }
func (t *Tag) Value() interface{} { return t.value }
func (t *Tag) Timestamp() time.Time { return t.timestamp }
func (t *Tag) Quality() Quality   { return t.quality }

func (t *Tag) UpdateValue(newValue interface{}) {
    t.value = newValue
    t.timestamp = time.Now()
}
```

#### 3.2.2 Сущность License (`internal/domain/entity/license.go`)
```go
package entity

type License struct {
    id           string
    hardwareHash string
    expiresAt    time.Time
    features     []string
    isValid      bool
}

func (l *License) IsExpired() bool {
    return time.Now().After(l.expiresAt)
}

func (l *License) HasFeature(feature string) bool {
    for _, f := range l.features {
        if f == feature {
            return true
        }
    }
    return false
}
```

#### 3.2.3 Интерфейс TagRepository (`internal/domain/repository/tag_repository.go`)
```go
package repository

import "github.com/vpomo/industrial-mcp/internal/domain/entity"

type TagReader interface {
    GetByID(ctx context.Context, id string) (*entity.Tag, error)
    GetByName(ctx context.Context, name string) (*entity.Tag, error)
    List(ctx context.Context) ([]*entity.Tag, error)
}

type TagWriter interface {
    Save(ctx context.Context, tag *entity.Tag) error
    Delete(ctx context.Context, id string) error
}

type TagRepository interface {
    TagReader
    TagWriter
}
```

#### 3.2.4 Интерфейс MetricsRepository (`internal/domain/repository/metrics_repository.go`)
```go
package repository

import "context"

type RequestMetrics struct {
    ID          string    `json:"id"`
    Method      string    `json:"method"`
    Params      string    `json:"params"`
    Duration    int64     `json:"duration_ms"`
    Timestamp   time.Time `json:"timestamp"`
    StatusCode int       `json:"status_code"`
    ClientAddr string    `json:"client_addr"`
}

type MetricsRepository interface {
    SaveRequest(ctx context.Context, metrics RequestMetrics) error
    ListRequests(ctx context.Context, limit int) ([]RequestMetrics, error)
}
```

#### 3.2.5 Доменный сервис TagService (`internal/domain/service/tag_service.go`)
```go
package service

import (
    "context"
    "errors"
    "github.com/vpomo/industrial-mcp/internal/domain/entity"
    "github.com/vpomo/industrial-mcp/internal/domain/repository"
)

var (
    ErrTagNotFound = errors.New("tag not found")
)

type TagService struct {
    repo repository.TagRepository
}

func NewTagService(repo repository.TagRepository) *TagService {
    return &TagService{repo: repo}
}

func (s *TagService) GetTag(ctx context.Context, name string) (*entity.Tag, error) {
    tag, err := s.repo.GetByName(ctx, name)
    if err != nil {
        return nil, ErrTagNotFound
    }
    return tag, nil
}

func (s *TagService) UpdateTag(ctx context.Context, name string, value interface{}) error {
    tag, err := s.repo.GetByName(ctx, name)
    if err != nil {
        return ErrTagNotFound
    }
    tag.UpdateValue(value)
    return s.repo.Save(ctx, tag)
}
```

---

### 3.3 Application Layer (Слой приложения)

#### 3.3.1 Query: ReadTagQuery (`internal/application/query/read_tag.go`)
```go
package query

import (
    "context"
    "github.com/vpomo/industrial-mcp/internal/domain/service"
)

type ReadTagQuery struct {
    TagName string `json:"tag_name"`
}

type ReadTagHandler struct {
    tagService *service.TagService
}

func NewReadTagHandler(ts *service.TagService) *ReadTagHandler {
    return &ReadTagHandler{tagService: ts}
}

type ReadTagResponse struct {
    ID        string      `json:"id"`
    Name      string      `json:"name"`
    Value     interface{} `json:"value"`
    Timestamp string      `json:"timestamp"`
    Quality   int         `json:"quality"`
}

func (h *ReadTagHandler) Handle(ctx context.Context, q ReadTagQuery) (*ReadTagResponse, error) {
    tag, err := h.tagService.GetTag(ctx, q.TagName)
    if err != nil {
        return nil, err
    }
    return &ReadTagResponse{
        ID:        tag.ID(),
        Name:      tag.Name(),
        Value:     tag.Value(),
        Timestamp: tag.Timestamp().Format(time.RFC3339),
        Quality:   int(tag.Quality()),
    }, nil
}
```

#### 3.3.2 Command: WriteTagCommand (`internal/application/command/write_tag.go`)
```go
package command

import (
    "context"
    "github.com/vpomo/industrial-mcp/internal/domain/entity"
    "github.com/vpomo/industrial-mcp/internal/domain/repository"
)

type WriteTagCommand struct {
    TagName string      `json:"tag_name"`
    Value   interface{} `json:"value"`
}

type WriteTagHandler struct {
    repo      repository.TagRepository
    publisher PublishPublisher
}

type PublishPublisher interface {
    Publish(ctx context.Context, topic string, payload []byte) error
}

func NewWriteTagHandler(repo repository.TagRepository, pub PublishPublisher) *WriteTagHandler {
    return &WriteTagHandler{repo: repo, publisher: pub}
}

type WriteTagResponse struct {
    ID      string `json:"id"`
    Success bool   `json:"success"`
}

func (h *WriteTagHandler) Handle(ctx context.Context, cmd WriteTagCommand) (*WriteTagResponse, error) {
    tag, err := entity.NewTag(cmd.TagName, cmd.Value)
    if err != nil {
        return nil, err
    }
    if err := h.repo.Save(ctx, tag); err != nil {
        return nil, err
    }
    payload, _ := json.Marshal(tag)
    h.publisher.Publish(ctx, "mcp/tag/written", payload)
    return &WriteTagResponse{ID: tag.ID(), Success: true}, nil
}
```

#### 3.3.3 Command: SubscribeTagCommand (`internal/application/command/subscribe_tag.go`)
```go
package command

type SubscribeTagCommand struct {
    TagName string   `json:"tag_name"`
    QoS     int      `json:"qos"`
}

type SubscribeTagHandler struct {
    subscriber TagSubscriber
}

type TagSubscriber interface {
    Subscribe(ctx context.Context, topic string, handler func([]byte)) error
}

func NewSubscribeTagHandler(sub TagSubscriber) *SubscribeTagHandler {
    return &SubscribeTagHandler{subscriber: sub}
}

type SubscribeTagResponse struct {
    SubscriptionID string `json:"subscription_id"`
    Topic          string `json:"topic"`
}

func (h *SubscribeTagHandler) Handle(ctx context.Context, cmd SubscribeTagCommand) (*SubscribeTagResponse, error) {
    subID := uuid.New().String()
    topic := "mcp/tag/" + cmd.TagName
    err := h.subscriber.Subscribe(ctx, topic, func(data []byte) {
        // обработка данных
    })
    if err != nil {
        return nil, err
    }
    return &SubscribeTagResponse{
        SubscriptionID: subID,
        Topic:          topic,
    }, nil
}
```

---

### 3.4 Infrastructure Layer (Инфраструктурный слой)

#### 3.4.1 MQTT Client (`internal/infrastructure/mqtt/client.go`)
```go
package mqtt

import (
    "context"
    mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient struct {
    client mqtt.Client
    prefix string
}

func NewMQTTClient(brokerURL, clientID, prefix string) (*MQTTClient, error) {
    opts := mqtt.NewClientOptions().
        AddBroker(brokerURL).
        SetClientID(clientID)

    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        return nil, token.Error()
    }
    return &MQTTClient{client: client, prefix: prefix}, nil
}

func (m *MQTTClient) Publish(ctx context.Context, topic string, payload []byte) error {
    fullTopic := m.prefix + topic
    token := m.client.Publish(fullTopic, 0, false, payload)
    token.Wait()
    return token.Error()
}

func (m *MQTTClient) Subscribe(ctx context.Context, topic string, handler mqtt.MessageHandler) error {
    fullTopic := m.prefix + topic
    token := m.client.Subscribe(fullTopic, 0, handler)
    token.Wait()
    return token.Error()
}

func (m *MQTTClient) Disconnect() {
    m.client.Disconnect(250)
}
```

#### 3.4.2 OPC UA Client (`internal/infrastructure/opcua/client.go`)
```go
package opcua

import (
    "context"
    "github.com/vpomo/industrial-mcp/internal/domain/entity"
    "github.com/google/uuid"
    opcua "github.com/vpomo/go-opcua"
)

type OPCUAClient struct {
    client    *opcua.Client
    endpoint  string
    tagCache  map[string]*entity.Tag
}

func NewOPCUAClient(endpoint string) (*OPCUAClient, error) {
    client, err := opcua.Connect(endpoint)
    if err != nil {
        return nil, err
    }
    return &OPCUAClient{
        client:   client,
        endpoint: endpoint,
        tagCache: make(map[string]*entity.Tag),
    }, nil
}

func (c *OPCUAClient) ReadTag(ctx context.Context, nodeID string) (*entity.Tag, error) {
    value, quality, timestamp, err := c.client.Read(nodeID)
    if err != nil {
        return nil, err
    }
    tag, _ := entity.NewTag(nodeID, value)
    tag.quality = qualityToQuality(quality)
    tag.timestamp = timestamp
    c.tagCache[nodeID] = tag
    return tag, nil
}

func (c *OPCUAClient) WriteTag(ctx context.Context, nodeID string, value interface{}) error {
    return c.client.Write(nodeID, value)
}

func (c *OPCUAClient) Disconnect() {
    c.client.Close()
}

func qualityToQuality(q opcua.StatusCode) entity.Quality {
    if q == opcua.StatusGood {
        return entity.QualityGood
    }
    return entity.QualityBad
}
```

#### 3.4.3 In-Memory Tag Repository (`internal/infrastructure/repository/memory_tag_repo.go`)
```go
package repository

import (
    "context"
    "sync"
    "github.com/vpomo/industrial-mcp/internal/domain/entity"
)

type MemoryTagRepository struct {
    mu   sync.RWMutex
    tags map[string]*entity.Tag
}

func NewMemoryTagRepository() *MemoryTagRepository {
    return &MemoryTagRepository{tags: make(map[string]*entity.Tag)}
}

func (r *MemoryTagRepository) GetByID(ctx context.Context, id string) (*entity.Tag, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    tag, ok := r.tags[id]
    if !ok {
        return nil, errors.New("tag not found")
    }
    return tag, nil
}

func (r *MemoryTagRepository) GetByName(ctx context.Context, name string) (*entity.Tag, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    for _, tag := range r.tags {
        if tag.Name() == name {
            return tag, nil
        }
    }
    return nil, errors.New("tag not found")
}

func (r *MemoryTagRepository) List(ctx context.Context) ([]*entity.Tag, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    result := make([]*entity.Tag, 0, len(r.tags))
    for _, tag := range r.tags {
        result = append(result, tag)
    }
    return result, nil
}

func (r *MemoryTagRepository) Save(ctx context.Context, tag *entity.Tag) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.tags[tag.ID()] = tag
    return nil
}

func (r *MemoryTagRepository) Delete(ctx context.Context, id string) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    delete(r.tags, id)
    return nil
}
```

#### 3.4.4 In-Memory Metrics Repository (`internal/infrastructure/repository/memory_metrics_repo.go`)
```go
package repository

import (
    "context"
    "encoding/json"
    "os"
    "sync"
    "time"
)

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
    if limit > len(r.metrics) {
        limit = len(r.metrics)
    }
    return r.metrics[len(r.metrics)-limit:], nil
}

func (r *MemoryMetricsRepository) Close() error {
    return r.file.Close()
}
```

---

### 3.5 Interfaces Layer (Слой интерфейсов)

#### 3.5.1 MCP Server (`internal/interfaces/mcp/server.go`)
```go
package mcp

import (
    "context"
    "encoding/json"
    "net/http"
    "sync"

    "github.com/vpomo/industrial-mcp/internal/application/command"
    "github.com/vpomo/industrial-mcp/internal/application/query"
    "github.com/vpomo/industrial-mcp/internal/infrastructure/repository"
    "github.com/vpomo/industrial-mcp/internal/pkg/license"
    "github.com/vpomo/industrial-mcp/internal/pkg/logger"
    "github.com/vpomo/industrial-mcp/internal/pkg/x402"
)

type MCPServer struct {
    httpServer *http.Server
    readTagH   *query.ReadTagHandler
    writeTagH  *command.WriteTagHandler
    subTagH    *command.SubscribeTagHandler
    logger     *logger.Logger
    license    *license.LicenseValidator
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

func NewMCPServer(cfg Config) *MCPServer {
    tagRepo := repository.NewMemoryTagRepository()
    metricsRepo, _ := repository.NewMemoryMetricsRepository(cfg.MetricsFile)
    tagService := service.NewTagService(tagRepo)

    readTagH := query.NewReadTagHandler(tagService)
    writeTagH := command.NewWriteTagHandler(tagRepo, mqttClient)
    subTagH := command.NewSubscribeTagHandler(mqttClient)

    return &MCPServer{
        readTagH:  readTagH,
        writeTagH: writeTagH,
        subTagH:   subTagH,
        logger:    logger.New(cfg.LogLevel),
        license:   license.New(cfg.LicensePublicKeyPath),
        x402:      x402.NewHandler(cfg.X402Enabled, cfg.X402PaymentAddress),
        metrics:   metricsRepo,
    }
}

func (s *MCPServer) HandleRequest(ctx context.Context, req MCPRequest) (*MCPResponse, error) {
    start := time.Now()
    s.logger.Info().Str("method", req.Method).Msg("request received")

    if s.license.IsEnabled() {
        if err := s.license.Validate(); err != nil {
            return nil, &MCPError{Code: -32001, Message: "license invalid"}, nil
        }
    }

    if s.x402.IsEnabled() {
        if err := s.x402.ValidatePayment(req.Params); err != nil {
            return nil, err
        }
    }

    var resp interface{}
    var err error

    switch req.Method {
    case "read_tag":
        var q query.ReadTagQuery
        json.Unmarshal(req.Params, &q)
        resp, err = s.readTagH.Handle(ctx, q)
    case "write_tag":
        var c command.WriteTagCommand
        json.Unmarshal(req.Params, &c)
        resp, err = s.writeTagH.Handle(ctx, c)
    case "subscribe_tag":
        var c command.SubscribeTagCommand
        json.Unmarshal(req.Params, &c)
        resp, err = s.subTagH.Handle(ctx, c)
    default:
        err = errors.New("method not found")
    }

    duration := time.Since(start).Milliseconds()
    s.metrics.SaveRequest(ctx, repository.RequestMetrics{
        ID:        uuid.New().String(),
        Method:    req.Method,
        Params:    string(req.Params),
        Duration:  duration,
        Timestamp: time.Now(),
    })

    if err != nil {
        return &MCPResponse{
            JSONRPC: "2.0",
            Error:   &MCPError{Code: -32000, Message: err.Error()},
            ID:      req.ID,
        }, nil
    }

    return &MCPResponse{
        JSONRPC: "2.0",
        Result:  resp,
        ID:      req.ID,
    }, nil
}

func (s *MCPServer) Start(ctx context.Context) error {
    s.httpServer = &http.Server{
        Addr:    s.listenAddr,
        Handler: s,
    }
    return s.httpServer.ListenAndServe()
}
```

#### 3.5.2 MCP Middleware (`internal/interfaces/mcp/middleware.go`)
```go
package mcp

import (
    "context"
    "net/http"
    "time"
)

func (s *MCPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/health" {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"ok"}`))
        return
    }

    if r.Header.Get("Content-Type") != "application/json" {
        http.Error(w, "application/json required", http.StatusBadRequest)
        return
    }

    var req MCPRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid JSON", http.StatusBadRequest)
        return
    }

    resp, _ := s.HandleRequest(r.Context(), req)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}
```

---

### 3.6 License Module (Модуль лицензирования)

#### 3.6.1 Hardware Info (`pkg/license/hardware.go`)
```go
package license

import (
    "crypto/sha256"
    "encoding/hex"
    "os"
    "runtime"
)

type HardwareInfo struct {
    CPUID    string
    MACAddr  string
    Hostname string
}

func GetHardwareInfo() (HardwareInfo, error) {
    hostname, _ := os.Hostname()
    cpuID := getCPUID()
    macAddr := getMACAddress()

    return HardwareInfo{
        CPUID:    cpuID,
        MACAddr:  macAddr,
        Hostname: hostname,
    }, nil
}

func (h HardwareInfo) Hash() string {
    data := h.CPUID + h.MACAddr + h.Hostname
    hash := sha256.Sum256([]byte(data))
    return hex.EncodeToString(hash[:])
}

func getCPUID() string {
    var cpuInfo [32]byte
    runtime.RunCPUProfileSample(cpuInfo[:], 0, 0)
    return hex.EncodeToString(cpuInfo[:])
}

func getMACAddress() string {
    return "00:00:00:00:00:00"
}
```

#### 3.6.2 RSA Crypto (`pkg/license/crypto.go`)
```go
package license

import (
    "crypto"
    "crypto/rsa"
    "crypto/x509"
    "encoding/base64"
    "encoding/pem"
    "os"
)

type RSACrypto struct {
    publicKey *rsa.PublicKey
}

func NewRSACrypto(publicKeyPath string) (*RSACrypto, error) {
    keyData, err := os.ReadFile(publicKeyPath)
    if err != nil {
        return nil, err
    }

    block, _ := pem.Decode(keyData)
    pub, err := x509.ParsePKIXPublicKey(block.Bytes)
    if err != nil {
        return nil, err
    }

    return &RSACrypto{publicKey: pub.(*rsa.PublicKey)}, nil
}

func (r *RSACrypto) Verify(data, signature string) bool {
    sigBytes, _ := base64.StdEncoding.DecodeString(signature)
    hash := sha256.Sum256([]byte(data))
    return rsa.VerifyPKCS1v15(r.publicKey, crypto.SHA256, hash[:], sigBytes) == nil
}

func (r *RSACrypto) Sign(data string) (string, error) {
    hash := sha256.Sum256([]byte(data))
    sig, err := rsa.SignPKCS1v15(nil, crypto.SHA256, hash[:], nil)
    if err != nil {
        return "", err
    }
    return base64.StdEncoding.EncodeToString(sig), nil
}
```

#### 3.6.3 License Validator (`pkg/license/validator.go`)
```go
package license

import (
    "os"
    "time"
)

type LicenseValidator struct {
    crypto       *RSACrypto
    hardware     HardwareInfo
    enabled      bool
    checkInterval time.Duration
}

type LicenseData struct {
    HardwareHash string    `json:"hardware_hash"`
    ExpiresAt    time.Time `json:"expires_at"`
    Features     []string   `json:"features"`
    Signature    string     `json:"signature"`
}

func New(publicKeyPath string) *LicenseValidator {
    hw, _ := GetHardwareInfo()
    return &LicenseValidator{
        crypto:       NewRSACrypto(publicKeyPath),
        hardware:     hw,
        enabled:      os.Getenv("LICENSE_ENABLED") == "true",
        checkInterval: time.Hour,
    }
}

func (v *LicenseValidator) IsEnabled() bool {
    return v.enabled
}

func (v *LicenseValidator) Validate() error {
    if !v.enabled {
        return nil
    }
    licenseData, err := v.loadLicenseFile()
    if err != nil {
        return err
    }

    if time.Now().After(licenseData.ExpiresAt) {
        return ErrLicenseExpired
    }

    if licenseData.HardwareHash != v.hardware.Hash() {
        return ErrHardwareMismatch
    }

    if !v.crypto.Verify(v hardware.Hash(), licenseData.Signature) {
        return ErrInvalidSignature
    }

    return nil
}

func (v *LicenseValidator) loadLicenseFile() (*LicenseData, error) {
    data, err := os.ReadFile("/app/license/license.dat")
    if err != nil {
        return nil, err
    }
    var license LicenseData
    json.Unmarshal(data, &license)
    return &license, nil
}

var (
    ErrLicenseExpired     = errors.New("license expired")
    ErrHardwareMismatch   = errors.New("hardware mismatch")
    ErrInvalidSignature   = errors.New("invalid signature")
)
```

---

### 3.7 Payment x402 Module

#### 3.7.1 x402 Client (`pkg/x402/client.go`)
```go
package x402

import (
    "context"
    "encoding/json"
    "net/http"
)

type x402Client struct {
    paymentAddress string
    httpClient     *http.Client
}

type PaymentRequest struct {
    Amount      string `json:"amount"`
    Currency    string `json:"currency"`
    Destination string `json:"destination"`
    Memo        string `json:"memo"`
}

type PaymentResponse struct {
    Paid       bool   `json:"paid"`
    PaymentURL string `json:"payment_url,omitempty"`
}

func NewX402Client(paymentAddress string) *x402Client {
    return &x402Client{
        paymentAddress: paymentAddress,
        httpClient:     &http.Client{},
    }
}

func (c *x402Client) CreatePayment(ctx context.Context, amount, currency, memo string) (*PaymentResponse, error) {
    req := PaymentRequest{
        Amount:      amount,
        Currency:    currency,
        Destination: c.paymentAddress,
        Memo:        memo,
    }

    data, _ := json.Marshal(req)
    resp, err := c.httpClient.Post("https://pay.example.com/api/v1/payment", "application/json", bytes.NewReader(data))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var paymentResp PaymentResponse
    json.NewDecoder(resp.Body).Decode(&paymentResp)
    return &paymentResp, nil
}
```

#### 3.7.2 x402 Handler (`pkg/x402/handler.go`)
```go
package x402

import (
    "context"
    "errors"
    "os"
)

type Handler struct {
    enabled        bool
    paymentAddress  string
    client         *x402Client
}

func NewHandler(enabled bool, paymentAddress string) *Handler {
    return &Handler{
        enabled:        enabled,
        paymentAddress: paymentAddress,
        client:         NewX402Client(paymentAddress),
    }
}

func (h *Handler) IsEnabled() bool {
    return h.enabled
}

func (h *Handler) ValidatePayment(ctx context.Context, params json.RawMessage) error {
    if !h.enabled {
        return nil
    }

    var paymentReq struct {
        Payment *PaymentRequest `json:"x402_payment,omitempty"`
    }

    if err := json.Unmarshal(params, &paymentReq); err != nil {
        return err
    }

    if paymentReq.Payment == nil {
        return ErrPaymentRequired
    }

    if paymentReq.Payment.Destination != h.paymentAddress {
        return ErrInvalidPaymentDestination
    }

    return nil
}

var (
    ErrPaymentRequired             = errors.New("payment required")
    ErrInvalidPaymentDestination   = errors.New("invalid payment destination")
)
```

---

### 3.8 Logger Module (Модуль логирования)

#### 3.8.1 ZeroLog Logger (`pkg/logger/zap.go`)
```go
package logger

import (
    "os"
    "time"

    "github.com/rs/zerolog"
)

type Logger struct {
    zerolog.Logger
}

func New(level string) *Logger {
    zerolog.TimeFieldFormat = time.RFC3339
    log := zerolog.New(os.Stdout).
        Level(parseLevel(level)).
        With().
        Timestamp().
        Caller().
        Logger()
    return &Logger{Logger: log}
}

func parseLevel(level string) zerolog.Level {
    switch level {
    case "debug":
        return zerolog.DebugLevel
    case "info":
        return zerolog.InfoLevel
    case "warn":
        return zerolog.WarnLevel
    case "error":
        return zerolog.ErrorLevel
    default:
        return zerolog.InfoLevel
    }
}

func (l *Logger) WithRequestID(reqID string) *Logger {
    return &Logger{
        Logger: l.With().Str("request_id", reqID).Logger(),
    }
}
```

#### 3.8.2 Metrics Logger (`pkg/logger/metrics.go`)
```go
package logger

import (
    "context"
    "encoding/json"
    "os"
    "sync"
    "time"
)

type RequestMetric struct {
    ID          string    `json:"id"`
    Method      string    `json:"method"`
    Path        string    `json:"path"`
    Duration    int64     `json:"duration_ms"`
    StatusCode  int       `json:"status_code"`
    ClientIP    string    `json:"client_ip"`
    Timestamp   time.Time `json:"timestamp"`
    UserAgent   string    `json:"user_agent"`
    LicenseID   string    `json:"license_id,omitempty"`
    IsPaid      bool      `json:"is_paid"`
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
    return m.file.Close()
}
```

---

## 4. Конфигурация через ENV

### config.yaml
```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

mqtt:
  broker_url: "tcp://localhost:1883"
  client_id: "mcp_server"
  username: ""
  password: ""
  topic_prefix: "mcp/"
  qos: 0

opcua:
  endpoint: "opc.tcp://localhost:4840"
  security_mode: "None"
  cert_file: ""
  key_file: ""

license:
  enabled: true
  public_key_path: "/app/license/public.pem"
  check_interval: 3600

x402:
  enabled: true
  payment_address: "0x..."

metrics:
  enabled: true
  file: "/app/logs/requests.jsonl"
  buffer_size: 1000

logging:
  level: "info"
  format: "json"
```

### ENV Mapping
| ENV Variable | config.yaml path | Type |
|---|---|---|
| `APP_HOST` | `server.host` | string |
| `APP_PORT` | `server.port` | int |
| `MQTT_BROKER_URL` | `mqtt.broker_url` | string |
| `MQTT_CLIENT_ID` | `mqtt.client_id` | string |
| `OPCUA_ENDPOINT` | `opcua.endpoint` | string |
| `LICENSE_ENABLED` | `license.enabled` | bool |
| `LICENSE_PUBLIC_KEY_PATH` | `license.public_key_path` | string |
| `X402_ENABLED` | `x402.enabled` | bool |
| `X402_PAYMENT_ADDRESS` | `x402.payment_address` | string |
| `METRICS_FILE` | `metrics.file` | string |
| `LOG_LEVEL` | `logging.level` | string |

---

## 5. Сборка и запуск

### Makefile
```makefile
.PHONY: build run test lint clean docker-build docker-run

build:
    CGO_ENABLED=0 GOOS=linux go build -o bin/server ./cmd/server

run:
    go run ./cmd/server

test:
    go test -v -cover ./...

lint:
    golangci-lint run ./...

clean:
    rm -rf bin/

docker-build:
    docker build -t industrial-mcp:latest .

docker-run:
    docker-compose up -d

docker-stop:
    docker-compose down
```

### Запуск
```bash
# Клонирование и сборка
git clone https://github.com/vpomo/industrial-mcp.git
cd industrial-mcp

# Настройка .env
cp .env.example .env
# Редактирование .env

# Docker сборка и запуск
make docker-build
make docker-run

# Проверка здоровья
curl http://localhost:8080/health
```

### Тестирование MCP
```bash
# Чтение тэга
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "read_tag",
    "params": {"tag_name": "temperature"},
    "id": 1
  }'

# Запись тэга
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "write_tag",
    "params": {"tag_name": "temperature", "value": 25.5},
    "id": 2
  }'
```

---

## Принципы LiteDDD, применяемые в проекте

| Принцип | Реализация |
|---|---|
| Domain-First | `internal/domain/` — начинаем с сущностей и интерфейсов |
| Сущности как стражи | Поля `struct` приватные, доступ через методы |
| Именование поведения | `UpdateValue()`, `IsExpired()`, `HasFeature()` |
| Валидация при создании | `NewTag()`, `NewLicense()` возвращают ошибку |
| Команды как точки входа | `WriteTagCommand`, `SubscribeTagCommand` |
| Сервис приложения | Хендлеры координации в `application/` |
| Изоляция бизнес-логики | `domain/` не импортирует внешние пакеты |
| Избегаем анемичных моделей | Бизнес-логика внутри методов сущностей |

---

## План по шагам (Checklist)

- [x] 1. Инициализация Go модуля, go.mod
- [x] 2. Создание структуры папок
- [x] 3. Dockerfile и docker-compose.yml
- [x] 4. .env.example и config.yaml
- [x] 5. Domain: entity/tag.go, entity/license.go
- [x] 6. Domain: repository interfaces
- [x] 7. Domain: service/tag_service.go
- [x] 8. Application: queries и commands
- [x] 9. Infrastructure: MQTT client
- [x] 10. Infrastructure: OPC UA client
- [x] 11. Infrastructure: in-memory repositories
- [x] 12. Interfaces: MCP server
- [x] 13. Interfaces: middleware
- [x] 14. pkg/license: hardware, crypto, validator
- [x] 15. pkg/x402: client, handler
- [x] 16. pkg/logger: zap, metrics
- [x] 17. cmd/server/main.go
- [x] 18. Makefile
- [x] 19. Тесты
- [x] 20. Линтер
- [x] 21. Docker сборка и запуск