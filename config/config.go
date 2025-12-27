package config

import (
	"flag"
	"os"
)

// Config holds all configuration for the proxy server
type Config struct {
	Port        string
	MaxClients  int
	CacheSize   int
	LogLevel    string
}

// LoadConfig loads configuration from command line flags and environment variables
func LoadConfig() *Config {
	cfg := &Config{
		Port:       getEnvOrDefault("PROXY_PORT", "9000"),
		MaxClients: 100,
		CacheSize:  100,
		LogLevel:   getEnvOrDefault("LOG_LEVEL", "info"),
	}

	flag.StringVar(&cfg.Port, "port", cfg.Port, "Port to listen on (default: 9000)")
	flag.IntVar(&cfg.MaxClients, "max-clients", cfg.MaxClients, "Maximum concurrent client connections (default: 100)")
	flag.IntVar(&cfg.CacheSize, "cache-size", cfg.CacheSize, "LRU cache capacity (default: 100)")
	flag.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "Log level: info, debug, error (default: info)")
	flag.Parse()

	return cfg
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

