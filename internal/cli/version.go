package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print floxybot version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("floxybot %s\n", appVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
