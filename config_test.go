package main

import (
	"strings"
	"testing"
)

func TestLoadConfig_AllRequired(t *testing.T) {
	t.Setenv("MINIFLUX_URL", "https://rss.example.com")
	t.Setenv("MINIFLUX_API_KEY", "mf-key")
	t.Setenv("ANTHROPIC_API_KEY", "ant-key")
	t.Setenv("WEBHOOK_SECRET", "secret")
	t.Setenv("DATABASE_URL", "postgres://localhost/miniflux")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MinifluxURL != "https://rss.example.com" {
		t.Errorf("MinifluxURL = %q", cfg.MinifluxURL)
	}
	if cfg.DatabaseURL != "postgres://localhost/miniflux" {
		t.Errorf("DatabaseURL = %q", cfg.DatabaseURL)
	}
	if cfg.Port != "3000" {
		t.Errorf("Port = %q, want 3000", cfg.Port)
	}
}

func TestLoadConfig_MissingRequired(t *testing.T) {
	all := map[string]string{
		"MINIFLUX_URL":      "u",
		"MINIFLUX_API_KEY":  "k",
		"ANTHROPIC_API_KEY": "k",
		"WEBHOOK_SECRET":    "s",
		"DATABASE_URL":      "d",
	}

	for missing := range all {
		t.Run("missing "+missing, func(t *testing.T) {
			for k, v := range all {
				if k == missing {
					continue
				}
				t.Setenv(k, v)
			}
			_, err := LoadConfig()
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), missing) {
				t.Errorf("error = %q, want substring %q", err, missing)
			}
		})
	}
}
