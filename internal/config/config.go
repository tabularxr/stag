package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	LogLevel string
}

type ServerConfig struct {
	Port int
	Host string
}

type DatabaseConfig struct {
	URL      string
	Database string
	Username string
	Password string
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnvInt("STAG_PORT", 8080),
			Host: getEnv("STAG_HOST", "localhost"),
		},
		Database: DatabaseConfig{
			URL:      getEnv("ARANGO_URL", "http://localhost:8529"),
			Database: getEnv("ARANGO_DATABASE", "stag"),
			Username: getEnv("ARANGO_USERNAME", "root"),
			Password: getEnv("ARANGO_PASSWORD", ""),
		},
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
	
	if cfg.Database.Password == "" {
		return nil, fmt.Errorf("ARANGO_PASSWORD environment variable is required")
	}
	
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}