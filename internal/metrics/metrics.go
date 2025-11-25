package metrics

import (
	"context"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"streaming-orders/internal/orders"
)

var (
	orderProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_processed_total",
			Help: "Count of processed orders labeled by type and result",
		},
		[]string{"type", "result"},
	)
)

func init() {
	prometheus.MustRegister(orderProcessed)
}

// StartServer runs a /metrics endpoint until ctx is cancelled.
func StartServer(ctx context.Context, addr string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()

	log.Printf("metrics exposed at %s/metrics\n", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("metrics server error: %v\n", err)
	}
}

func RecordSuccess(orderType orders.Type) {
	orderProcessed.WithLabelValues(string(orderType), "success").Inc()
}

func RecordFailure(orderType string) {
	orderProcessed.WithLabelValues(orderType, "failed").Inc()
}
