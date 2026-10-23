package config

import (
	"os"
	"strconv"
	"time"

	infraserver "github.com/CMAK12/gonference/infra/server/grpc"
)

func MustLoad() Config {
	var cfg Config

	cfg.GRPC.Addr = getEnv("GRPC_ADDR", ":10003")

	cfg.GRPC.MaxRecvMsgSize = getEnvInt("GRPC_MAX_RECV_MSG_SIZE", infraserver.DefaultMaxRecvMsgSize)
	cfg.GRPC.MaxSendMsgSize = getEnvInt("GRPC_MAX_SEND_MSG_SIZE", infraserver.DefaultMaxSendMsgSize)

	cfg.GRPC.ConnectionTimeout = getEnvDuration("GRPC_CONNECTION_TIMEOUT", infraserver.DefaultConnectionTimeout)
	cfg.GRPC.ShutdownTimeout = getEnvDuration("GRPC_SHUTDOWN_TIMEOUT", infraserver.DefaultShutdownTimeout)

	cfg.GRPC.MaxConnectionIdle = getEnvDuration("GRPC_MAX_CONNECTION_IDLE", infraserver.DefaultMaxConnectionIdle)
	cfg.GRPC.KeepaliveTime = getEnvDuration("GRPC_KEEPALIVE_TIME", infraserver.DefaultKeepaliveTime)
	cfg.GRPC.KeepaliveTimeout = getEnvDuration("GRPC_KEEPALIVE_TIMEOUT", infraserver.DefaultKeepaliveTimeout)
	cfg.GRPC.MinClientPingInterval = getEnvDuration("GRPC_MIN_CLIENT_PING_INTERVAL", infraserver.DefaultMinClientPingInterval)

	cfg.GRPC.Reflection = getEnvBool("GRPC_REFLECTION", false)

	cfg.Relay.AdvertiseAddr = getEnv("RELAY_ADVERTISE_ADDR", "localhost:10003")

	cfg.Signaling.MaxResolveRetries = getEnvInt("SIGNALING_MAX_RESOLVE_RETRIES", 2)
	cfg.Signaling.RelayOwnershipTTL = getEnvDuration("SIGNALING_RELAY_OWNERSHIP_TTL", 2*time.Minute)

	cfg.InMemoryDB.Addr = getEnv("DRAGONFLY_ADDR", "127.0.0.1:6379")
	cfg.InMemoryDB.Password = getEnv("DRAGONFLY_PASSWORD", "")
	cfg.InMemoryDB.Database = getEnvInt("DRAGONFLY_DB", 0)

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

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v, err := time.ParseDuration(os.Getenv(key)); err == nil {
		return v
	}

	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v, err := strconv.ParseBool(os.Getenv(key)); err == nil {
		return v
	}

	return fallback
}
