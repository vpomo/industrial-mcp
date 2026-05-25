# industrial-mcp

Использование MCP сервера
Запуск
# Docker
docker-compose up -d
make docker-run
# Или напрямую
go run ./cmd/server
API Endpoints
Health check:
curl http://localhost:8080/health
# {"status":"ok"}
JSON-RPC методы:
1. Запись тега (write_tag)
   curl -X POST http://localhost:8080 \
   -H "Content-Type: application/json" \
   -d '{
   "jsonrpc": "2.0",
   "method": "write_tag",
   "params": {"tag_name": "temperature", "value": 25.5},
   "id": 1
   }'
2. Чтение тега (read_tag)
   curl -X POST http://localhost:8080 \
   -H "Content-Type: application/json" \
   -d '{
   "jsonrpc": "2.0",
   "method": "read_tag",
   "params": {"tag_name": "temperature"},
   "id": 2
   }'
3. Подписка на тег (subscribe_tag)
   curl -X POST http://localhost:8080 \
   -H "Content-Type: application/json" \
   -d '{
   "jsonrpc": "2.0",
   "method": "subscribe_tag",
   "params": {"tag_name": "temperature", "qos": 0},
   "id": 3
   }'
4. Список всех тегов (list_tags)
   curl -X POST http://localhost:8080 \
   -H "Content-Type: application/json" \
   -d '{
   "jsonrpc": "2.0",
   "method": "list_tags",
   "params": {"limit": 100},
   "id": 4
   }'
   Конфигурация
   Настройки в configs/config.yaml или через .env:
   Параметр	Описание
   mqtt.broker_url	MQTT брокер (tcp://localhost:1883)
   opcua.endpoint	OPC UA сервер (opc.tcp://localhost:4840)
   license.enabled	Включить проверку лицензии
   x402.enabled	Включить x402 платежи
   Нагрузочные тесты
# Тесты
go test -v -cover ./...

#### Создание `license-tool` под Ubuntu и Windows
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o license-tool ./cmd/license-tool
CGO_ENABLED=0 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 -o license-tool ./cmd/license-tool
```


#### Для subscribe_tag и MQTT-уведомлений — поднять брокер, например Mosquitto:
```bash
docker run -d --name mosquitto -p 1883:1883 eclipse-mosquitto
```
#### В config.yaml:

```json
mqtt:
  broker_url: "tcp://localhost:1883"
```

Итог: брокер = транспорт сообщений, не источник данных. Источники — OPC UA, Modbus и т.д.
Брокер нужен, если:
- несколько систем должны узнавать об изменениях без опроса MCP
- нужен pub/sub между сервисами
Не нужен, если:
- один клиент пишет/читает через HTTP JSON-RPC (write_tag / read_tag)


Для subscribe_tag и MQTT-уведомлений — поднять брокер, например Mosquitto
```
docker run -d --name mosquitto -p 1883:1883 eclipse-mosquitto
```

