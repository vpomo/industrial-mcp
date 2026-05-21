FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata wget

WORKDIR /app

COPY --from=builder /server /app/server
COPY cmd/server/config.yaml /app/config.yaml
COPY cmd/server/license.dat /app/license.dat
COPY cmd/server/key/public.pem /app/key/public.pem

ENV GIN_MODE=release
EXPOSE 8080

CMD ["/app/server", "-config", "/app/config.yaml"]
