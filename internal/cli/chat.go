package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/8BitTacoSupreme/floxybot/internal/canon"
	"github.com/8BitTacoSupreme/floxybot/internal/claude"
	"github.com/8BitTacoSupreme/floxybot/internal/config"
	"github.com/8BitTacoSupreme/floxybot/internal/floxctx"
	"github.com/8BitTacoSupreme/floxybot/internal/rag"
	"github.com/8BitTacoSupreme/floxybot/internal/tui"
	"github.com/8BitTacoSupreme/floxybot/internal/voyage"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Launch the tabbed TUI for interactive Q&A",
	RunE:  runChat,
}

func init() {
	rootCmd.AddCommand(chatCmd)
}

func runChat(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	applyFlags(cfg)

	fctx, _ := floxctx.Detect()

	// Set up Claude client (nil if no API key).
	var cc *claude.Client
	if cfg.AnthropicAPIKey != "" {
		cc = claude.NewClient(cfg.AnthropicAPIKey, cfg.Model)
	} else {
		fmt.Fprintln(os.Stderr, "Warning: ANTHROPIC_API_KEY not set — chat will show RAG results only.")
	}

	// Set up retriever (nil if no Voyage key or no snapshot).
	var retriever *rag.Retriever
	if cfg.VoyageAPIKey != "" {
		snapPath := filepath.Join(cfg.CanonDir, "canon.gob")
		if _, err := os.Stat(snapPath); err == nil {
			snap, err := canon.LoadSnapshot(snapPath)
			if err == nil {
				voyageEmbed := voyage.NewEmbeddingClient(cfg.VoyageAPIKey)
				store, err := rag.NewStore("", voyageEmbed)
				if err == nil {
					if err := store.AddChunks(cmd.Context(), snap.Chunks); err == nil {
						voyageRerank := voyage.NewRerankClient(cfg.VoyageAPIKey)
						retriever = rag.NewRetriever(store, voyageRerank)
					}
				}
			}
		}
	}

	app := tui.NewApp(cc, retriever, fctx)
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	return nil
}

func applyFlags(cfg *config.Config) {
	if flagAPIKey != "" {
		cfg.AnthropicAPIKey = flagAPIKey
	}
	if flagModel != "" {
		cfg.Model = flagModel
	}
	cfg.Verbose = flagVerbose
}
