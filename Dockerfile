FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build all services
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/webhooks-api ./cmd/webhooks-api && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/webhooks-enrich ./cmd/webhooks-enrich && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/order-service ./cmd/order-service && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/catalog-service ./cmd/catalog-service && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/settings-service ./cmd/settings-service && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/report-service ./cmd/report-service && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/firefly-connector ./cmd/firefly-connector && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/mantis-connector ./cmd/mantis-connector && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/ladybug-connector ./cmd/ladybug-connector && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/hermes-connector ./cmd/hermes-connector && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/dragonfly-connector ./cmd/dragonfly-connector && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/locust-connector ./cmd/locust-connector

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/webhooks-api /app/webhooks-api
COPY --from=builder /app/webhooks-enrich /app/webhooks-enrich
COPY --from=builder /app/order-service /app/order-service
COPY --from=builder /app/catalog-service /app/catalog-service
COPY --from=builder /app/settings-service /app/settings-service
COPY --from=builder /app/report-service /app/report-service
COPY --from=builder /app/firefly-connector /app/firefly-connector
COPY --from=builder /app/mantis-connector /app/mantis-connector
COPY --from=builder /app/ladybug-connector /app/ladybug-connector
COPY --from=builder /app/hermes-connector /app/hermes-connector
COPY --from=builder /app/dragonfly-connector /app/dragonfly-connector
COPY --from=builder /app/locust-connector /app/locust-connector

CMD ["/app/webhooks-api"]

