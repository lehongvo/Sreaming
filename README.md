# Streaming Orders Demo

Small Kafka streaming playground written in Go. Contains a producer that emits random BUY/SELL orders and a consumer that processes them.

## Project Layout

- `cmd/producer`: standalone binary that pushes `orders.Event` messages to Kafka.
- `cmd/consumer`: standalone binary that reads messages and logs them. This is where you can add further business logic.
- `internal/orders`: shared domain model (`Event`, `Type` constants).
- `internal/config`: helper to load Kafka host/topic from environment variables.
- `Dockerfile` + `docker-compose.yml`: run Kafka, Zookeeper, producer, and consumer together.

## Run Locally (without Docker)

```bash
export KAFKA_BROKER_ADDR=localhost:9092
export KAFKA_ORDER_TOPIC=orders

go run ./cmd/consumer   # in terminal 1
go run ./cmd/producer   # in terminal 2
```

Before running, make sure you already have Kafka + Zookeeper up (outside of this repo) and a topic named `orders`.

## Run Everything with Docker Compose

```bash
docker compose build
docker compose up
```

This will start Zookeeper, Kafka, producer, consumer, plus the monitoring stack:

- `kafka-exporter` exposes Kafka broker metrics at `http://localhost:9308/metrics`.
- `prometheus` scrapes the exporter (UI available at `http://localhost:9090`).
- `grafana` auto-loads Prometheus as a datasource plus a dashboard named **Kafka Orders Overview** (login `admin` / `admin` on `http://localhost:3000`).

Stop the whole stack with:

```bash
docker compose down
```


