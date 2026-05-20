# Industrial MCP Server - Руководство пользователя

## Обзор

Industrial MCP — сервер для работы с промышленными данными через MQTT и OPC UA протоколы. Использует JSON-RPC API для взаимодействия.

## Запуск

```bash
# Docker
docker-compose up -d

# Или напрямую
go run ./cmd/server
```

## Конфигурация

Настройки в `configs/config.yaml`:

| Параметр | Описание |
|----------|----------|
| `mqtt.broker_url` | MQTT брокер (tcp://localhost:1883) |
| `opcua.endpoint` | OPC UA сервер (opc.tcp://localhost:4840) |
| `license.enabled` | Включить проверку лицензии |
| `x402.enabled` | Включить x402 платежи |
| `metrics.file` | Путь к файлу логов запросов |

## JSON-RPC API

### Health check
```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

### Запись тега (write_tag)
```bash
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "write_tag",
    "params": {"tag_name": "temperature", "value": 25.5},
    "id": 1
  }'
```

### Чтение тега (read_tag)
```bash
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "read_tag",
    "params": {"tag_name": "temperature"},
    "id": 2
  }'
```

### Подписка на тег (subscribe_tag)
```bash
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "subscribe_tag",
    "params": {"tag_name": "temperature", "qos": 0},
    "id": 3
  }'
```

### Список всех тегов (list_tags)
```bash
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "list_tags",
    "params": {"limit": 100},
    "id": 4
  }'
```

## Ограничения

**Сервер не имеет REST API для:**
- Добавления/настройки внешних источников данных
- Задания тегов через конфигурацию
- Сканирования источников данных

Теги создаются программно в коде. Для расширения функциональности требуется доработка сервера.