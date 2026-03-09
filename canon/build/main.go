package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/8BitTacoSupreme/floxybot/internal/canon"
	"github.com/8BitTacoSupreme/floxybot/internal/rag"
	"github.com/8BitTacoSupreme/floxybot/internal/voyage"
)

func main() {
	outputDir := os.Getenv("FLOXYBOT_CANON_DIR")
	if outputDir == "" {
		outputDir = filepath.Join("data", "canon")
	}
	outputPath := filepath.Join(outputDir, "canon.gob")

	// If a chunks.json exists (from Python scraper), convert it to gob.
	jsonPath := filepath.Join(outputDir, "chunks.json")
	if _, err := os.Stat(jsonPath); err == nil {
		fmt.Printf("Converting %s -> %s\n", jsonPath, outputPath)
		if err := convertJSON(jsonPath, outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Done.")
		return
	}

	// Otherwise, scrape directly (works on machines with enough memory).
	fmt.Println("Building floxybot canon snapshot...")
	fmt.Printf("Output: %s\n", outputPath)
	if err := canon.BuildCanon(context.Background(), outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Done.")
}

type jsonChunk struct {
	Text  string `json:"text"`
	URL   string `json:"url"`
	Title string `json:"title"`
	Index int    `json:"index"`
}

func convertJSON(jsonPath, gobPath string) error {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}
	var jchunks []jsonChunk
	if err := json.Unmarshal(data, &jchunks); err != nil {
		return fmt.Errorf("parsing JSON: %w", err)
	}

	chunks := make([]rag.Chunk, len(jchunks))
	for i, jc := range jchunks {
		chunks[i] = rag.Chunk{
			Text:  jc.Text,
			URL:   jc.URL,
			Title: jc.Title,
			Index: jc.Index,
		}
	}

	// Pre-compute embeddings if VOYAGE_API_KEY is set.
	voyageKey := os.Getenv("VOYAGE_API_KEY")
	if voyageKey != "" {
		fmt.Printf("Pre-computing embeddings for %d chunks...\n", len(chunks))
		if err := embedChunks(chunks, voyageKey); err != nil {
			return fmt.Errorf("embedding chunks: %w", err)
		}
	} else {
		fmt.Println("No VOYAGE_API_KEY set — skipping embedding pre-computation.")
		fmt.Println("Users will need VOYAGE_API_KEY at runtime.")
	}

	snap := &canon.Snapshot{
		Version:   "2",
		CreatedAt: time.Now(),
		Chunks:    chunks,
	}

	if err := canon.SaveSnapshot(gobPath, snap); err != nil {
		return err
	}

	fmt.Printf("Converted %d chunks -> %s\n", len(chunks), gobPath)
	return nil
}

func embedChunks(chunks []rag.Chunk, apiKey string) error {
	client := voyage.NewEmbeddingClient(apiKey)
	ctx := context.Background()

	// Batch embed to minimize API calls. Voyage accepts up to 128 texts per call.
	const batchSize = 64
	for i := 0; i < len(chunks); i += batchSize {
		end := i + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}

		texts := make([]string, end-i)
		for j := i; j < end; j++ {
			texts[j-i] = chunks[j].Text
		}

		embeddings, err := client.Embed(ctx, texts)
		if err != nil {
			return fmt.Errorf("batch %d-%d: %w", i, end, err)
		}

		for j, emb := range embeddings {
			chunks[i+j].Embedding = emb
		}

		done := end
		if done > len(chunks) {
			done = len(chunks)
		}
		fmt.Printf("  embedded %d/%d chunks\n", done, len(chunks))
	}
	return nil
}
