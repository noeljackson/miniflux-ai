package main

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	MinifluxURL        string
	MinifluxAPIKey     string
	AnthropicAPIKey    string
	WebhookSecret      string
	RelevanceThreshold int
	Port               string
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

	threshold := os.Getenv("RELEVANCE_THRESHOLD")
	if threshold == "" {
		cfg.RelevanceThreshold = 75
	} else {
		t, err := strconv.Atoi(threshold)
		if err != nil {
			return nil, fmt.Errorf("RELEVANCE_THRESHOLD must be an integer: %w", err)
		}
		if t < 0 || t > 100 {
			return nil, fmt.Errorf("RELEVANCE_THRESHOLD must be 0-100, got %d", t)
		}
		cfg.RelevanceThreshold = t
	}

	return cfg, nil
}
