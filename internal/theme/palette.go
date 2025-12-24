package theme

import (
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

// Palette defines the semantic colors for a theme
type Palette struct {
	Name string

	// base colors
	Fg       string // primary foreground
	FgMuted  string // secondary/muted foreground
	FgSubtle string // subtle/disabled foreground
	Bg       string // primary background
	BgMuted  string // secondary background (selections, highlights)
	Border   string // border color

	// semantic colors
	Red     string
	Green   string
	Yellow  string
	Blue    string
	Magenta string
	Cyan    string
	Orange  string
	Gray    string
}

// Color converts a palette color string to a lipgloss.TerminalColor.
// If the string is 1-2 characters, it's treated as an ANSI color code.
// Otherwise, it's treated as a hex color.
func (p *Palette) Color(c string) lipgloss.TerminalColor {
	if c == "" {
		return lipgloss.NoColor{}
	}

	if len(c) <= 2 {
		n, err := strconv.Atoi(c)
		if err == nil {
			return lipgloss.ANSIColor(n)
		}
	}

	return lipgloss.Color(c)
}

// ToTheme converts a Palette to a Theme with lipgloss styles
func (p *Palette) ToTheme() *Theme {
	return &Theme{
		Name: p.Name,
		Styles: Styles{
			Header: lipgloss.NewStyle().
				Bold(true).
				Foreground(p.Color(p.Fg)),

			Border: lipgloss.NewStyle().
				Foreground(p.Color(p.Border)),

			Selected: lipgloss.NewStyle().
				Bold(true).
				Foreground(p.Color(p.Fg)),

			Watched: lipgloss.NewStyle().
				Bold(true).
				Foreground(p.Color(p.Orange)),

			Normal: lipgloss.NewStyle().
				Foreground(p.Color(p.FgMuted)),

			Error: lipgloss.NewStyle().
				Foreground(p.Color(p.Red)),

			Success: lipgloss.NewStyle().
				Foreground(p.Color(p.Green)),

			Warning: lipgloss.NewStyle().
				Foreground(p.Color(p.Yellow)),

			Footer: lipgloss.NewStyle().
				Foreground(p.Color(p.FgSubtle)),

			Background: lipgloss.NewStyle(),

			Proto: ProtoStyles{
				TCP:  lipgloss.NewStyle().Foreground(p.Color(p.Green)),
				UDP:  lipgloss.NewStyle().Foreground(p.Color(p.Magenta)),
				Unix: lipgloss.NewStyle().Foreground(p.Color(p.Gray)),
				TCP6: lipgloss.NewStyle().Foreground(p.Color(p.Cyan)),
				UDP6: lipgloss.NewStyle().Foreground(p.Color(p.Blue)),
			},

			State: StateStyles{
				Listen:      lipgloss.NewStyle().Foreground(p.Color(p.Green)),
				Established: lipgloss.NewStyle().Foreground(p.Color(p.Blue)),
				TimeWait:    lipgloss.NewStyle().Foreground(p.Color(p.Yellow)),
				CloseWait:   lipgloss.NewStyle().Foreground(p.Color(p.Orange)),
				SynSent:     lipgloss.NewStyle().Foreground(p.Color(p.Magenta)),
				SynRecv:     lipgloss.NewStyle().Foreground(p.Color(p.Magenta)),
				FinWait1:    lipgloss.NewStyle().Foreground(p.Color(p.Red)),
				FinWait2:    lipgloss.NewStyle().Foreground(p.Color(p.Red)),
				Closing:     lipgloss.NewStyle().Foreground(p.Color(p.Red)),
				LastAck:     lipgloss.NewStyle().Foreground(p.Color(p.Red)),
				Closed:      lipgloss.NewStyle().Foreground(p.Color(p.Gray)),
			},
		},
	}
}

