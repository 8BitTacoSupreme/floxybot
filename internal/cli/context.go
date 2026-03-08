package cli

import (
	"fmt"

	"github.com/8BitTacoSupreme/floxybot/internal/floxctx"
	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Show detected Flox environment info",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := floxctx.Detect()
		if err != nil {
			return fmt.Errorf("detecting Flox context: %w", err)
		}
		fmt.Print(ctx.Format())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(contextCmd)
}
