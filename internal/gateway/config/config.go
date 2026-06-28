package config

import (
	"os"
	"strconv"
)

type Config struct {
	REST  REST
	Admin Admin
}

type REST struct {
	Port int
}

type Admin struct {
	Port int
}

func MustLoad() Config {
	var cfg Config

	cfg.REST.Port = getEnvInt("REST_PORT", 8080)
	cfg.Admin.Port = getEnvInt("ADMIN_PANEL_PORT", 6060)

	return cfg
}

func getEnvInt(key string, fallback int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return v
	}

	return fallback
}
