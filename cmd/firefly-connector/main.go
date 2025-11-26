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
	cfg.ServiceName = "firefly-connector"

	consumer := kafka.NewConsumer(cfg.KafkaBroker, "orders", "firefly-group")
	defer consumer.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Printf("[%s] Starting connector to Core (MSI)", cfg.ServiceName)

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

			if order.Platform == "msi" || shouldRouteToMSI(order) {
				if err := sendToMSI(order); err != nil {
					log.Printf("[%s] Failed to send to MSI: %v", cfg.ServiceName, err)
					continue
				}
				log.Printf("[%s] Sent order %s to Core (MSI)", cfg.ServiceName, order.ID)
			}
		}
	}
}

func shouldRouteToMSI(order models.Order) bool {
	// Routing logic for MSI
	return false
}

func sendToMSI(order models.Order) error {
	// Simulate API call to MSI
	log.Printf("[firefly] Sending order to MSI API: %+v", order)
	time.Sleep(100 * time.Millisecond)
	return nil
}
