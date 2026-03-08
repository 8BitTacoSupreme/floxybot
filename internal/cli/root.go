package cli

import (
	"github.com/spf13/cobra"
)

var (
	flagAPIKey  string
	flagModel   string
	flagVerbose bool
	appVersion  = "dev"
)

func SetVersion(v string) { appVersion = v }

var rootCmd = &cobra.Command{
	Use:   "floxybot",
	Short: "AI assistant for Flox — advice, collaboration, automation",
	Long: `Floxybot is a standalone CLI and TUI for getting help with Flox.

It provides:
  - Chat: Conversational Q&A backed by Claude + local RAG over Flox docs
  - Co-Pilot: Agent automation via Flox MCP server
  - Context: Awareness of your current Flox environment`,
	// Default action is to launch the TUI (same as `chat`).
	RunE: func(cmd *cobra.Command, args []string) error {
		return chatCmd.RunE(cmd, args)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagAPIKey, "api-key", "", "Anthropic API key (or set ANTHROPIC_API_KEY)")
	rootCmd.PersistentFlags().StringVar(&flagModel, "model", "claude-sonnet-4-20250514", "Claude model to use")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "Enable verbose output")
}

func Execute() error {
	return rootCmd.Execute()
}
