package config

import (
	"os"
	"strconv"
	"time"

	infraserver "github.com/CMAK12/gonference/infra/server/grpc"
)

const (
	DriverPostgres  = "postgres"
	DriverDragonfly = "dragonfly"

	serviceSchema = "conference"
)

func MustLoad() Config {
	var cfg Config

	cfg.GRPC.Addr = getEnv("GRPC_ADDR", ":10002")

	cfg.GRPC.MaxRecvMsgSize = getEnvInt("GRPC_MAX_RECV_MSG_SIZE", infraserver.DefaultMaxRecvMsgSize)
	cfg.GRPC.MaxSendMsgSize = getEnvInt("GRPC_MAX_SEND_MSG_SIZE", infraserver.DefaultMaxSendMsgSize)

	cfg.GRPC.ConnectionTimeout = getEnvDuration("GRPC_CONNECTION_TIMEOUT", infraserver.DefaultConnectionTimeout)
	cfg.GRPC.ShutdownTimeout = getEnvDuration("GRPC_SHUTDOWN_TIMEOUT", infraserver.DefaultShutdownTimeout)

	cfg.GRPC.MaxConnectionIdle = getEnvDuration("GRPC_MAX_CONNECTION_IDLE", infraserver.DefaultMaxConnectionIdle)
	cfg.GRPC.KeepaliveTime = getEnvDuration("GRPC_KEEPALIVE_TIME", infraserver.DefaultKeepaliveTime)
	cfg.GRPC.KeepaliveTimeout = getEnvDuration("GRPC_KEEPALIVE_TIMEOUT", infraserver.DefaultKeepaliveTimeout)
	cfg.GRPC.MinClientPingInterval = getEnvDuration("GRPC_MIN_CLIENT_PING_INTERVAL", infraserver.DefaultMinClientPingInterval)

	cfg.GRPC.Reflection = getEnvBool("GRPC_REFLECTION", false)

	cfg.RelationalDB.Driver = getEnv("STORAGE_DRIVER", DriverPostgres)

	cfg.RelationalDB.Host = getEnv("PG_HOST", "127.0.0.1")
	cfg.RelationalDB.Port = getEnv("PG_PORT", "5432")
	cfg.RelationalDB.User = getEnv("PG_USER", DriverPostgres)
	cfg.RelationalDB.Password = getEnv("PG_PASSWORD", "")
	cfg.RelationalDB.Database = getEnv("PG_DATABASE", "gonference")
	cfg.RelationalDB.Schema = getEnv("PG_SCHEMA", serviceSchema)
	cfg.RelationalDB.SSLMode = getEnv("PG_SSLMODE", "disable")

	cfg.RelationalDB.MaxConnections = getEnvInt32("PG_MAX_CONNECTIONS", 0)
	cfg.RelationalDB.MinConnections = getEnvInt32("PG_MIN_CONNECTIONS", 0)
	cfg.RelationalDB.MaxConnLifetime = getEnvDuration("PG_MAX_CONN_LIFETIME", 0)
	cfg.RelationalDB.MaxConnIdleTime = getEnvDuration("PG_MAX_CONN_IDLE_TIME", 0)

	cfg.InMemoryDB.Driver = getEnv("CACHE_DRIVER", DriverDragonfly)

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

func getEnvInt32(key string, fallback int32) int32 {
	if v, err := strconv.ParseInt(os.Getenv(key), 10, 32); err == nil {
		return int32(v)
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
