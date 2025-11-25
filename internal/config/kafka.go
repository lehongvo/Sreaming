package config

import "os"

const (
	defaultBroker = "localhost:9092"
	defaultTopic  = "orders"
)

type Kafka struct {
	Broker string
	Topic  string
}

func LoadKafka() Kafka {
	cfg := Kafka{
		Broker: os.Getenv("KAFKA_BROKER_ADDR"),
		Topic:  os.Getenv("KAFKA_ORDER_TOPIC"),
	}

	if cfg.Broker == "" {
		cfg.Broker = defaultBroker
	}
	if cfg.Topic == "" {
		cfg.Topic = defaultTopic
	}

	return cfg
}
