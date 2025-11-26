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
)

func main() {
	cfg := config.Load()
	cfg.ServiceName = "settings-service"

	router := mux.NewRouter()
	router.HandleFunc("/settings", getSettings).Methods("GET")
	router.HandleFunc("/settings", updateSettings).Methods("PUT")
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

func getSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Get settings endpoint"})
}

func updateSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Update settings endpoint"})
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

