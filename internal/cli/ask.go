package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/8BitTacoSupreme/floxybot/internal/canon"
	"github.com/8BitTacoSupreme/floxybot/internal/claude"
	"github.com/8BitTacoSupreme/floxybot/internal/config"
	"github.com/8BitTacoSupreme/floxybot/internal/floxctx"
	"github.com/8BitTacoSupreme/floxybot/internal/rag"
	"github.com/8BitTacoSupreme/floxybot/internal/voyage"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/spf13/cobra"
)

var noRAG bool

var askCmd = &cobra.Command{
	Use:   "ask [question]",
	Short: "Single-shot Q&A — prints answer to stdout and exits",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runAsk,
}

func init() {
	askCmd.Flags().BoolVar(&noRAG, "no-rag", false, "Skip RAG retrieval, ask Claude directly")
	rootCmd.AddCommand(askCmd)
}

func runAsk(cmd *cobra.Command, args []string) error {
	question := strings.Join(args, " ")
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Build system prompt with Flox context if available.
	systemPrompt := "You are floxybot, an AI assistant specializing in Flox (https://flox.dev). " +
		"You help users install packages, set up development environments, and use Flox effectively. " +
		"Be concise and practical."

	if fctx, err := floxctx.Detect(); err == nil {
		systemPrompt += "\n\nCurrent Flox environment:\n" + fctx.ForPrompt()
	}

	if noRAG {
		return askClaude(ctx, cfg, systemPrompt, question)
	}

	return askWithRAG(ctx, cfg, systemPrompt, question)
}

func askClaude(ctx context.Context, cfg *config.Config, systemPrompt, question string) error {
	if cfg.AnthropicAPIKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY is required (set env or config)")
	}

	client := claude.NewClient(cfg.AnthropicAPIKey, cfg.Model)
	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(question)),
	}

	fmt.Fprintf(os.Stderr, "Asking Claude (no RAG)...\n")
	response, err := client.Chat(ctx, systemPrompt, messages)
	if err != nil {
		return fmt.Errorf("claude: %w", err)
	}

	fmt.Println(response)
	return nil
}

func askWithRAG(ctx context.Context, cfg *config.Config, systemPrompt, question string) error {
	voyageKey := cfg.VoyageAPIKey
	if voyageKey == "" {
		return fmt.Errorf("VOYAGE_API_KEY is required (set env or config, or use --no-rag)")
	}

	voyageEmbed := voyage.NewEmbeddingClient(voyageKey)
	voyageRerank := voyage.NewRerankClient(voyageKey)

	// Load canon snapshot.
	snapPath := filepath.Join(cfg.CanonDir, "canon.gob")
	snap, err := canon.LoadSnapshot(snapPath)
	if err != nil {
		return fmt.Errorf("loading canon snapshot from %s: %w\nRun 'floxybot canon update' to download one.", snapPath, err)
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

	// If we have an Anthropic key, pipe through Claude. Otherwise print raw chunks.
	if cfg.AnthropicAPIKey != "" {
		var ragContext strings.Builder
		for _, r := range results {
			ragContext.WriteString(fmt.Sprintf("Source: %s (%s)\n%s\n\n", r.Title, r.URL, r.Text))
		}
		enrichedPrompt := systemPrompt + "\n\nRelevant Flox documentation:\n" + ragContext.String()

		client := claude.NewClient(cfg.AnthropicAPIKey, cfg.Model)
		messages := []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(question)),
		}

		response, err := client.Chat(ctx, enrichedPrompt, messages)
		if err != nil {
			return fmt.Errorf("claude: %w", err)
		}
		fmt.Println(response)
		return nil
	}

	// No Claude key — print raw retrieved chunks.
	fmt.Printf("Question: %s\n\n", question)
	for i, r := range results {
		fmt.Printf("--- Result %d (score: %.4f) ---\n", i+1, r.Score)
		fmt.Printf("Source: %s (%s)\n", r.Title, r.URL)
		fmt.Printf("%s\n\n", r.Text)
	}
	return nil
}
