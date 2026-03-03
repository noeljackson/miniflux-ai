package main

import (
	"fmt"
	"os"
)

type Config struct {
	MinifluxURL     string
	MinifluxAPIKey  string
	AnthropicAPIKey string
	WebhookSecret   string
	Port            string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		MinifluxURL:     os.Getenv("MINIFLUX_URL"),
		MinifluxAPIKey:  os.Getenv("MINIFLUX_API_KEY"),
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
		WebhookSecret:   os.Getenv("WEBHOOK_SECRET"),
		Port:            os.Getenv("PORT"),
	}

	if cfg.MinifluxURL == "" {
		return nil, fmt.Errorf("MINIFLUX_URL is required")
	}
	if cfg.MinifluxAPIKey == "" {
		return nil, fmt.Errorf("MINIFLUX_API_KEY is required")
	}
	if cfg.AnthropicAPIKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is required")
	}
	if cfg.WebhookSecret == "" {
		return nil, fmt.Errorf("WEBHOOK_SECRET is required")
	}
	if cfg.Port == "" {
		cfg.Port = "3000"
	}

	return cfg, nil
}
