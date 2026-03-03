package main

import (
	"context"
	"log"
	"net/http"

	miniflux "miniflux.app/v2/client"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx := context.Background()

	store, err := NewVectorStore(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("vectorstore error: %v", err)
	}
	defer store.Close()

	if err := store.Migrate(ctx); err != nil {
		log.Fatalf("migration error: %v", err)
	}
	log.Println("migration complete")

	embedder := NewOllamaEmbedder("http://ollama:11434")
	curator := NewClaudeCurator(cfg.AnthropicAPIKey)
	updater := NewMinifluxUpdater(cfg.MinifluxURL, cfg.MinifluxAPIKey)
	mfClient := miniflux.NewClient(cfg.MinifluxURL, cfg.MinifluxAPIKey)

	syncer := NewSyncer(mfClient, store, embedder)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go syncer.Run(ctx)

	webhook := &WebhookHandler{
		Secret:   cfg.WebhookSecret,
		Curator:  curator,
		Updater:  updater,
		Store:    store,
		Embedder: embedder,
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
