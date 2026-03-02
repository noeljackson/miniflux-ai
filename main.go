package main

import (
	"log"
	"net/http"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	curator := NewClaudeCurator(cfg.AnthropicAPIKey)
	updater := NewMinifluxUpdater(cfg.MinifluxURL, cfg.MinifluxAPIKey)

	webhook := &WebhookHandler{
		Secret:    cfg.WebhookSecret,
		Threshold: cfg.RelevanceThreshold,
		Curator:   curator,
		Updater:   updater,
	}

	mux := http.NewServeMux()
	mux.Handle("POST /webhook", webhook)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	log.Printf("miniflux-ai listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
