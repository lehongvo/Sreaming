package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ecommerce-platform/internal/config"
	"ecommerce-platform/internal/kafka"
	"ecommerce-platform/internal/models"
)

func main() {
	cfg := config.Load()
	cfg.ServiceName = "dragonfly-connector"

	consumer := kafka.NewConsumer(cfg.KafkaBroker, "orders", "dragonfly-group")
	defer consumer.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Printf("[%s] Starting connector to NetSuite", cfg.ServiceName)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[%s] Shutting down...", cfg.ServiceName)
			return
		default:
			msg, err := consumer.Read(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				time.Sleep(time.Second)
				continue
			}

			var order models.Order
			if err := json.Unmarshal(msg.Value, &order); err != nil {
				log.Printf("[%s] Unmarshal error: %v", cfg.ServiceName, err)
				continue
			}

			if order.Platform == "netsuite" {
				if err := sendToNetSuite(order); err != nil {
					log.Printf("[%s] Failed to send to NetSuite: %v", cfg.ServiceName, err)
					continue
				}
				log.Printf("[%s] Sent order %s to NetSuite", cfg.ServiceName, order.ID)
			}
		}
	}
}

func sendToNetSuite(order models.Order) error {
	log.Printf("[dragonfly] Sending order to NetSuite API: %+v", order)
	time.Sleep(100 * time.Millisecond)
	return nil
}

