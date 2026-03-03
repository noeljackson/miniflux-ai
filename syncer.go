package main

import (
	"context"
	"log"
	"time"

	miniflux "miniflux.app/v2/client"
)

type Syncer struct {
	miniflux *miniflux.Client
	store    *VectorStore
	embedder Embedder
	interval time.Duration
}

func NewSyncer(client *miniflux.Client, store *VectorStore, embedder Embedder) *Syncer {
	return &Syncer{
		miniflux: client,
		store:    store,
		embedder: embedder,
		interval: 30 * time.Minute,
	}
}

func (s *Syncer) Run(ctx context.Context) {
	if err := s.syncOnce(ctx); err != nil {
		log.Printf("syncer: initial sync failed: %v", err)
	}

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("syncer: stopped")
			return
		case <-ticker.C:
			if err := s.syncOnce(ctx); err != nil {
				log.Printf("syncer: sync failed: %v", err)
			}
		}
	}
}

func (s *Syncer) syncOnce(ctx context.Context) error {
	result, err := s.miniflux.Entries(&miniflux.Filter{
		Starred:   miniflux.FilterOnlyStarred,
		Limit:     100,
		Order:     "published_at",
		Direction: "desc",
	})
	if err != nil {
		return err
	}

	existing, err := s.store.EntryIDs(ctx)
	if err != nil {
		return err
	}

	var added int
	for _, entry := range result.Entries {
		if existing[entry.ID] {
			continue
		}

		embedding, err := s.embedder.Embed(ctx, entry.Title)
		if err != nil {
			log.Printf("syncer: embed failed for entry %d: %v", entry.ID, err)
			continue
		}

		if err := s.store.Upsert(ctx, entry.ID, entry.Title, embedding); err != nil {
			log.Printf("syncer: upsert failed for entry %d: %v", entry.ID, err)
			continue
		}
		added++
	}

	log.Printf("syncer: refreshed (%d starred, %d new embeddings)", len(result.Entries), added)
	return nil
}
