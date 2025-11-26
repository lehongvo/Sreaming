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
	cfg.ServiceName = "mantis-connector"

	consumer := kafka.NewConsumer(cfg.KafkaBroker, "orders", "mantis-group")
	defer consumer.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Printf("[%s] Starting connector to BigCommerce", cfg.ServiceName)

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

			if order.Platform == "bigcommerce" {
				if err := sendToBigCommerce(order); err != nil {
					log.Printf("[%s] Failed to send to BigCommerce: %v", cfg.ServiceName, err)
					continue
				}
				log.Printf("[%s] Sent order %s to BigCommerce", cfg.ServiceName, order.ID)
			}
		}
	}
}

func sendToBigCommerce(order models.Order) error {
	log.Printf("[mantis] Sending order to BigCommerce API: %+v", order)
	time.Sleep(100 * time.Millisecond)
	return nil
}

