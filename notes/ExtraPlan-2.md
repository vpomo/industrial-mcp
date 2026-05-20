# План реализации: Персистентное хранение и идентификация агента

## Принцип: Self-registration + Multi-payment protocol

**Не нужен администратор.** Агент идентифицируется через:
1. **Agent ID** — произвольный идентификатор (UUID), агент сам генерирует
2. **Оплата** — X402 или MPP (один из них, или оба)

---

## Ключевая идея

```
Agent генерирует UUID → сохраняет в своем config
                        ↓
При запросе: Agent ID + Payment (X402 или MPP)
                        ↓
MCP проверяет платеж → валиден
                        ↓
Создает/находит Agent по Agent ID
                        ↓
Все данные агента привязаны к Agent ID
```

---

## Этап 1: Agent identity (произвольный ID)

### 1.1 Agent entity
**Файл:** `internal/domain/entity/agent.go`

```go
type Agent struct {
    id        string  // UUID, генерируется агентом
    name      string  // опционально, для отображения
    active    bool
    createdAt time.Time
    lastSeenAt time.Time
}
```

### 1.2 AgentService (self-registration)
**Файл:** `internal/domain/service/agent_service.go`

```go
type AgentService struct {
    repo      repository.AgentRepository
    payments  []PaymentValidator  // X402, MPP
}

type PaymentValidator interface {
    Validate(ctx context.Context, r *http.Request) (*PaymentInfo, error)
    Supported() []PaymentProtocol
}

type PaymentInfo struct {
    AgentID string  // может совпадать с payment identity или отдельный
    Amount  int64
    Protocol PaymentProtocol
}

func (s *AgentService) GetOrCreateAgent(ctx context.Context, agentID string) (*Agent, error) {
    agent, err := s.repo.GetByID(ctx, agentID)
    if err == nil {
        agent.UpdateLastSeen()
        s.repo.Save(ctx, agent)
        return agent, nil
    }

    if errors.Is(err, repository.ErrAgentNotFound) {
        agent = entity.NewAgent(agentID)
        if err := s.repo.Save(ctx, agent); err != nil {
            return nil, err
        }
        return agent, nil
    }

    return nil, err
}
```

### 1.3 Payment validators
**Файл:** `internal/domain/service/payment_validator.go`

```go
type X402Validator struct{}
type MPPValidator struct{}

func (v *X402Validator) Validate(ctx context.Context, r *http.Request) (*PaymentInfo, error) {
    // Извлекаем X402 payment из заголовков
    payment := r.Header.Get("X-Payment")
    if payment == "" {
        return nil, errors.New("X402 payment required")
    }
    // Валидация X402...
    return &PaymentInfo{
        AgentID: extractAgentID(payment),
        Amount:  extractAmount(payment),
        Protocol: ProtocolX402,
    }, nil
}

func (v *MPPValidator) Validate(ctx context.Context, r *http.Request) (*PaymentInfo, error) {
    // Извлекаем MPP payment
    mpp := r.Header.Get("X-MPP-Payment")
    if mpp == "" {
        return nil, errors.New("MPP payment required")
    }
    // Валидация MPP...
    return &PaymentInfo{
        AgentID: extractAgentID(mpp),
        Amount:  extractAmount(mpp),
        Protocol: ProtocolMPP,
    }, nil
}
```

### 1.4 Auth middleware (Multi-protocol)
**Файл:** `internal/interfaces/middleware/auth.go`

```go
func MultiPaymentAuthMiddleware(validators ...PaymentValidator) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var paymentInfo *PaymentInfo
        var err error

        // Пробуем каждый валидатор
        for _, v := range validators {
            paymentInfo, err = v.Validate(r.Context(), r)
            if err == nil {
                break
            }
        }

        if paymentInfo == nil {
            http.Error(w, "Valid payment required (X402 or MPP)", 402)
            return
        }

        agentID := paymentInfo.AgentID

        // Get or create agent
        agent, err := agentService.GetOrCreateAgent(r.Context(), agentID)
        if err != nil {
            http.Error(w, "Agent error", 500)
            return
        }

        ctx := context.WithValue(r.Context(), "agent", agent)
        ctx = context.WithValue(ctx, "payment", paymentInfo)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

---

## Этап 2: Модель данных (SQLite/PostgreSQL)

### 2.1 Таблицы

```sql
-- Агенты
CREATE TABLE agents (
    id TEXT PRIMARY KEY,
    name TEXT,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    last_seen_at TIMESTAMP DEFAULT NOW()
);

-- Платежи агента (история)
CREATE TABLE agent_payments (
    id UUID PRIMARY KEY,
    agent_id TEXT REFERENCES agents(id),
    protocol VARCHAR(10) NOT NULL,  -- 'x402' или 'mpp'
    amount INTEGER NOT NULL,
    paid_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_payments_agent ON agent_payments(agent_id);

-- Источники данных
CREATE TABLE data_sources (
    id UUID PRIMARY KEY,
    agent_id TEXT REFERENCES agents(id),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    config JSONB NOT NULL,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_data_sources_agent ON data_sources(agent_id);

-- Теги
CREATE TABLE exposed_tags (
    id UUID PRIMARY KEY,
    agent_id TEXT REFERENCES agents(id),
    data_source_id UUID REFERENCES data_sources(id),
    name VARCHAR(255) NOT NULL,
    node_id VARCHAR(500) NOT NULL,
    data_type VARCHAR(50) NOT NULL,
    read_only BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_exposed_tags_agent ON exposed_tags(agent_id);

-- Результаты сканирования
CREATE TABLE scan_results (
    id UUID PRIMARY KEY,
    agent_id TEXT REFERENCES agents(id),
    data_source_id UUID REFERENCES data_sources(id),
    node_id VARCHAR(500) NOT NULL,
    name VARCHAR(255),
    data_type VARCHAR(50),
    discovered_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_scan_results_agent ON scan_results(agent_id);
```

### 2.2 Repository interfaces

```go
type AgentRepository interface {
    GetByID(ctx context.Context, id string) (*Agent, error)
    Save(ctx context.Context, agent *Agent) error
}

type PaymentRepository interface {
    Save(ctx context.Context, payment *Payment) error
    ListByAgent(ctx context.Context, agentID string) ([]*Payment, error)
}

type DataSourceRepository interface {
    GetByAgent(ctx context.Context, agentID string) ([]*DataSource, error)
    Save(ctx context.Context, ds *DataSource) error
    // ...
}
```

---

## Этап 3: Хранение результатов сканирования

### 3.1 ScanResult entity
**Файл:** `internal/domain/entity/scan_result.go`

```go
type ScanResult struct {
    id            string
    agentID       string
    dataSourceID  string
    nodeID        string
    name          string
    dataType      string
    discoveredAt  time.Time
}
```

### 3.2 Scan service
```go
func (s *ScanService) ScanAndSave(ctx context.Context, dsID string) ([]*entity.ScanResult, error) {
    agentID := getAgentIDFromContext(ctx)

    // 1. Сканируем
    results, err := s.dsService.Scan(ctx, dsID)

    // 2. Сохраняем с привязкой к agent
    for _, r := range results {
        scanResult := entity.NewScanResult(agentID, dsID, r.NodeID, r.Name, r.DataType)
        s.scanRepo.Save(ctx, scanResult)
    }

    return results, nil
}
```

---

## Этап 4: MCP интеграция

### 4.1 Контекст агента
```go
type AgentContext struct {
    AgentID   string
    Agent     *Agent
    Payment   *PaymentInfo
}

func GetAgentIDFromContext(ctx context.Context) string {
    return ctx.Value("agent_id").(string)
}
```

### 4.2 MCP handlers используют agentID из контекста
```go
func (h *DataSourceHandler) Create(w http.ResponseWriter, r *http.Request) {
    agentID := GetAgentIDFromContext(r.Context())

    var req CreateDataSourceRequest
    json.NewDecoder(r.Body).Decode(&req)

    ds, _ := h.svc.Create(r.Context(), agentID, req.Name, req.Type, req.Config)
    // ...
}
```

---

## Этап 5: Миграция с memory на DB

### 5.1 SQLite implementation
**Файл:** `internal/infrastructure/repository/sqlite/sqlite.go`

```go
func NewSQLiteDB(path string) (*SQLiteDB, error) {
    db, err := sql.Open("sqlite3", path)
    if err != nil {
        return nil, err
    }

    db.Exec(`
        CREATE TABLE IF NOT EXISTS agents (
            id TEXT PRIMARY KEY,
            name TEXT,
            active INTEGER DEFAULT 1,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            last_seen_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
        CREATE TABLE IF NOT EXISTS agent_payments (
            id TEXT PRIMARY KEY,
            agent_id TEXT REFERENCES agents(id),
            protocol VARCHAR(10) NOT NULL,
            amount INTEGER NOT NULL,
            paid_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
        CREATE TABLE IF NOT EXISTS data_sources (...);
        CREATE TABLE IF NOT EXISTS exposed_tags (...);
        CREATE TABLE IF NOT EXISTS scan_results (...);
    `)
    return &SQLiteDB{db: db}, nil
}
```

---

## Файлы для реализации

```
internal/domain/entity/
  agent.go              # Agent (UUID as ID) (NEW)
  scan_result.go        # ScanResult (NEW)
  payment.go            # Payment entity (NEW)

internal/domain/repository/
  agent_repository.go   # Agent repo (NEW)
  payment_repository.go # Payment repo (NEW)

internal/domain/service/
  agent_service.go      # Self-registration (NEW)
  payment_validator.go  # X402 + MPP validators (NEW)

internal/infrastructure/repository/
  sqlite/               # SQLite implementation (NEW)

internal/interfaces/middleware/
  auth.go               # Multi-payment auth (NEW)

internal/interfaces/mcp/
  context.go            # Agent context (UPDATE)
```

---

## Как работает

### Первый запуск агента

```
1. Агент генерирует UUID: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
2. Агент сохраняет в config: ~/.industrial-mcp/agent.json
3. Агент отправляет запрос: Agent-ID + Payment (X402 или MPP)
4. Auth middleware: валидирует платеж, извлекает Agent ID
5. AgentService.GetOrCreateAgent():
   - Создает нового Agent
6. Все операции привязаны к Agent ID
```

### Агент возвращается через день

```
1. Агент читает Agent ID из config
2. Отправляет новый Payment
3. Находит существующего Agent
4. last_seen_at обновляется
5. Агент видит свои данные
```

---

## Отличие от крипто-привязки

| | Wallet as ID | UUID + Multi-payment |
|--|-------------|---------------------|
| Agent ID | Хаш кошелька | Произвольный UUID |
| Регистрация | Привязана к блокчейну | Агент генерирует сам |
| Payments | Только один протокол | X402 + MPP |
| Потеря wallet | Теряет доступ | Можно создать новый ID |

---

## Порядок реализации

1. **Этап 1** — Agent entity, Payment entity, Payment validators (X402 + MPP)
2. **Этап 2** — Модель данных, SQLite схема, repositories
3. **Этап 3** — Хранение результатов сканирования
4. **Этап 4** — MCP интеграция (agentID context)
5. **Этап 5** — Миграция с memory на DB

---

## Пример API

### Агент отправляет запрос (X402)

```bash
curl -X POST http://localhost:8080/api/v1/data-sources \
  -H "Content-Type: application/json" \
  -H "X-Agent-ID: a1b2c3d4-e5f6-7890-abcd-ef1234567890" \
  -H "X-Payment: pay_abc123..." \
  -d '{
    "name": "OpcuaServer1",
    "type": "opcua",
    "config": {"endpoint": "opc.tcp://localhost:4840"}
  }'
```

### Агент отправляет запрос (MPP)

```bash
curl -X POST http://localhost:8080/api/v1/data-sources \
  -H "Content-Type: application/json" \
  -H "X-Agent-ID: a1b2c3d4-e5f6-7890-abcd-ef1234567890" \
  -H "X-MPP-Payment: mpp_def456..." \
  -d '{
    "name": "OpcuaServer1",
    "type": "opcua",
    "config": {"endpoint": "opc.tcp://localhost:4840"}
  }'
```

### Агент сканирует источник

```bash
curl -X POST http://localhost:8080/api/v1/data-sources/{id}/scan \
  -H "X-Agent-ID: a1b2c3d4-e5f6-7890-abcd-ef1234567890" \
  -H "X-Payment: pay_abc123..."
```

---

## Ключевые преимущества

| До | После |
|----|-------|
| Данные в памяти | Персистентное хранение в БД |
| Нет идентификации | Agent ID (UUID) |
| Нужен администратор | Self-registration |
| Один протокол оплаты | X402 + MPP |
| Результаты не сохраняются | Scan results persisted |

---

## Q&A

**Q: Как агент узнает свой ID?**
A: Агент сам генерирует UUID при первом запуске и сохраняет в `~/.industrial-mcp/agent.json`.

**Q: Что если агент потеряет config?**
A: Создаст новый UUID = новый Agent ID = новые данные. Старые данные не восстановить (как и в любой системе с self-registration).

**Q: Какой payment выбрать - X402 или MPP?**
A: Оба поддерживаются параллельно. Агент выбирает любой.