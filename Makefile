.PHONY: build run test lint clean docker-build docker-init-identity docker-run docker-hwid docker-hwid-offline docker-stop

DOCKER_HOSTNAME ?= mcp-server
DOCKER_MAC ?= 02:42:ac:14:00:01

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

docker-init-identity:
	bash scripts/docker-init-identity.sh

docker-run: docker-init-identity
	docker-compose up -d

docker-hwid:
	curl -s http://localhost:8080/api/v1/license/hwid | python3 -m json.tool

docker-hwid-offline: docker-init-identity
	docker run --rm \
	  --hostname $(DOCKER_HOSTNAME) \
	  --mac-address $(DOCKER_MAC) \
	  -v "$(CURDIR)/docker/identity/machine-id:/etc/machine-id:ro" \
	  -v "$(CURDIR):/src" -w /src \
	  golang:1.26-alpine \
	  sh -c "apk add --no-cache iproute2 >/dev/null && go run ./cmd/license-tool export-hwid"

docker-stop:
	docker-compose down