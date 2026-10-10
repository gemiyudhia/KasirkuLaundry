package config

import (
	"net"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func (c DBConfig) DSN() string {
	cfg := mysql.NewConfig()
	cfg.User = c.User
	cfg.Passwd = c.Password
	cfg.Net = "tcp"
	cfg.Addr = net.JoinHostPort(c.Host, c.Port)
	cfg.DBName = c.Name
	cfg.ParseTime = true
	cfg.Loc = time.Local
	return cfg.FormatDSN()
}

type Config struct {
	AppEnv   string
	HTTPPort string
	DB       DBConfig
}

func Load() Config {
	return Config{
		AppEnv:   getEnv("APP_ENV", "Development"),
		HTTPPort: getEnv("HTTP_PORT", "8082"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "127.0.0.1"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "kasirku_laundry_orders"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
