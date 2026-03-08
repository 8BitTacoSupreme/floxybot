package cli

import (
	"fmt"

	"github.com/8BitTacoSupreme/floxybot/internal/canon"
	"github.com/8BitTacoSupreme/floxybot/internal/config"
	"github.com/spf13/cobra"
)

var canonCmd = &cobra.Command{
	Use:   "canon",
	Short: "Manage the documentation canon",
}

var canonUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Force canon refresh from remote",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		updater := canon.NewUpdater(cfg.BackendURL, cfg.CanonDir)
		return updater.Download()
	},
}

func init() {
	canonCmd.AddCommand(canonUpdateCmd)
	rootCmd.AddCommand(canonCmd)
}
