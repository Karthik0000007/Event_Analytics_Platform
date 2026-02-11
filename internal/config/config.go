package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ServiceName string
	Port        string
	LogLevel    string

	KafkaBrokers  []string
	KafkaTopic    string
	KafkaGroupID  string
	KafkaDLQTopic string // Dead-letter topic for failed messages

	DatabaseDSN string

	// Retry discipline
	MaxRetries int
}

func Load() *Config {
	return &Config{
		ServiceName:   getEnv("SERVICE_NAME", "ingestion-api"),
		Port:          getEnv("PORT", "8080"),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		KafkaBrokers:  strings.Split(getEnv("KAFKA_BROKERS", "localhost:9093"), ","),
		KafkaTopic:    getEnv("KAFKA_TOPIC", "events"),
		KafkaGroupID:  getEnv("KAFKA_GROUP_ID", "event-consumer-group"),
		KafkaDLQTopic: getEnv("KAFKA_DLQ_TOPIC", "events.dlq"),
		MaxRetries:    getEnvInt("MAX_RETRIES", 5),
		DatabaseDSN: fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			getEnv("DB_USER", "events_user"),
			getEnv("DB_PASSWORD", "events_password"),
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "5432"),
			getEnv("DB_NAME", "events_db"),
		),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
