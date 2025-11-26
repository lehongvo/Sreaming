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
	cfg.ServiceName = "catalog-service"

	consumer := kafka.NewConsumer(cfg.KafkaBroker, cfg.KafkaTopic+"-enriched", "catalog-service-group")
	defer consumer.Close()

	router := mux.NewRouter()
	router.HandleFunc("/products", listProducts).Methods("GET")
	router.HandleFunc("/products/{id}", getProduct).Methods("GET")
	router.HandleFunc("/health", healthCheck).Methods("GET")

	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	products := make(map[string]models.CatalogProduct)

	go processCatalog(ctx, consumer, products)

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

func processCatalog(ctx context.Context, consumer *kafka.Consumer, products map[string]models.CatalogProduct) {
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
				log.Printf("[catalog-service] Unmarshal error: %v", err)
				continue
			}

			if enriched.EventType == "product.created" || enriched.EventType == "product.updated" {
				product := convertToProduct(enriched)
				products[product.ID] = product
				log.Printf("[catalog-service] Processed product: %s from %s", product.ID, enriched.Platform)
			}
		}
	}
}

func convertToProduct(enriched models.EnrichedEvent) models.CatalogProduct {
	product := models.CatalogProduct{
		ID:        enriched.ID,
		UpdatedAt: time.Now(),
		PlatformIDs: make(map[string]string),
	}

	if payload, ok := enriched.Payload["product"].(map[string]interface{}); ok {
		if name, ok := payload["name"].(string); ok {
			product.Name = name
		}
		if sku, ok := payload["sku"].(string); ok {
			product.SKU = sku
		}
		if price, ok := payload["price"].(float64); ok {
			product.Price = price
		}
	}

	product.PlatformIDs[enriched.Platform] = enriched.ID
	return product
}

func listProducts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Products list endpoint"})
}

func getProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Get product endpoint"})
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

