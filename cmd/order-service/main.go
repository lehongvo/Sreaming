package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"ecommerce-platform/internal/config"
	"ecommerce-platform/internal/kafka"
	"ecommerce-platform/internal/models"
)

func main() {
	cfg := config.Load()
	cfg.ServiceName = "order-service"

	consumer := kafka.NewConsumer(cfg.KafkaBroker, cfg.KafkaTopic+"-enriched", "order-service-group")
	defer consumer.Close()

	producer := kafka.NewProducer(cfg.KafkaBroker, "orders")
	defer producer.Close()

	router := mux.NewRouter()
	router.HandleFunc("/orders", listOrders).Methods("GET")
	router.HandleFunc("/orders/{id}", getOrder).Methods("GET")
	router.HandleFunc("/health", healthCheck).Methods("GET")

	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	orders := make(map[string]models.Order)

	go processOrders(ctx, consumer, producer, orders)

	go func() {
		log.Printf("[%s] Starting server on port %s", cfg.ServiceName, cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-ctx.Done()
	log.Printf("[%s] Shutting down...", cfg.ServiceName)
	server.Shutdown(context.Background())
}

func processOrders(ctx context.Context, consumer *kafka.Consumer, producer *kafka.Producer, orders map[string]models.Order) {
	for {
		select {
		case <-ctx.Done():
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

			var enriched models.EnrichedEvent
			if err := json.Unmarshal(msg.Value, &enriched); err != nil {
				log.Printf("[order-service] Unmarshal error: %v", err)
				continue
			}

			if enriched.EventType == "order.created" || enriched.EventType == "order.updated" {
				order := convertToOrder(enriched)
				orders[order.ID] = order

				if err := producer.Send(ctx, order.ID, order); err != nil {
					log.Printf("[order-service] Failed to send order: %v", err)
					continue
				}

				log.Printf("[order-service] Processed order: %s from %s", order.ID, order.Platform)
			}
		}
	}
}

func convertToOrder(enriched models.EnrichedEvent) models.Order {
	order := models.Order{
		ID:        enriched.ID,
		Platform:  enriched.Platform,
		CreatedAt: enriched.ReceivedAt,
		Status:    "pending",
	}

	if payload, ok := enriched.Payload["order"].(map[string]interface{}); ok {
		if id, ok := payload["id"].(string); ok {
			order.ID = id
		}
		if total, ok := payload["total"].(float64); ok {
			order.Total = total
		}
		if status, ok := payload["status"].(string); ok {
			order.Status = status
		}
	}

	return order
}

func listOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Orders list endpoint"})
}

func getOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Get order endpoint"})
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

