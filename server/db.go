package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/lib/pq"
)

// FeedbackStore is the interface for persisting feedback.
type FeedbackStore interface {
	SaveFeedback(ctx context.Context, f FeedbackPayload) error
	Close() error
}

// PostgresStore stores feedback in PostgreSQL.
type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(connStr string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	// Create table if not exists.
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS feedback (
			id SERIAL PRIMARY KEY,
			query TEXT NOT NULL,
			response_hash TEXT,
			vote TEXT NOT NULL CHECK (vote IN ('up', 'down')),
			comment TEXT,
			sources JSONB,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("creating table: %w", err)
	}

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) SaveFeedback(ctx context.Context, f FeedbackPayload) error {
	sourcesJSON, _ := json.Marshal(f.Sources)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO feedback (query, response_hash, vote, comment, sources) VALUES ($1, $2, $3, $4, $5)`,
		f.Query, f.ResponseHash, f.Vote, f.Comment, sourcesJSON,
	)
	return err
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}

// NoopStore discards feedback (used when no database is configured).
type NoopStore struct{}

func (s *NoopStore) SaveFeedback(ctx context.Context, f FeedbackPayload) error { return nil }
func (s *NoopStore) Close() error                                               { return nil }
