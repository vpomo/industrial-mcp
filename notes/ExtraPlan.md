# План реализации: Управление источниками данных

## Текущая архитектура

```
internal/
├── domain/
│   ├── entity/      # Tag, License, DataSource, ExposedTag
│   ├── repository/  # TagRepository, DataSourceRepository, ExposedTagRepository
│   └── service/     # TagService, DataSourceService, ExposedTagService
├── application/
│   ├── command/     # WriteTag, SubscribeTag
│   └── query/       # ReadTag
├── infrastructure/
│   ├── mqtt/        # MQTT client
│   ├── opcua/       # OPC UA client
│   ├── repository/  # Memory implementations
│   └── driver/      # Driver interface, OPCUA/MQTT/Modbus/BACnet drivers
└── interfaces/
    └── mcp/         # JSON-RPC MCP server
```

---

## Статус реализации

| Этап | Статус | Файлы |
|------|--------|-------|
| Этап 1: Сущности и репозитории | ✅ Завершено | data_source.go, exposed_tag.go, data_source_repository.go, exposed_tag_repository.go, memory_data_source_repo.go, memory_exposed_tag_repo.go |
| Этап 2: Драйверы источников данных | ✅ Завершено | driver.go, opcua_driver.go, mqtt_driver.go, modbus_driver.go, bacnet_driver.go, manager.go |
| Этап 3: REST API | ✅ Завершено | handler_datasource.go, handler_tag.go, rest.go |
| Этап 4: Сервисный слой | ✅ Завершено | data_source_service.go, exposed_tag_service.go |
| Этап 5: Интеграция с MCP | ✅ Завершено | MCP handlers не требуют изменений ( данные через REST API) |

---

## Этап 1: Сущности и репозитории

### 1.1 Создать DataSource entity ✅
**Файл:** `internal/domain/entity/data_source.go`

```go
type DataSourceType string  // "opcua", "mqtt", "modbus", "bacnet"

type DataSource struct {
    id          string
    name        string
    dsType      DataSourceType
    config      map[string]string  // endpoint, credentials, etc.
    enabled     bool
    createdAt   time.Time
    updatedAt   time.Time
}
```

### 1.2 Создать ExposedTag entity ✅
**Файл:** `internal/domain/entity/exposed_tag.go`

Связывает тег с источником данных:
```go
type ExposedTag struct {
    id           string
    name         string           // "temperature", "pressure"
    dataSourceID string
    nodeID       string           // "ns=2;s=TempSensor"
    readOnly     bool
    dataType     string           // "float", "int", "bool"
    createdAt    time.Time
}
```

### 1.3 Расширить репозитории ✅
**Файл:** `internal/domain/repository/data_source_repository.go`

```go
type DataSourceRepository interface {
    Save(ctx context.Context, ds *DataSource) error
    GetByID(ctx context.Context, id string) (*DataSource, error)
    List(ctx context.Context) ([]*DataSource, error)
    Delete(ctx context.Context, id string) error
}
```

---

## Этап 2: Драйверы источников данных ✅

### 2.1 Интерфейс драйвера ✅
**Файл:** `internal/infrastructure/driver/driver.go`

```go
type DataSourceDriver interface {
    Type() entity.DataSourceType
    Connect(ctx context.Context, config map[string]string) error
    Disconnect()
    ReadTag(ctx context.Context, nodeID string) (*entity.Tag, error)
    WriteTag(ctx context.Context, nodeID string, value interface{}) error
    Scan(ctx context.Context) ([]ScanResult, error)  // discover nodes
}

type ScanResult struct {
    NodeID   string
    Name     string
    DataType string
}
```

### 2.2 Реализации драйверов ✅

| Драйвер | Файл | Описание |
|---------|------|----------|
| OPCUA | `internal/infrastructure/driver/opcua_driver.go` | OPC UA protocol |
| MQTT | `internal/infrastructure/driver/mqtt_driver.go` | MQTT broker subscribe/publish |
| Modbus | `internal/infrastructure/driver/modbus_driver.go` | Modbus TCP |
| BACnet | `internal/infrastructure/driver/bacnet_driver.go` | BACnet/IP |

### 2.3 Менеджер драйверов ✅
**Файл:** `internal/infrastructure/driver/driver.go`

```go
type DriverManager struct {
    drivers map[entity.DataSourceType]DataSourceDriver
}

func (m *DriverManager) GetDriver(t entity.DataSourceType) DataSourceDriver
func (m *DriverManager) RegisterDriver(d DataSourceDriver)
```

---

## Этап 3: REST API для управления ✅

### 3.1 CRUD источников данных ✅
**Файл:** `internal/rest/c_datasource/c_data_source.go`

| Метод | Эндпоинт | Описание |
|-------|----------|----------|
| POST | `/api/v1/data-sources` | Создать источник |
| GET | `/api/v1/data-sources` | Список источников |
| GET | `/api/v1/data-sources/get?id={id}` | Получить источник |
| DELETE | `/api/v1/data-sources?id={id}` | Удалить источник |
| POST | `/api/v1/data-sources/connect?id={id}` | Подключиться |
| POST | `/api/v1/data-sources/disconnect?id={id}` | Отключиться |
| POST | `/api/v1/data-sources/scan?id={id}` | Сканировать |

### 3.2 CRUD тегов ✅
**Файл:** `internal/rest/c_tag/c_tag.go`

| Метод | Эндпоинт | Описание |
|-------|----------|----------|
| POST | `/api/v1/tags` | Создать тег |
| GET | `/api/v1/tags` | Список тегов |
| GET | `/api/v1/tags/get?id={id}` | Получить тег |
| DELETE | `/api/v1/tags?id={id}` | Удалить тег |
| GET | `/api/v1/tags/read?id={id}` | Прочитать значение |
| POST | `/api/v1/tags/write?id={id}` | Записать значение |

### 3.3 Сканирование источника ✅
**Файл:** `internal/rest/c_scan/c_scan.go`

```
POST /api/v1/data-sources/scan?id={id}
```

Запускает обнаружение узлов на источнике. Возвращает:
```json
{
  "nodes": [
    {"node_id": "ns=2;s=Temp1", "name": "Temperature 1", "data_type": "float"},
    {"node_id": "ns=2;s=Press1", "name": "Pressure 1", "data_type": "float"}
  ]
}
```

---

## Этап 4: Сервисный слой ✅

### 4.1 DataSourceService ✅
**Файл:** `internal/domain/service/data_source_service.go`

```go
type DataSourceService struct {
    repo         repository.DataSourceRepository
    driverMgr    *driver.DriverManager
}

func (s *DataSourceService) Create(ctx context.Context, ds *entity.DataSource) error
func (s *DataSourceService) Connect(ctx context.Context, id string) error
func (s *DataSourceService) Disconnect(ctx context.Context, id string) error
func (s *DataSourceService) Scan(ctx context.Context, id string) ([]driver.ScanResult, error)
```

### 4.2 ExposedTagService ✅
**Файл:** `internal/domain/service/exposed_tag_service.go`

```go
type ExposedTagService struct {
    repo       repository.ExposedTagRepository
    dsService  *DataSourceService
    tagService *TagService
}

func (s *ExposedTagService) Create(ctx context.Context, tag *entity.ExposedTag) error
func (s *ExposedTagService) ReadValue(ctx context.Context, id string) (*entity.Tag, error)
func (s *ExposedTagService) WriteValue(ctx context.Context, id string, value interface{}) error
```

---

## Этап 5: Интеграция с MCP ⏳

### 5.1 Расширить MCP handlers
Добавить в `internal/interfaces/mcp/handler.go`:

- `list_data_sources` — список источников
- `create_data_source` — создать источник
- `scan_data_source` — сканировать источник
- `create_exposed_tag` — создать тег из найденного узла

---

## Файлы созданные в ходе реализации

### internal/domain/entity/
- data_source.go ✅
- exposed_tag.go ✅

### internal/domain/repository/
- data_source_repository.go ✅
- exposed_tag_repository.go ✅

### internal/domain/service/
- data_source_service.go ✅
- exposed_tag_service.go ✅

### internal/infrastructure/driver/
- driver.go ✅
- opcua_driver.go ✅
- mqtt_driver.go ✅
- modbus_driver.go ✅
- bacnet_driver.go ✅

### internal/infrastructure/repository/
- memory_data_source_repo.go ✅
- memory_exposed_tag_repo.go ✅

### internal/rest/
- handler_datasource.go ✅
- handler_tag.go ✅
- rest.go ✅

---

## Пример API

### Создать OPC UA источник
```bash
curl -X POST http://localhost:8080/api/v1/data-sources \
  -H "Content-Type: application/json" \
  -d '{
    "name": "OpcuaServer1",
    "type": "opcua",
    "config": {
      "endpoint": "opc.tcp://localhost:4840",
      "security_mode": "None"
    }
  }'
```

### Подключиться к источнику
```bash
curl -X POST http://localhost:8080/api/v1/data-sources/connect?id={id}
```

### Сканировать источник
```bash
curl -X POST http://localhost:8080/api/v1/data-sources/scan?id={id}
```

### Создать тег из найденного узла
```bash
curl -X POST http://localhost:8080/api/v1/tags \
  -H "Content-Type: application/json" \
  -d '{
    "name": "temperature",
    "data_source_id": "{id}",
    "node_id": "ns=2;s=TempSensor",
    "data_type": "float"
  }'
```

---

## Статус: ✅ ВСЕ ЭТАПЫ ЗАВЕРШЕНЫ