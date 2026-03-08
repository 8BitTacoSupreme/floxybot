package rag

import (
	"context"
	"fmt"
	"strings"

	"github.com/8BitTacoSupreme/floxybot/internal/voyage"
)

// RetrievedChunk is a document chunk with its relevance score after reranking.
type RetrievedChunk struct {
	Text  string
	URL   string
	Title string
	Score float64
}

// Retriever performs the full RAG retrieval pipeline:
// query → vector search (top-20) → Voyage rerank → top-5.
type Retriever struct {
	store    *Store
	reranker *voyage.RerankClient
}

func NewRetriever(store *Store, reranker *voyage.RerankClient) *Retriever {
	return &Retriever{store: store, reranker: reranker}
}

// Retrieve executes the full pipeline and returns the top-k reranked chunks.
func (r *Retriever) Retrieve(ctx context.Context, query string, topK int) ([]RetrievedChunk, error) {
	// Step 1: Vector search — get top 20 candidates.
	candidates, err := r.store.Query(ctx, query, 20)
	if err != nil {
		return nil, fmt.Errorf("vector search: %w", err)
	}

	if len(candidates) == 0 {
		return nil, nil
	}

	// Step 2: Rerank with Voyage.
	docs := make([]string, len(candidates))
	for i, c := range candidates {
		docs[i] = c.Content
	}

	reranked, err := r.reranker.Rerank(ctx, query, docs, topK)
	if err != nil {
		// Fall back to raw vector results if reranking fails.
		fmt.Printf("rerank failed, using raw vector results: %v\n", err)
		var results []RetrievedChunk
		limit := topK
		if limit > len(candidates) {
			limit = len(candidates)
		}
		for _, c := range candidates[:limit] {
			results = append(results, RetrievedChunk{
				Text:  c.Content,
				URL:   c.Metadata["url"],
				Title: c.Metadata["title"],
				Score: float64(c.Similarity),
			})
		}
		return results, nil
	}

	// Step 3: Map reranked indices back to candidates.
	var results []RetrievedChunk
	for _, rr := range reranked {
		if rr.Index >= len(candidates) {
			continue
		}
		c := candidates[rr.Index]
		results = append(results, RetrievedChunk{
			Text:  c.Content,
			URL:   c.Metadata["url"],
			Title: c.Metadata["title"],
			Score: rr.RelevanceScore,
		})
	}
	return results, nil
}

// FormatForPrompt formats retrieved chunks as context for injection into a system prompt.
func FormatForPrompt(chunks []RetrievedChunk) string {
	if len(chunks) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("Relevant documentation:\n\n")
	for i, c := range chunks {
		b.WriteString(fmt.Sprintf("--- Source %d: %s (%s) ---\n%s\n\n", i+1, c.Title, c.URL, c.Text))
	}
	return b.String()
}
