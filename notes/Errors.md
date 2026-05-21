# Справочник ошибок industrial-mcp

Ошибки сгруппированы по уровню: старт процесса, HTTP-транспорт, JSON-RPC (MCP), лицензия, x402, бизнес-логика handlers.

---

## 1. Уровни ответа

| Уровень | Формат | Когда |
|---------|--------|-------|
| Старт процесса | `log.Fatalf`, процесс завершается | До `ListenAndServe` |
| HTTP 400 | plain text | Неверный Content-Type или JSON тела |
| HTTP 200 + JSON-RPC `error` | JSON | Логическая ошибка MCP-запроса |
| HTTP 200 + JSON-RPC `result` | JSON | Успех |

**Важно:** при JSON-RPC ошибках HTTP-код обычно **200**. Смотрите поле `error` в теле.

---

## 2. Ошибки при старте процесса

Проявляются в stderr, сервер **не** слушает порт.

| Сообщение (префикс) | Причина | Действие |
|---------------------|---------|----------|
| `failed to resolve config path` | Некорректный путь `-config` | Указать существующий файл |
| `failed to load config` | Нет файла / битый YAML | Проверить `config.yaml` |
| `failed to read license public key` | Нет PEM по `license.public_key_path` | Положить ключ рядом с config или абсолютный путь |
| `license system error` | Ошибка `GetHardwareInfo()` при `license.New` | Права / окружение |
| `license validation failed at startup` | Невалидный `license.dat` рядом с config | Создать/обновить лицензию (`license-tool`) |

### Внутренние причины `license validation failed at startup`

| Текст `error=` | Значение |
|----------------|----------|
| `license not found` | Нет файла `license.dat` в папке config |
| `license corrupted` | Файл не JSON / не читается |
| `license expired` | `expires_at` в прошлом |
| `hardware mismatch` | `hardware_hash` ≠ HWID машины |
| `invalid signature` | Подпись не сходится с PEM (если ключ задан) |

Путь к лицензии: `{dirname(config.yaml)}/license.dat` (имя фиксировано).

### Периодическая проверка (каждые 20 мин)

При сбое в логах:

```text
periodic license validation failed file=... error=...
```

Сервер вызывает `cancel()` и завершает работу.

---

## 3. HTTP-транспорт (не JSON-RPC)

Источник: `internal/interfaces/mcp/middleware.go`.

| HTTP | Тело | Условие |
|------|------|---------|
| **400** | `application/json required` | Заголовок `Content-Type` ≠ `application/json` |
| **400** | `invalid JSON` | Тело не парсится как JSON |
| **200** | `{"status":"ok"}` | `GET /health` — не ошибка |

Любой другой путь, кроме `/health`, с `POST` идёт в JSON-RPC.

---

## 4. JSON-RPC: коды ошибок MCP

Источник: `internal/interfaces/mcp/server.go`.

| Code | Message (пример) | Когда |
|------|------------------|-------|
| **-32001** | `license invalid` | `license.Validate()` не прошла (детали в логах, не в JSON) |
| **-32002** | `payment required` | x402 включён, нет `params.x402_payment` |
| **-32002** | `invalid payment destination` | `x402_payment.destination` ≠ config |
| **-32002** | *(текст ошибки unmarshal)* | Битый JSON в `params` при разборе x402 |
| **-32601** | `method not found` | Неизвестный `method` |
| **-32000** | *(текст из handler)* | Ошибки handlers, парсинг `params`, прочее |

### Структура ответа с ошибкой

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32000,
    "message": "tag not found"
  },
  "id": 1
}
```

Поле `id` эхом возвращает `id` запроса (если был передан).

---

## 5. Лицензия

### 5.1. Пакет `pkg/license`

| Переменная / константа | Текст ошибки | Когда |
|------------------------|--------------|-------|
| `ErrLicenseNotFound` | `license not found` | Нет `license.dat` |
| `ErrLicenseCorrupted` | `license corrupted` | Ошибка чтения/парсинга |
| `ErrLicenseExpired` | `license expired` | `IsExpired()` |
| `ErrHardwareMismatch` | `hardware mismatch` | HWID не совпадает |
| `ErrInvalidSignature` | `invalid signature` | RSA verify failed |
| `ErrFeatureMissing` | `feature missing` | `ValidateFeature()` — REST, не MCP HTTP |

### 5.2. MCP HTTP

Все внутренние ошибки лицензии при запросе сводятся к:

```json
{"code": -32001, "message": "license invalid"}
```

### 5.3. Отключение проверки

| Условие | Эффект |
|---------|--------|
| `license.enabled: false` в config | Валидатор не создаётся, MCP без проверки |
| `LICENSE_ENABLED=false` | `Validator.IsEnabled()` → false, `Validate()` no-op |
| `GET /health` | Лицензия не проверяется никогда |

---

## 6. x402 (оплата)

Источник: `pkg/x402/handler.go`. Только если `x402.enabled: true` в config.

| Ошибка | Code MCP | message в JSON-RPC |
|--------|----------|-------------------|
| `ErrPaymentRequired` | -32002 | `payment required` |
| `ErrInvalidPaymentDestination` | -32002 | `invalid payment destination` |
| Ошибка `json.Unmarshal` params | -32002 | текст Go-ошибки, напр. `unexpected end of JSON input` |

### Ожидаемый фрагмент `params`

```json
{
  "x402_payment": {
    "amount": "0.01",
    "currency": "USDC",
    "destination": "<совпадает с payment_address в config>",
    "memo": "optional"
  }
}
```

---

## 7. Ошибки методов MCP (code -32000)

### 7.1. `write_tag`

| message | Причина |
|---------|---------|
| `tag name is required` | Пустой `params.tag_name` |
| Синтаксическая ошибка JSON в `params` | Неверная структура `params` |

Успешный `result`: `{"id":"<uuid>","success":true}`.

### 7.2. `read_tag`

| message | Причина |
|---------|---------|
| `tag not found` | Тег не записан через `write_tag` или другое имя |
| Ошибка unmarshal `params` | Нет/битый `tag_name` в JSON |

### 7.3. `subscribe_tag`

| message | Причина |
|---------|---------|
| Текст ошибки MQTT | Брокер недоступен, отказ подписки |
| panic (нет recovery) | MQTT-клиент не создан (`nil` subscriber) |

`params` должны содержать `tag_name`; `qos` в MQTT API не используется.

### 7.4. Неизвестный метод (code -32601)

| method | Ответ |
|--------|-------|
| `list_tags`, `foo`, … | `method not found` |

---

## 8. Ошибки домена и инфраструктуры (не HTTP MCP сейчас)

Эти ошибки определены в коде, но **не экспонируются** через текущий `cmd/server` HTTP API (REST не подключён).

### 8.1. Теги и сущности

| message | Пакет |
|---------|-------|
| `tag not found` | `domain/service`, `infrastructure/repository` |
| `tag name is required` | `domain/entity` |

### 8.2. Data source / drivers (будущий REST)

| message | Пакет |
|---------|-------|
| `data source not found` | repository |
| `data source name is required` | entity |
| `invalid data source type` | entity |
| `driver not found for data source type` | service |
| `data source not connected` | service |
| `data source ID is required` | service / entity |
| `tag is read-only` | exposed tag service |
| `exposed tag not found` | repository |
| `node ID is required` | entity |

### 8.3. OPC UA / drivers

| Тип | message |
|-----|---------|
| `OpcuaError` | произвольный текст |
| `DriverError` | `not implemented`, `driver not connected`, `no results returned` |

---

## 9. REST license API (не в cmd/server)

Код: `internal/rest/c_license`. **Не** монтируется в `cmd/server` — только для справки.

| HTTP | Условие | Тело |
|------|---------|------|
| 400 | Нет query `feature` | plain: `feature parameter required` |
| 403 | `ValidateFeature` fail | `{"error":"feature_not_available","message":"feature missing"}` |
| 200 | OK | пустое тело |

---

## 10. license-tool (CLI)

| Команда | Код выхода | Типичная причина |
|---------|------------|------------------|
| `create` | 1 | Нет private key, неверная дата, ошибка генерации |
| `verify` | 1 | Нет license file |
| `export-hwid` | 1 | Ошибка чтения железа |

Сообщения в stderr, не JSON-RPC.

---

## 11. Предупреждения при старте (сервер работает)

| Лог | Причина |
|-----|---------|
| `MQTT disabled` | Не подключился к `mqtt.broker_url` |
| `metrics disabled` | Не создался файл метрик |

MCP `write_tag` / `read_tag` при этом могут работать (in-memory). `subscribe_tag` без MQTT — риск паники.

---

## 12. Матрица: симптом → что проверить

| Симптом | Проверить |
|---------|-----------|
| Connection refused | Сервер запущен, порт 8080, firewall |
| Процесс сразу выходит | `license.dat`, `public.pem`, лог Fatalf |
| HTTP 400 plain text | `Content-Type: application/json` |
| `-32001` | `license.dat`, срок, HWID, подпись |
| `-32002` | x402 в config или добавить `x402_payment` |
| `-32601` | Имя метода: только `read_tag`, `write_tag`, `subscribe_tag` |
| `-32000` tag not found | Сначала `write_tag` |
| `-32000` connection refused (subscribe) | MQTT broker, `mqtt.broker_url` |
| `/health` ok, MCP fail | Лицензия/x402 только на POST `/` |

---

## 13. Ссылки на код

| Компонент | Файл |
|-----------|------|
| HTTP + JSON-RPC | `internal/interfaces/mcp/middleware.go`, `server.go` |
| Старт + лицензия | `cmd/server/main.go` |
| Лицензия | `pkg/license/validator.go` |
| x402 | `pkg/x402/handler.go` |
| write/read tag | `internal/application/command/write_tag.go`, `query/read_tag.go` |

Диагностика по шагам: [Diagnostic.md](./Diagnostic.md).
