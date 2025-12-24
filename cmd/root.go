package cmd

import (
	"fmt"
	"os"

	"github.com/karol-broda/snitch/internal/config"
	"github.com/spf13/cobra"
)


var (
	cfgFile string
)

var rootCmd = &cobra.Command{
	Use:   "snitch",
	Short: "snitch is a tool for inspecting network connections",
	Long:  `snitch is a tool for inspecting network connections

A modern, unix-y tool for inspecting network connections, with a focus on a clear usage API and a solid testing strategy.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if _, err := config.Load(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Error loading config: %v\n", err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// default to top - flags are shared so they work here too
		topCmd.Run(cmd, args)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/snitch/snitch.toml)")
	rootCmd.PersistentFlags().Bool("debug", false, "enable debug logs to stderr")

	// add top's flags to root so `snitch -l` works (defaults to top command)
	cfg := config.Get()
	rootCmd.Flags().StringVar(&topTheme, "theme", cfg.Defaults.Theme, "Theme for TUI (see 'snitch themes')")
	rootCmd.Flags().DurationVarP(&topInterval, "interval", "i", 0, "Refresh interval (default 1s)")

	// shared flags for root command
	addFilterFlags(rootCmd)
	addResolutionFlags(rootCmd)
}