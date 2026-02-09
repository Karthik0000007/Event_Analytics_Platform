package config

import (
	"os"
)

type Config struct {
	ServiceName string
	Port        string
	LogLevel    string
}

func Load() *Config {
	return &Config{
		ServiceName: getEnv("SERVICE_NAME", "ingestion-api"),
		Port:        getEnv("PORT", "8080"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
