package cmd

import (
	"fmt"

	"github.com/karol-broda/snitch/internal/theme"
	"github.com/spf13/cobra"
)

var themesCmd = &cobra.Command{
	Use:   "themes",
	Short: "List available themes",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Available themes (default: %s):\n\n", theme.DefaultTheme)
		for _, name := range theme.ListThemes() {
			fmt.Printf("  %s\n", name)
		}
	},
}

func init() {
	rootCmd.AddCommand(themesCmd)
}

