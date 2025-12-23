package cmd

import (
	"log"
	"github.com/karol-broda/snitch/internal/config"
	"github.com/karol-broda/snitch/internal/tui"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// top-specific flags
var (
	topTheme        string
	topInterval     time.Duration
	topResolveAddrs bool
	topResolvePorts bool
)

var topCmd = &cobra.Command{
	Use:   "top",
	Short: "Live TUI for inspecting connections",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()

		theme := topTheme
		if theme == "" {
			theme = cfg.Defaults.Theme
		}

		opts := tui.Options{
			Theme:        theme,
			Interval:     topInterval,
			ResolveAddrs: topResolveAddrs,
			ResolvePorts: topResolvePorts,
		}

		// if any filter flag is set, use exclusive mode
		if filterTCP || filterUDP || filterListen || filterEstab {
			opts.TCP = filterTCP
			opts.UDP = filterUDP
			opts.Listening = filterListen
			opts.Established = filterEstab
			opts.Other = false
			opts.FilterSet = true
		}

		m := tui.New(opts)

		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(topCmd)
	cfg := config.Get()

	// top-specific flags
	topCmd.Flags().StringVar(&topTheme, "theme", cfg.Defaults.Theme, "Theme for TUI (dark, light, mono, auto)")
	topCmd.Flags().DurationVarP(&topInterval, "interval", "i", time.Second, "Refresh interval")
	topCmd.Flags().BoolVar(&topResolveAddrs, "resolve-addrs", !cfg.Defaults.Numeric, "Resolve IP addresses to hostnames")
	topCmd.Flags().BoolVar(&topResolvePorts, "resolve-ports", false, "Resolve port numbers to service names")

	// shared filter flags
	addFilterFlags(topCmd)
}