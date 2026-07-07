package config

import (
	"os"
	"strconv"
)

func MustLoad() Config {
	var cfg Config

	cfg.GRPC.Address = getEnv("GRPC_ADDRESS", "0.0.0.0:9090")

	return cfg
}

func getEnv(key string, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return v
	}

	return fallback
}
