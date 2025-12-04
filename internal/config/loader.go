package config

import (
	"os"
	"strconv"
)

func MustLoad() Config {
	var config Config

	config.REST.Port = getEnvInt("REST_PORT", 8080)

	config.AdminPanel.Port = getEnvInt("ADMIN_PANEL_PORT", 6060)

	return config
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func getEnvInt(key string, fallback int) int {
	v, err := strconv.Atoi(os.Getenv(key))
	if err == nil {
		return v
	}

	return fallback
}
