package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/8BitTacoSupreme/floxybot/internal/canon"
	"github.com/8BitTacoSupreme/floxybot/internal/config"
	"github.com/8BitTacoSupreme/floxybot/internal/rag"
	"github.com/8BitTacoSupreme/floxybot/internal/voyage"
	"github.com/spf13/cobra"
)

var askCmd = &cobra.Command{
	Use:   "ask [question]",
	Short: "Single-shot Q&A — prints answer to stdout and exits",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runAsk,
}

func init() {
	rootCmd.AddCommand(askCmd)
}

func runAsk(cmd *cobra.Command, args []string) error {
	question := strings.Join(args, " ")
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	voyageKey := cfg.VoyageAPIKey
	if voyageKey == "" {
		return fmt.Errorf("VOYAGE_API_KEY is required (set env or config)")
	}

	voyageEmbed := voyage.NewEmbeddingClient(voyageKey)
	voyageRerank := voyage.NewRerankClient(voyageKey)

	// Load canon snapshot.
	snapPath := filepath.Join(cfg.CanonDir, "canon.gob")
	snap, err := canon.LoadSnapshot(snapPath)
	if err != nil {
		return fmt.Errorf("loading canon snapshot from %s: %w\nRun 'go run ./canon/build/' to build one first.", snapPath, err)
	}

	// Build in-memory vector store from snapshot chunks.
	store, err := rag.NewStore("", voyageEmbed)
	if err != nil {
		return fmt.Errorf("creating store: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Loading %d chunks into vector store...\n", len(snap.Chunks))
	if err := store.AddChunks(ctx, snap.Chunks); err != nil {
		return fmt.Errorf("loading chunks: %w", err)
	}

	// Retrieve.
	retriever := rag.NewRetriever(store, voyageRerank)
	results, err := retriever.Retrieve(ctx, question, 5)
	if err != nil {
		return fmt.Errorf("retrieval: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("No relevant documentation found.")
		return nil
	}

	// Phase 3 will pipe these into Claude. For now, print retrieved chunks.
	fmt.Printf("Question: %s\n\n", question)
	for i, r := range results {
		fmt.Printf("--- Result %d (score: %.4f) ---\n", i+1, r.Score)
		fmt.Printf("Source: %s (%s)\n", r.Title, r.URL)
		fmt.Printf("%s\n\n", r.Text)
	}
	return nil
}
