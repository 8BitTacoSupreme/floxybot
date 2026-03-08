package cli

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/8BitTacoSupreme/floxybot/internal/agent"
	"github.com/8BitTacoSupreme/floxybot/internal/claude"
	"github.com/8BitTacoSupreme/floxybot/internal/config"
	"github.com/8BitTacoSupreme/floxybot/internal/floxctx"
	"github.com/8BitTacoSupreme/floxybot/internal/mcp"
	"github.com/spf13/cobra"
)

var flagYes bool

var agentCmd = &cobra.Command{
	Use:   "agent [task]",
	Short: "Run a task via the Claude tool-use agent loop",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runAgent,
}

func init() {
	agentCmd.Flags().BoolVar(&flagYes, "yes", false, "Auto-approve destructive tool calls")
	rootCmd.AddCommand(agentCmd)
}

func runAgent(cmd *cobra.Command, args []string) error {
	task := strings.Join(args, " ")
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	applyFlags(cfg)

	if cfg.AnthropicAPIKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY is required for agent mode")
	}

	// Find flox-mcp binary.
	mcpBin, err := exec.LookPath("flox-mcp-server")
	if err != nil {
		return fmt.Errorf("flox-mcp-server not found in PATH (install via Flox): %w", err)
	}

	// Start MCP subprocess.
	fmt.Println("Starting Flox MCP server...")
	mcpClient, err := mcp.NewClient(mcpBin)
	if err != nil {
		return fmt.Errorf("starting MCP: %w", err)
	}
	defer mcpClient.Close()

	tools := mcpClient.Tools()
	fmt.Printf("Discovered %d MCP tools\n", len(tools))

	// Detect Flox context.
	fctx, _ := floxctx.Detect()

	// Create Claude client.
	cc := claude.NewClient(cfg.AnthropicAPIKey, cfg.Model)

	// Create executor and agent loop.
	executor := agent.NewExecutor(mcpClient)
	loop := agent.NewLoop(cc, executor, tools, fctx, flagYes, func(step string) {
		fmt.Printf("  %s\n", step)
	})

	fmt.Printf("Running task: %s\n\n", task)
	result, err := loop.Run(ctx, task)
	if err != nil {
		return fmt.Errorf("agent loop: %w", err)
	}

	fmt.Println("\n" + result)
	return nil
}
