package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"go-api-template/internal/server"
)

func main() {
	addr := getEnv("ADDR", ":8080")

	api := &http.Server{
		Addr:              addr,
		Handler:           server.NewRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info("starting api server", "addr", addr)
	if err := api.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("api server stopped", "error", err)
		os.Exit(1)
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
