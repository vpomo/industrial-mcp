FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates gcc musl-dev

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
COPY pkg /app/pkg

ENV GIN_MODE=release
EXPOSE 8080

CMD ["/app/server"]