package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	REST      REST
	Streaming Streaming
}

type REST struct {
	Port int
}

// Streaming addresses the streaming service's gRPC endpoint that the gateway
// forwards WHIP offers to.
type Streaming struct {
	Host string
	Port int
}

// Addr returns the host:port the gateway dials to reach the streaming service.
func (s Streaming) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func MustLoad() Config {
	var cfg Config

	cfg.REST.Port = getEnvInt("REST_PORT", 8080)
	cfg.Streaming.Host = getEnv("STREAMING_GRPC_HOST", "127.0.0.1")
	cfg.Streaming.Port = getEnvInt("STREAMING_GRPC_PORT", 9090)

	return cfg
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}

	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return v
	}

	return fallback
}
