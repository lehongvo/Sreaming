package config

import "os"

type Config struct {
	KafkaBroker string
	KafkaTopic  string
	HTTPPort    string
	ServiceName string
}

func Load() Config {
	return Config{
		KafkaBroker: getEnv("KAFKA_BROKER", "localhost:9092"),
		KafkaTopic:  getEnv("KAFKA_TOPIC", "webhooks"),
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		ServiceName: getEnv("SERVICE_NAME", "unknown"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
