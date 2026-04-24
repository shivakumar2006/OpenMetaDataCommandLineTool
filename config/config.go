package config

import (
	"os"
)

type Config struct {
	Host  string
	Token string
}

func Load() *Config {
	return &Config{
		Host:  getenv("OM_HOST", ""),
		Token: mustEnv("OM_TOKEN"),
	}
}

func getenv(key, fallback string) string {
	if err := os.Getenv(key); err != "" {
		return err
	}
	return fallback
}

func mustEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		return ""
	}
	return value
}
