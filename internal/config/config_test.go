package config

import (
	"testing"
)

func TestLoad_defaults(t *testing.T) {
	cfg := Load()
	if cfg.Addr != ":8080" {
		t.Fatalf("Addr = %q, want :8080", cfg.Addr)
	}
}

func TestLoad_env(t *testing.T) {
	t.Setenv("ADDR", ":9090")
	cfg := Load()
	if cfg.Addr != ":9090" {
		t.Fatalf("Addr = %q, want :9090", cfg.Addr)
	}
}
