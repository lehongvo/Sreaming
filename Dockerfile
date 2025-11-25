FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/producer ./cmd/producer && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/consumer ./cmd/consumer

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/producer /app/producer
COPY --from=builder /app/consumer /app/consumer

ENV KAFKA_BROKER_ADDR=kafka:9092
ENV KAFKA_ORDER_TOPIC=orders

CMD ["/app/producer"]


