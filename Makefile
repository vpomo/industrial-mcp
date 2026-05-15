.PHONY: build run test lint clean docker-build docker-run docker-stop

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
	docker build -t mcp_mqtt_opcua:latest .

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down