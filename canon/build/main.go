package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/8BitTacoSupreme/floxybot/internal/canon"
	"github.com/8BitTacoSupreme/floxybot/internal/rag"
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

	snap := &canon.Snapshot{
		Version: "1",
		Chunks:  chunks,
	}

	if err := canon.SaveSnapshot(gobPath, snap); err != nil {
		return err
	}

	fmt.Printf("Converted %d chunks -> %s\n", len(chunks), gobPath)
	return nil
}
