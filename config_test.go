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

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MinifluxURL != "https://rss.example.com" {
		t.Errorf("MinifluxURL = %q", cfg.MinifluxURL)
	}
	if cfg.Port != "3000" {
		t.Errorf("Port = %q, want 3000", cfg.Port)
	}
	if cfg.RelevanceThreshold != 75 {
		t.Errorf("RelevanceThreshold = %d, want 75", cfg.RelevanceThreshold)
	}
}

func TestLoadConfig_MissingRequired(t *testing.T) {
	tests := []struct {
		name    string
		setEnv  map[string]string
		wantErr string
	}{
		{"missing MINIFLUX_URL", map[string]string{"MINIFLUX_API_KEY": "k", "ANTHROPIC_API_KEY": "k", "WEBHOOK_SECRET": "s"}, "MINIFLUX_URL"},
		{"missing MINIFLUX_API_KEY", map[string]string{"MINIFLUX_URL": "u", "ANTHROPIC_API_KEY": "k", "WEBHOOK_SECRET": "s"}, "MINIFLUX_API_KEY"},
		{"missing ANTHROPIC_API_KEY", map[string]string{"MINIFLUX_URL": "u", "MINIFLUX_API_KEY": "k", "WEBHOOK_SECRET": "s"}, "ANTHROPIC_API_KEY"},
		{"missing WEBHOOK_SECRET", map[string]string{"MINIFLUX_URL": "u", "MINIFLUX_API_KEY": "k", "ANTHROPIC_API_KEY": "k"}, "WEBHOOK_SECRET"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.setEnv {
				t.Setenv(k, v)
			}
			_, err := LoadConfig()
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestLoadConfig_CustomThreshold(t *testing.T) {
	t.Setenv("MINIFLUX_URL", "u")
	t.Setenv("MINIFLUX_API_KEY", "k")
	t.Setenv("ANTHROPIC_API_KEY", "k")
	t.Setenv("WEBHOOK_SECRET", "s")
	t.Setenv("RELEVANCE_THRESHOLD", "90")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.RelevanceThreshold != 90 {
		t.Errorf("RelevanceThreshold = %d, want 90", cfg.RelevanceThreshold)
	}
}

func TestLoadConfig_InvalidThreshold(t *testing.T) {
	t.Setenv("MINIFLUX_URL", "u")
	t.Setenv("MINIFLUX_API_KEY", "k")
	t.Setenv("ANTHROPIC_API_KEY", "k")
	t.Setenv("WEBHOOK_SECRET", "s")
	t.Setenv("RELEVANCE_THRESHOLD", "abc")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for non-integer threshold")
	}
}
