package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	pgvector "github.com/pgvector/pgvector-go"
)

type VectorStore struct {
	pool *pgxpool.Pool
}

func NewVectorStore(ctx context.Context, databaseURL string) (*VectorStore, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &VectorStore{pool: pool}, nil
}

func (v *VectorStore) Migrate(ctx context.Context) error {
	_, err := v.pool.Exec(ctx, `
		CREATE EXTENSION IF NOT EXISTS vector;
		CREATE TABLE IF NOT EXISTS starred_embeddings (
			entry_id BIGINT PRIMARY KEY,
			title TEXT NOT NULL,
			embedding vector(768) NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS starred_embeddings_cosine_idx
			ON starred_embeddings USING hnsw (embedding vector_cosine_ops);
	`)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	return nil
}

func (v *VectorStore) Upsert(ctx context.Context, entryID int64, title string, embedding []float32) error {
	vec := pgvector.NewVector(embedding)
	_, err := v.pool.Exec(ctx, `
		INSERT INTO starred_embeddings (entry_id, title, embedding)
		VALUES ($1, $2, $3)
		ON CONFLICT (entry_id) DO UPDATE SET title = $2, embedding = $3
	`, entryID, title, vec)
	if err != nil {
		return fmt.Errorf("upsert embedding: %w", err)
	}
	return nil
}

func (v *VectorStore) Similarity(ctx context.Context, embedding []float32) (float64, error) {
	vec := pgvector.NewVector(embedding)
	var similarity float64
	err := v.pool.QueryRow(ctx, `
		SELECT COALESCE(MAX(1 - (embedding <=> $1)), 0)
		FROM (
			SELECT embedding FROM starred_embeddings
			ORDER BY embedding <=> $1
			LIMIT 5
		) sub
	`, vec).Scan(&similarity)
	if err != nil {
		return 0, fmt.Errorf("similarity query: %w", err)
	}
	return similarity, nil
}

func (v *VectorStore) EntryIDs(ctx context.Context) (map[int64]bool, error) {
	rows, err := v.pool.Query(ctx, `SELECT entry_id FROM starred_embeddings`)
	if err != nil {
		return nil, fmt.Errorf("query entry IDs: %w", err)
	}
	defer rows.Close()

	ids := make(map[int64]bool)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan entry ID: %w", err)
		}
		ids[id] = true
	}
	return ids, rows.Err()
}

func (v *VectorStore) Close() {
	v.pool.Close()
}
