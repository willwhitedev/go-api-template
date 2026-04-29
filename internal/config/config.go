package config

import "os"

type Config struct {
	Addr string
}

func Load() Config {
	return Config{
		Addr: getEnv("ADDR", ":8080"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
