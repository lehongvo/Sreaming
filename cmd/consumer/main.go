package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/segmentio/kafka-go"

	"streaming-orders/internal/config"
	"streaming-orders/internal/metrics"
	"streaming-orders/internal/orders"
)

const metricsAddr = ":2112"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go metrics.StartServer(ctx, metricsAddr)

	if err := run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("consumer error: %v", err)
	}
}

func run(ctx context.Context) error {
	cfg := config.LoadKafka()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{cfg.Broker},
		Topic:    cfg.Topic,
		GroupID:  "order-service",
		MaxBytes: 10e6,
	})
	defer reader.Close()

	log.Printf("Consumer subscribed to %s, topic=%s\n", cfg.Broker, cfg.Topic)

	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("Consumer stopped")
				return ctx.Err()
			}
			log.Printf("read error: %v\n", err)
			continue
		}

		var event orders.Event
		if err := json.Unmarshal(m.Value, &event); err != nil {
			log.Printf("unmarshal error: %v\n", err)
			metrics.RecordFailure("unknown")
			continue
		}

		if err := processOrder(event); err != nil {
			metrics.RecordFailure(string(event.Type))
			continue
		}

		metrics.RecordSuccess(event.Type)
	}
}

func processOrder(o orders.Event) error {
	switch o.Type {
	case orders.Buy:
		log.Printf("[BUY ] user=%s symbol=%s qty=%d price=%.2f id=%s\n",
			o.UserID, o.Symbol, o.Quantity, o.Price, o.OrderID)
	case orders.Sell:
		log.Printf("[SELL] user=%s symbol=%s qty=%d price=%.2f id=%s\n",
			o.UserID, o.Symbol, o.Quantity, o.Price, o.OrderID)
	default:
		log.Printf("[UNKNOWN] %+v\n", o)
		return errors.New("unknown order type")
	}

	return nil
}
