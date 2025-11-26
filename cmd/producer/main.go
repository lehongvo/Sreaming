package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"

	"streaming-orders/internal/config"
	"streaming-orders/internal/orders"
)

const workerCount = 10

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("producer error: %v", err)
	}
}

func run(ctx context.Context) error {
	cfg := config.LoadKafka()

	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Broker),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
	}
	defer writer.Close()

	log.Printf("Producer connected to %s, topic=%s\n", cfg.Broker, cfg.Topic)

	var idCounter atomic.Int64
	g, gctx := errgroup.WithContext(ctx)

	for i := 0; i < workerCount; i++ {
		workerID := i
		g.Go(func() error {
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))
			for {
				select {
				case <-gctx.Done():
					return gctx.Err()
				default:
				}

				id := idCounter.Add(1)
				orderType := orders.Buy
				if r.Intn(2) == 0 {
					orderType = orders.Sell
				}

				event := orders.Event{
					OrderID:   "ORD-" + strconv.FormatInt(id, 10),
					UserID:    "user-" + strconv.Itoa(r.Intn(5)+1),
					Symbol:    []string{"AAPL", "GOOG", "TSLA", "MSFT"}[r.Intn(4)],
					Type:      orderType,
					Quantity:  r.Intn(100) + 1,
					Price:     100 + r.Float64()*50,
					Timestamp: time.Now().UTC(),
				}

				value, err := json.Marshal(event)
				if err != nil {
					log.Printf("marshal error: %v\n", err)
					continue
				}

				msg := kafka.Message{
					Key:   []byte(event.OrderID),
					Value: value,
				}

				if err := writer.WriteMessages(gctx, msg); err != nil {
					log.Printf("write error: %v\n", err)
					continue
				}

				log.Printf("[worker-%d] sent order: %+v\n", workerID, event)
				// print
			}
		})
	}

	if err := g.Wait(); err != nil && err != context.Canceled {
		return err
	}

	return ctx.Err()
}
