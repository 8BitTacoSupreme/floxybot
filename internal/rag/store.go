package rag

import (
	"context"
	"fmt"

	"github.com/8BitTacoSupreme/floxybot/internal/voyage"
	"github.com/philippgille/chromem-go"
)

// Store wraps chromem-go for document storage and retrieval.
type Store struct {
	db         *chromem.DB
	collection *chromem.Collection
}

// NewStore creates a new vector store. If persistDir is empty, uses in-memory storage.
func NewStore(persistDir string, voyageClient *voyage.EmbeddingClient) (*Store, error) {
	var db *chromem.DB
	var err error

	if persistDir != "" {
		db, err = chromem.NewPersistentDB(persistDir, false)
	} else {
		db = chromem.NewDB()
	}
	if err != nil {
		return nil, fmt.Errorf("creating chromem db: %w", err)
	}

	embeddingFunc := voyageEmbeddingFunc(voyageClient)

	col, err := db.GetOrCreateCollection("flox_docs", nil, embeddingFunc)
	if err != nil {
		return nil, fmt.Errorf("creating collection: %w", err)
	}

	return &Store{db: db, collection: col}, nil
}

// AddChunks adds document chunks to the store. Embeddings are computed via Voyage.
func (s *Store) AddChunks(ctx context.Context, chunks []Chunk) error {
	docs := make([]chromem.Document, len(chunks))
	for i, c := range chunks {
		docs[i] = chromem.Document{
			ID:      fmt.Sprintf("%s#%d", c.URL, c.Index),
			Content: c.Text,
			Metadata: map[string]string{
				"url":   c.URL,
				"title": c.Title,
			},
		}
	}

	// Use small batches with concurrency=1 to respect Voyage rate limits.
	const batchSize = 10
	for i := 0; i < len(docs); i += batchSize {
		end := i + batchSize
		if end > len(docs) {
			end = len(docs)
		}
		if err := s.collection.AddDocuments(ctx, docs[i:end], 1); err != nil {
			return fmt.Errorf("adding batch %d: %w", i/batchSize, err)
		}
	}
	return nil
}

// Query returns the top-n most similar documents to the query text.
func (s *Store) Query(ctx context.Context, query string, n int) ([]chromem.Result, error) {
	return s.collection.Query(ctx, query, n, nil, nil)
}

// Count returns the number of documents in the store.
func (s *Store) Count() int {
	return s.collection.Count()
}

// voyageEmbeddingFunc adapts the Voyage client to chromem-go's EmbeddingFunc type.
func voyageEmbeddingFunc(vc *voyage.EmbeddingClient) chromem.EmbeddingFunc {
	return func(ctx context.Context, text string) ([]float32, error) {
		return vc.EmbedOne(ctx, text)
	}
}
