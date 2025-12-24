package cmd

import (
	"log"
	"time"

	"github.com/karol-broda/snitch/internal/config"
	"github.com/karol-broda/snitch/internal/resolver"
	"github.com/karol-broda/snitch/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/karol-broda/snitch/internal/config"
	"github.com/karol-broda/snitch/internal/tui"
	"github.com/spf13/cobra"
)

// top-specific flags
var (
	topTheme    string
	topInterval time.Duration
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

		// configure resolver with cache setting
		effectiveNoCache := noCache || !cfg.Defaults.DNSCache
		resolver.SetNoCache(effectiveNoCache)

		opts := tui.Options{
			Theme:        theme,
			Interval:     topInterval,
			ResolveAddrs: resolveAddrs,
			ResolvePorts: resolvePorts,
			NoCache:      effectiveNoCache,
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
	topCmd.Flags().StringVar(&topTheme, "theme", cfg.Defaults.Theme, "Theme for TUI (see 'snitch themes')")
	topCmd.Flags().DurationVarP(&topInterval, "interval", "i", time.Second, "Refresh interval")

	// shared flags
	addFilterFlags(topCmd)
	addResolutionFlags(topCmd)
}