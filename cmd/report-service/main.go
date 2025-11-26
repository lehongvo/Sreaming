package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"

	"ecommerce-platform/internal/config"
	"ecommerce-platform/internal/kafka"
	"ecommerce-platform/internal/models"
)

func main() {
	cfg := config.Load()
	cfg.ServiceName = "report-service"

	consumer := kafka.NewConsumer(cfg.KafkaBroker, "orders", "report-service-group")
	defer consumer.Close()

	router := mux.NewRouter()
	router.HandleFunc("/reports/sales", getSalesReport).Methods("GET")
	router.HandleFunc("/reports/orders", getOrdersReport).Methods("GET")
	router.HandleFunc("/health", healthCheck).Methods("GET")

	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	stats := make(map[string]int)

	go processReports(ctx, consumer, stats)

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

func processReports(ctx context.Context, consumer *kafka.Consumer, stats map[string]int) {
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
				continue
			}

			var order models.Order
			if err := json.Unmarshal(msg.Value, &order); err != nil {
				log.Printf("[report-service] Unmarshal error: %v", err)
				continue
			}

			stats[order.Platform]++
			log.Printf("[report-service] Updated stats for %s: %d orders", order.Platform, stats[order.Platform])
		}
	}
}

func getSalesReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Sales report endpoint"})
}

func getOrdersReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Orders report endpoint"})
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
