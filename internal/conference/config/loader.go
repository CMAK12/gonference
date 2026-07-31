package config

import (
	"os"
	"strconv"
)

const (
	DriverPostgres  = "postgres"
	DriverDragonfly = "dragonfly"
)

func MustLoad() Config {
	var cfg Config

	cfg.GRPC.Address = getEnv("GRPC_ADDRESS", "0.0.0.0:9090")

	cfg.Storage.Driver = getEnv("STORAGE_DRIVER", DriverPostgres)

	cfg.Postgres.Host = getEnv("PG_HOST", "127.0.0.1")
	cfg.Postgres.Port = getEnv("PG_PORT", "5432")
	cfg.Postgres.User = getEnv("PG_USER", "postgres")
	cfg.Postgres.Password = getEnv("PG_PASSWORD", "")
	cfg.Postgres.Database = getEnv("PG_DATABASE", "gonference")
	cfg.Postgres.SSLMode = getEnv("PG_SSLMODE", "disable")

	cfg.Dragonfly.Addr = getEnv("DRAGONFLY_ADDR", "127.0.0.1:6379")
	cfg.Dragonfly.Password = getEnv("DRAGONFLY_PASSWORD", "")
	cfg.Dragonfly.DB = getEnvInt("DRAGONFLY_DB", 0)

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
