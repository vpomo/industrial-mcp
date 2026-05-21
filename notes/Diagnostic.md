# Диагностика MCP через HTTP

Пошаговая инструкция для проверки работы MCP-сервера средствами HTTP (`curl`, Postman, браузерные REST-клиенты).

Сервер поднимает один HTTP listener (по умолчанию `0.0.0.0:8080`). MCP-методы вызываются **JSON-RPC 2.0** через `POST` на корень `/`. Отдельного пути `/mcp` нет.

---

## 0. Подготовка

### 0.1. Структура файлов

При запуске с `-config cmd/server/config.yaml`:

```
cmd/server/
  config.yaml      # конфигурация
  license.dat      # лицензия (обязательна, если license.enabled: true)
  key/public.pem   # публичный ключ (путь из config, относительно папки config)
```

### 0.2. Сборка и запуск

```bash
cd /path/to/industrial-mcp

# лицензия (если license.enabled: true)
go run ./cmd/license-tool export-hwid
go run ./cmd/license-tool create \
  --hardware-hash <HWID> \
  --expires 2027-12-31 \
  --output cmd/server/license.dat \
  --private-key pkg/license/keys/private.pem

# запуск
go run ./cmd/server -config cmd/server/config.yaml
```

**Ожидаемый вывод в логах (успех):**

```text
starting MCP server
license validated file=.../cmd/server/license.dat
server started addr=0.0.0.0:8080
```

Если `license.dat` отсутствует или невалиден — процесс завершится до прослушивания порта (см. [Errors.md](./Errors.md#ошибки-при-старте-процесса)).

### 0.3. Конфиг для «чистой» диагностики MCP

Чтобы не мешали x402 и лицензия, временно в `config.yaml`:

```yaml
license:
  enabled: false

x402:
  enabled: false
```

Для проверки полного продакшен-поведения оставьте `enabled: true` и подготовьте `license.dat` + корректный `x402_payment` в params.

---

## 1. Проверка: процесс слушает порт

### Запрос

```bash
curl -s -w "\nHTTP %{http_code}\n" http://localhost:8080/health
```

### Вход

- Метод: `GET`
- Путь: `/health`
- Тело: не требуется
- Лицензия: **не проверяется**

### Успешный ответ

**HTTP 200**

```json
{"status":"ok"}
```

### Интерпретация

| Результат | Значение |
|-----------|----------|
| HTTP 200 + `ok` | HTTP-сервер запущен |
| Connection refused | Сервер не запущен или другой порт |
| Таймаут | Сеть / firewall / неверный host |

`/health` **не** проверяет MCP-методы, лицензию и MQTT.

---

## 2. Проверка: запись тега (`write_tag`)

### Запрос

```bash
curl -s -X POST http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "write_tag",
    "params": {
      "tag_name": "temperature",
      "value": 25.5
    },
    "id": 1
  }'
```

### Вход (тело JSON-RPC)

| Поле | Тип | Обязательно | Описание |
|------|-----|-------------|----------|
| `jsonrpc` | string | да | `"2.0"` |
| `method` | string | да | `"write_tag"` |
| `params.tag_name` | string | да | Имя тега, не пустое |
| `params.value` | any | да | Число, строка, bool и т.д. |
| `id` | number/string | рекомендуется | Корреляция запроса |

**HTTP:** `POST /`, заголовок `Content-Type: application/json`.

### Успешный ответ

**HTTP 200** (даже при логической ошибке в JSON-RPC смотрите поле `error`).

```json
{
  "jsonrpc": "2.0",
  "result": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "success": true
  },
  "id": 1
}
```

`result.id` — UUID созданного тега в памяти (меняется при каждой записи с новым объектом).

### Типичные ошибки

| Ситуация | Пример ответа |
|----------|----------------|
| Пустое `tag_name` | `"error": {"code": -32000, "message": "tag name is required"}` |
| Нет лицензии | `"error": {"code": -32001, "message": "license invalid"}` |
| x402 без оплаты | `"error": {"code": -32002, "message": "payment required"}` |

Подробнее: [Errors.md](./Errors.md).

---

## 3. Проверка: чтение тега (`read_tag`)

Сначала выполните шаг 2 (`write_tag`), иначе тега не будет в памяти.

### Запрос

```bash
curl -s -X POST http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "read_tag",
    "params": {
      "tag_name": "temperature"
    },
    "id": 2
  }'
```

### Вход

| Поле | Тип | Обязательно |
|------|-----|-------------|
| `params.tag_name` | string | да |

### Успешный ответ

```json
{
  "jsonrpc": "2.0",
  "result": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "temperature",
    "value": 25.5,
    "timestamp": "2026-05-20T12:34:56Z",
    "quality": 0
  },
  "id": 2
}
```

| Поле `result` | Значение |
|---------------|----------|
| `quality` | `0` — Good, `1` — Uncertain, `2` — Bad |

### Ошибка: тег не найден

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32000,
    "message": "tag not found"
  },
  "id": 2
}
```

---

## 4. Проверка: подписка (`subscribe_tag`)

Требуется **работающий MQTT-брокер** (`mqtt.broker_url` в config). При недоступном брокере клиент MQTT не создаётся — вызов `subscribe_tag` приведёт к панике (nil subscriber) или ошибке подключения на этапе старта (MQTT только warning).

### Запрос

```bash
curl -s -X POST http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "subscribe_tag",
    "params": {
      "tag_name": "temperature",
      "qos": 0
    },
    "id": 3
  }'
```

### Вход

| Поле | Тип | Обязательно |
|------|-----|-------------|
| `params.tag_name` | string | да |
| `params.qos` | int | нет (в коде не передаётся в MQTT, всегда QoS 0) |

### Успешный ответ

```json
{
  "jsonrpc": "2.0",
  "result": {
    "subscription_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
    "topic": "mcp/tag/temperature"
  },
  "id": 3
}
```

Фактическая подписка MQTT: `{topic_prefix}mcp/tag/{tag_name}` (по умолчанию `mcp/mcp/tag/temperature` при `topic_prefix: "mcp/"`).

### Ошибка MQTT

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32000,
    "message": "connection refused"
  },
  "id": 3
}
```

Текст `message` зависит от ошибки брокера.

---

## 5. Проверка: неизвестный метод

### Запрос

```bash
curl -s -X POST http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "list_tags",
    "params": {"limit": 100},
    "id": 4
  }'
```

### Ответ

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32601,
    "message": "method not found"
  },
  "id": 4
}
```

Метод `list_tags` реализован в `internal/application/query`, но **не подключён** к HTTP MCP (`HandleRequest`). В README он указан ошибочно для текущей версии HTTP API.

---

## 6. Проверка: ошибки HTTP-транспорта

### 6.1. Без `Content-Type: application/json`

```bash
curl -s -w "\nHTTP %{http_code}\n" -X POST http://localhost:8080/ \
  -d '{"jsonrpc":"2.0","method":"read_tag","params":{},"id":1}'
```

**HTTP 400**, тело plain text:

```text
application/json required
```

JSON-RPC-обёртки нет.

### 6.2. Невалидный JSON

```bash
curl -s -w "\nHTTP %{http_code}\n" -X POST http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -d 'not json'
```

**HTTP 400**:

```text
invalid JSON
```

---

## 7. Проверка лицензии через MCP

При `license.enabled: true` каждый `POST /` (кроме того, что `/health` обходит лицензию) вызывает `license.Validate()`.

### Успех

Те же ответы, что в шагах 2–4, без поля `error`.

### Сбой лицензии

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32001,
    "message": "license invalid"
  },
  "id": 1
}
```

В логах сервера:

```text
license validation failed request_id=... error=license not found
```

Внутренние причины (`license not found`, `license expired`, …) в JSON-RPC **не** раскрываются — только `license invalid`. См. [Errors.md](./Errors.md#лицензия).

---

## 8. Проверка x402 (если включено)

При `x402.enabled: true` в `params` нужен блок оплаты:

```json
{
  "jsonrpc": "2.0",
  "method": "write_tag",
  "params": {
    "tag_name": "temperature",
    "value": 25.5,
    "x402_payment": {
      "amount": "0.01",
      "currency": "USDC",
      "destination": "0x...",
      "memo": "mcp-write"
    }
  },
  "id": 1
}
```

`destination` должен совпадать с `x402.payment_address` в config.

### Без оплаты

```json
{
  "error": {
    "code": -32002,
    "message": "payment required"
  }
}
```

### Неверный адрес

```json
{
  "error": {
    "code": -32002,
    "message": "invalid payment destination"
  }
}
```

---

## 9. Сквозной сценарий (чеклист)

| # | Действие | Ожидание |
|---|----------|----------|
| 1 | `GET /health` | HTTP 200, `{"status":"ok"}` |
| 2 | `write_tag` temperature=25.5 | `result.success: true` |
| 3 | `read_tag` temperature | `result.value: 25.5` |
| 4 | `read_tag` unknown | `error`, `tag not found` |
| 5 | `method: "foo"` | `error`, code `-32601` |
| 6 | POST без Content-Type | HTTP 400 |
| 7 | (опционально) `subscribe_tag` | `result.topic` при живом MQTT |

---

## 10. Логи и метрики

| Источник | Что смотреть |
|----------|----------------|
| stdout сервера | `request received`, `request completed`, `license validation failed` |
| `metrics.file` | JSONL с полями `method`, `params`, `duration` (если репозиторий метрик создан) |

Пример строки метрик (если файл доступен):

```json
{"id":"...","method":"write_tag","params":"{\"tag_name\":\"temperature\",\"value\":25.5}","duration":2,"timestamp":1716200000}
```

---

## 11. Переменные окружения

| Переменная | Эффект |
|------------|--------|
| `LICENSE_ENABLED=false` | Отключает проверки внутри `Validator`, даже если валидатор создан |
| `X402_ENABLED=true` | Альтернатива config (только при `NewHandlerFromEnv`, в `cmd/server` используется config) |

---

## Сводка эндпоинтов

| HTTP | Путь | Назначение |
|------|------|------------|
| GET | `/health` | Жив ли процесс |
| POST | `/` | JSON-RPC: `read_tag`, `write_tag`, `subscribe_tag` |

Коды ошибок JSON-RPC и HTTP: [Errors.md](./Errors.md).
