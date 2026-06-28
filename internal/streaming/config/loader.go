package config

import (
	"os"
	"strconv"
)

func MustLoad() Config {
	var cfg Config

	cfg.WS.Port = getEnvInt("WS_PORT", 8082)
	cfg.GRPC.Port = getEnvInt("GRPC_PORT", 9090)

	return cfg
}

func getEnvInt(key string, fallback int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return v
	}

	return fallback
}
