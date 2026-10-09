package config

import "os"

type Config struct {
	AppEnv   string
	HTTPPort string
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

func Load() Config {
	return Config{
		AppEnv:   getEnv("APP_ENV", "Development"),
		HTTPPort: getEnv("HTTP_PORT", "8082"),
	}
}
