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
	cfg.ServiceName = "webhooks-api"

	producer := kafka.NewProducer(cfg.KafkaBroker, cfg.KafkaTopic)
	defer producer.Close()

	router := mux.NewRouter()
	router.HandleFunc("/webhooks/{platform}", handleWebhook(producer)).Methods("POST")
	router.HandleFunc("/health", healthCheck).Methods("GET")

	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

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

func handleWebhook(producer *kafka.Producer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		platform := vars["platform"]

		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		event := models.WebhookEvent{
			ID:         generateID(),
			Platform:   platform,
			EventType:  getEventType(payload),
			Payload:    payload,
			ReceivedAt: time.Now(),
		}

		if err := producer.Send(r.Context(), event.ID, event); err != nil {
			http.Error(w, "Failed to process webhook", http.StatusInternalServerError)
			return
		}

		log.Printf("[webhooks-api] Received webhook from %s: %s", platform, event.ID)
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"status": "accepted", "id": event.ID})
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func generateID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

func getEventType(payload map[string]interface{}) string {
	if eventType, ok := payload["event_type"].(string); ok {
		return eventType
	}
	if eventType, ok := payload["type"].(string); ok {
		return eventType
	}
	return "unknown"
}
