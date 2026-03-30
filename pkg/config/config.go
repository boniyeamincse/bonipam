package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// ServiceConfig stores common runtime settings shared across services.
type ServiceConfig struct {
	ServiceName     string
	Env             string
	HTTPPort        int
	ShutdownTimeout time.Duration
}

func Load(serviceName string, defaultPort int) (ServiceConfig, error) {
	port, err := getenvInt("HTTP_PORT", defaultPort)
	if err != nil {
		return ServiceConfig{}, fmt.Errorf("invalid HTTP_PORT: %w", err)
	}

	shutdownSeconds, err := getenvInt("SHUTDOWN_TIMEOUT_SECONDS", 10)
	if err != nil {
		return ServiceConfig{}, fmt.Errorf("invalid SHUTDOWN_TIMEOUT_SECONDS: %w", err)
	}

	cfg := ServiceConfig{
		ServiceName:     serviceName,
		Env:             getenv("APP_ENV", "development"),
		HTTPPort:        port,
		ShutdownTimeout: time.Duration(shutdownSeconds) * time.Second,
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getenvInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return result, nil
}
