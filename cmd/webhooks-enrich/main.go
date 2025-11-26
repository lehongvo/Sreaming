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
	cfg.ServiceName = "webhooks-enrich"

	consumer := kafka.NewConsumer(cfg.KafkaBroker, cfg.KafkaTopic, "enrich-group")
	defer consumer.Close()

	producer := kafka.NewProducer(cfg.KafkaBroker, cfg.KafkaTopic+"-enriched")
	defer producer.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Printf("[%s] Starting enrichment service", cfg.ServiceName)

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
				log.Printf("[%s] Read error: %v", cfg.ServiceName, err)
				time.Sleep(time.Second)
				continue
			}

			var event models.WebhookEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("[%s] Unmarshal error: %v", cfg.ServiceName, err)
				continue
			}

			enriched := enrichEvent(event)
			if err := producer.Send(ctx, enriched.ID, enriched); err != nil {
				log.Printf("[%s] Failed to send enriched event: %v", cfg.ServiceName, err)
				continue
			}

			log.Printf("[%s] Enriched event: %s from %s", cfg.ServiceName, enriched.ID, enriched.Platform)
		}
	}
}

func enrichEvent(event models.WebhookEvent) models.EnrichedEvent {
	now := time.Now()
	event.ProcessedAt = &now

	enrichedData := make(map[string]interface{})
	enrichedData["source"] = event.Platform
	enrichedData["enriched_by"] = "ant-enrich"
	enrichedData["timestamp"] = now.Unix()

	// Add platform-specific enrichment
	switch event.Platform {
	case "shopify":
		enrichedData["store_id"] = extractStoreID(event.Payload)
	case "magento":
		enrichedData["website_id"] = extractWebsiteID(event.Payload)
	case "bigcommerce":
		enrichedData["store_hash"] = extractStoreHash(event.Payload)
	}

	return models.EnrichedEvent{
		WebhookEvent: event,
		EnrichedData: enrichedData,
		EnrichedAt:   now,
	}
}

func extractStoreID(payload map[string]interface{}) string {
	if shop, ok := payload["shop_domain"].(string); ok {
		return shop
	}
	return "unknown"
}

func extractWebsiteID(payload map[string]interface{}) string {
	if id, ok := payload["website_id"].(string); ok {
		return id
	}
	return "unknown"
}

func extractStoreHash(payload map[string]interface{}) string {
	if hash, ok := payload["store_hash"].(string); ok {
		return hash
	}
	return "unknown"
}

