package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Theme represents the visual styling for the TUI
type Theme struct {
	Name   string
	Styles Styles
}

// Styles contains all the styling definitions
type Styles struct {
	Header     lipgloss.Style
	Border     lipgloss.Style
	Selected   lipgloss.Style
	Watched    lipgloss.Style
	Normal     lipgloss.Style
	Error      lipgloss.Style
	Success    lipgloss.Style
	Warning    lipgloss.Style
	Proto      ProtoStyles
	State      StateStyles
	Footer     lipgloss.Style
	Background lipgloss.Style
}

// ProtoStyles contains protocol-specific colors
type ProtoStyles struct {
	TCP  lipgloss.Style
	UDP  lipgloss.Style
	Unix lipgloss.Style
	TCP6 lipgloss.Style
	UDP6 lipgloss.Style
}

// StateStyles contains connection state-specific colors
type StateStyles struct {
	Listen      lipgloss.Style
	Established lipgloss.Style
	TimeWait    lipgloss.Style
	CloseWait   lipgloss.Style
	SynSent     lipgloss.Style
	SynRecv     lipgloss.Style
	FinWait1    lipgloss.Style
	FinWait2    lipgloss.Style
	Closing     lipgloss.Style
	LastAck     lipgloss.Style
	Closed      lipgloss.Style
}

var (
	themes map[string]*Theme
)

func init() {
	themes = map[string]*Theme{
		"default":            createAdaptiveTheme(),
		"mono":               createMonoTheme(),
		"dracula":            createDraculaTheme(),
		"gruvbox":            createGruvboxTheme(),
		"gruvbox-light":      createGruvboxLightTheme(),
		"nord":               createNordTheme(),
		"catppuccin-latte":   createCatppuccinLatteTheme(),
		"catppuccin-frappe":  createCatppuccinFrappeTheme(),
		"catppuccin-macchiato": createCatppuccinMacchiatoTheme(),
		"catppuccin-mocha":   createCatppuccinMochaTheme(),
	}
}

// GetTheme returns a theme by name, with auto-detection support
func GetTheme(name string) *Theme {
	if name == "auto" {
		// lipgloss handles adaptive colors, so we just return the default
		return themes["default"]
	}

	if theme, exists := themes[name]; exists {
		return theme
	}

	// a specific theme was requested (e.g. "dark", "light"), but we now use adaptive
	// so we can just return the default theme and lipgloss will handle it
	if name == "dark" || name == "light" {
		return themes["default"]
	}

	// fallback to default
	return themes["default"]
}

// ListThemes returns available theme names
func ListThemes() []string {
	var names []string
	for name := range themes {
		names = append(names, name)
	}
	return names
}

// createAdaptiveTheme creates a clean, minimal theme
func createAdaptiveTheme() *Theme {
	return &Theme{
		Name: "default",
		Styles: Styles{
			Header: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.AdaptiveColor{Light: "#1F2937", Dark: "#F9FAFB"}),

			Watched: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.AdaptiveColor{Light: "#D97706", Dark: "#F59E0B"}),

			Border: lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#D1D5DB", Dark: "#374151"}),

			Selected: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.AdaptiveColor{Light: "#1F2937", Dark: "#F9FAFB"}),

			Normal: lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"}),

			Error: lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#F87171"}),

			Success: lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#059669", Dark: "#34D399"}),

			Warning: lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#D97706", Dark: "#FBBF24"}),

			Footer: lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#6B7280"}),

			Background: lipgloss.NewStyle(),

			Proto: ProtoStyles{
				TCP:  lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#059669", Dark: "#34D399"}),
				UDP:  lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"}),
				Unix: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"}),
				TCP6: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#059669", Dark: "#34D399"}),
				UDP6: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"}),
			},

			State: StateStyles{
				Listen:      lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#059669", Dark: "#34D399"}),
				Established: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#2563EB", Dark: "#60A5FA"}),
				TimeWait:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#D97706", Dark: "#FBBF24"}),
				CloseWait:   lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#D97706", Dark: "#FBBF24"}),
				SynSent:     lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"}),
				SynRecv:     lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"}),
				FinWait1:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#F87171"}),
				FinWait2:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#F87171"}),
				Closing:     lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#F87171"}),
				LastAck:     lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#F87171"}),
				Closed:      lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#6B7280"}),
			},
		},
	}
}

// createMonoTheme creates a monochrome theme (no colors)
func createMonoTheme() *Theme {
	baseStyle := lipgloss.NewStyle()
	boldStyle := lipgloss.NewStyle().Bold(true)

	return &Theme{
		Name: "mono",
		Styles: Styles{
			Header:     boldStyle,
			Border:     baseStyle,
			Selected:   boldStyle,
			Normal:     baseStyle,
			Error:      boldStyle,
			Success:    boldStyle,
			Warning:    boldStyle,
			Footer:     baseStyle,
			Background: baseStyle,

			Proto: ProtoStyles{
				TCP:  baseStyle,
				UDP:  baseStyle,
				Unix: baseStyle,
				TCP6: baseStyle,
				UDP6: baseStyle,
			},

			State: StateStyles{
				Listen:      baseStyle,
				Established: baseStyle,
				TimeWait:    baseStyle,
				CloseWait:   baseStyle,
				SynSent:     baseStyle,
				SynRecv:     baseStyle,
				FinWait1:    baseStyle,
				FinWait2:    baseStyle,
				Closing:     baseStyle,
				LastAck:     baseStyle,
				Closed:      baseStyle,
			},
		},
	}
}

// GetProtoStyle returns the appropriate style for a protocol
func (s *Styles) GetProtoStyle(proto string) lipgloss.Style {
	switch strings.ToLower(proto) {
	case "tcp":
		return s.Proto.TCP
	case "udp":
		return s.Proto.UDP
	case "unix":
		return s.Proto.Unix
	case "tcp6":
		return s.Proto.TCP6
	case "udp6":
		return s.Proto.UDP6
	default:
		return s.Normal
	}
}

// GetStateStyle returns the appropriate style for a connection state
func (s *Styles) GetStateStyle(state string) lipgloss.Style {
	switch strings.ToUpper(state) {
	case "LISTEN":
		return s.State.Listen
	case "ESTABLISHED":
		return s.State.Established
	case "TIME_WAIT":
		return s.State.TimeWait
	case "CLOSE_WAIT":
		return s.State.CloseWait
	case "SYN_SENT":
		return s.State.SynSent
	case "SYN_RECV":
		return s.State.SynRecv
	case "FIN_WAIT1":
		return s.State.FinWait1
	case "FIN_WAIT2":
		return s.State.FinWait2
	case "CLOSING":
		return s.State.Closing
	case "LAST_ACK":
		return s.State.LastAck
	case "CLOSED":
		return s.State.Closed
	default:
		return s.Normal
	}
}

// createDraculaTheme creates a theme based on the Dracula color scheme
func createDraculaTheme() *Theme {
	return &Theme{
		Name: "dracula",
		Styles: Styles{
			Header: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#BD93F9")),

			Watched: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFB86C")),

			Border: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6272A4")),

			Selected: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FF79C6")),

			Normal: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F8F8F2")),

			Error: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF5555")),

			Success: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#50FA7B")),

			Warning: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F1FA8C")),

			Footer: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6272A4")),

			Background: lipgloss.NewStyle().
				Background(lipgloss.Color("#282A36")),

			Proto: ProtoStyles{
				TCP:  lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")),
				UDP:  lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")),
				Unix: lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")),
				TCP6: lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")),
				UDP6: lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")),
			},

			State: StateStyles{
				Listen:      lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")),
				Established: lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD")),
				TimeWait:    lipgloss.NewStyle().Foreground(lipgloss.Color("#F1FA8C")),
				CloseWait:   lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C")),
				SynSent:     lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")),
				SynRecv:     lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")),
				FinWait1:    lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")),
				FinWait2:    lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")),
				Closing:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")),
				LastAck:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")),
				Closed:      lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")),
			},
		},
	}
}

// createGruvboxTheme creates a theme based on the Gruvbox Dark color scheme
func createGruvboxTheme() *Theme {
	return &Theme{
		Name: "gruvbox",
		Styles: Styles{
			Header: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FABD2F")),

			Watched: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FE8019")),

			Border: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#928374")),

			Selected: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FE8019")),

			Normal: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EBDBB2")),

			Error: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FB4934")),

			Success: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#B8BB26")),

			Warning: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FABD2F")),

			Footer: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#928374")),

			Background: lipgloss.NewStyle().
				Background(lipgloss.Color("#282828")),

			Proto: ProtoStyles{
				TCP:  lipgloss.NewStyle().Foreground(lipgloss.Color("#B8BB26")),
				UDP:  lipgloss.NewStyle().Foreground(lipgloss.Color("#D3869B")),
				Unix: lipgloss.NewStyle().Foreground(lipgloss.Color("#928374")),
				TCP6: lipgloss.NewStyle().Foreground(lipgloss.Color("#B8BB26")),
				UDP6: lipgloss.NewStyle().Foreground(lipgloss.Color("#D3869B")),
			},

			State: StateStyles{
				Listen:      lipgloss.NewStyle().Foreground(lipgloss.Color("#B8BB26")),
				Established: lipgloss.NewStyle().Foreground(lipgloss.Color("#83A598")),
				TimeWait:    lipgloss.NewStyle().Foreground(lipgloss.Color("#FABD2F")),
				CloseWait:   lipgloss.NewStyle().Foreground(lipgloss.Color("#FE8019")),
				SynSent:     lipgloss.NewStyle().Foreground(lipgloss.Color("#D3869B")),
				SynRecv:     lipgloss.NewStyle().Foreground(lipgloss.Color("#D3869B")),
				FinWait1:    lipgloss.NewStyle().Foreground(lipgloss.Color("#FB4934")),
				FinWait2:    lipgloss.NewStyle().Foreground(lipgloss.Color("#FB4934")),
				Closing:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FB4934")),
				LastAck:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FB4934")),
				Closed:      lipgloss.NewStyle().Foreground(lipgloss.Color("#928374")),
			},
		},
	}
}

// createGruvboxLightTheme creates a theme based on the Gruvbox Light color scheme
func createGruvboxLightTheme() *Theme {
	return &Theme{
		Name: "gruvbox-light",
		Styles: Styles{
			Header: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#D79921")),

			Watched: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#D65D0E")),

			Border: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#928374")),

			Selected: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#D65D0E")),

			Normal: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#3C3836")),

			Error: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#CC241D")),

			Success: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#98971A")),

			Warning: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#D79921")),

			Footer: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#928374")),

			Background: lipgloss.NewStyle().
				Background(lipgloss.Color("#FBF1C7")),

			Proto: ProtoStyles{
				TCP:  lipgloss.NewStyle().Foreground(lipgloss.Color("#98971A")),
				UDP:  lipgloss.NewStyle().Foreground(lipgloss.Color("#B16286")),
				Unix: lipgloss.NewStyle().Foreground(lipgloss.Color("#928374")),
				TCP6: lipgloss.NewStyle().Foreground(lipgloss.Color("#98971A")),
				UDP6: lipgloss.NewStyle().Foreground(lipgloss.Color("#B16286")),
			},

			State: StateStyles{
				Listen:      lipgloss.NewStyle().Foreground(lipgloss.Color("#98971A")),
				Established: lipgloss.NewStyle().Foreground(lipgloss.Color("#458588")),
				TimeWait:    lipgloss.NewStyle().Foreground(lipgloss.Color("#D79921")),
				CloseWait:   lipgloss.NewStyle().Foreground(lipgloss.Color("#D65D0E")),
				SynSent:     lipgloss.NewStyle().Foreground(lipgloss.Color("#B16286")),
				SynRecv:     lipgloss.NewStyle().Foreground(lipgloss.Color("#B16286")),
				FinWait1:    lipgloss.NewStyle().Foreground(lipgloss.Color("#CC241D")),
				FinWait2:    lipgloss.NewStyle().Foreground(lipgloss.Color("#CC241D")),
				Closing:     lipgloss.NewStyle().Foreground(lipgloss.Color("#CC241D")),
				LastAck:     lipgloss.NewStyle().Foreground(lipgloss.Color("#CC241D")),
				Closed:      lipgloss.NewStyle().Foreground(lipgloss.Color("#928374")),
			},
		},
	}
}

// createNordTheme creates a theme based on the Nord color scheme
func createNordTheme() *Theme {
	return &Theme{
		Name: "nord",
		Styles: Styles{
			Header: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#88C0D0")),

			Watched: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#D08770")),

			Border: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4C566A")),

			Selected: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#81A1C1")),

			Normal: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ECEFF4")),

			Error: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#BF616A")),

			Success: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A3BE8C")),

			Warning: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EBCB8B")),

			Footer: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4C566A")),

			Background: lipgloss.NewStyle().
				Background(lipgloss.Color("#2E3440")),

			Proto: ProtoStyles{
				TCP:  lipgloss.NewStyle().Foreground(lipgloss.Color("#A3BE8C")),
				UDP:  lipgloss.NewStyle().Foreground(lipgloss.Color("#B48EAD")),
				Unix: lipgloss.NewStyle().Foreground(lipgloss.Color("#4C566A")),
				TCP6: lipgloss.NewStyle().Foreground(lipgloss.Color("#A3BE8C")),
				UDP6: lipgloss.NewStyle().Foreground(lipgloss.Color("#B48EAD")),
			},

			State: StateStyles{
				Listen:      lipgloss.NewStyle().Foreground(lipgloss.Color("#A3BE8C")),
				Established: lipgloss.NewStyle().Foreground(lipgloss.Color("#88C0D0")),
				TimeWait:    lipgloss.NewStyle().Foreground(lipgloss.Color("#EBCB8B")),
				CloseWait:   lipgloss.NewStyle().Foreground(lipgloss.Color("#D08770")),
				SynSent:     lipgloss.NewStyle().Foreground(lipgloss.Color("#B48EAD")),
				SynRecv:     lipgloss.NewStyle().Foreground(lipgloss.Color("#B48EAD")),
				FinWait1:    lipgloss.NewStyle().Foreground(lipgloss.Color("#BF616A")),
				FinWait2:    lipgloss.NewStyle().Foreground(lipgloss.Color("#BF616A")),
				Closing:     lipgloss.NewStyle().Foreground(lipgloss.Color("#BF616A")),
				LastAck:     lipgloss.NewStyle().Foreground(lipgloss.Color("#BF616A")),
				Closed:      lipgloss.NewStyle().Foreground(lipgloss.Color("#4C566A")),
			},
		},
	}
}

// createCatppuccinLatteTheme creates a theme based on Catppuccin Latte (light)
func createCatppuccinLatteTheme() *Theme {
	return &Theme{
		Name: "catppuccin-latte",
		Styles: Styles{
			Header: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#1e66f5")),

			Watched: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#fe640b")),

			Border: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9ca0b0")),

			Selected: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#8839ef")),

			Normal: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4c4f69")),

			Error: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#d20f39")),

			Success: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#40a02b")),

			Warning: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#df8e1d")),

			Footer: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9ca0b0")),

			Background: lipgloss.NewStyle().
				Background(lipgloss.Color("#eff1f5")),

			Proto: ProtoStyles{
				TCP:  lipgloss.NewStyle().Foreground(lipgloss.Color("#40a02b")),
				UDP:  lipgloss.NewStyle().Foreground(lipgloss.Color("#8839ef")),
				Unix: lipgloss.NewStyle().Foreground(lipgloss.Color("#9ca0b0")),
				TCP6: lipgloss.NewStyle().Foreground(lipgloss.Color("#40a02b")),
				UDP6: lipgloss.NewStyle().Foreground(lipgloss.Color("#8839ef")),
			},

			State: StateStyles{
				Listen:      lipgloss.NewStyle().Foreground(lipgloss.Color("#40a02b")),
				Established: lipgloss.NewStyle().Foreground(lipgloss.Color("#1e66f5")),
				TimeWait:    lipgloss.NewStyle().Foreground(lipgloss.Color("#df8e1d")),
				CloseWait:   lipgloss.NewStyle().Foreground(lipgloss.Color("#fe640b")),
				SynSent:     lipgloss.NewStyle().Foreground(lipgloss.Color("#8839ef")),
				SynRecv:     lipgloss.NewStyle().Foreground(lipgloss.Color("#8839ef")),
				FinWait1:    lipgloss.NewStyle().Foreground(lipgloss.Color("#d20f39")),
				FinWait2:    lipgloss.NewStyle().Foreground(lipgloss.Color("#d20f39")),
				Closing:     lipgloss.NewStyle().Foreground(lipgloss.Color("#d20f39")),
				LastAck:     lipgloss.NewStyle().Foreground(lipgloss.Color("#d20f39")),
				Closed:      lipgloss.NewStyle().Foreground(lipgloss.Color("#9ca0b0")),
			},
		},
	}
}

// createCatppuccinFrappeTheme creates a theme based on Catppuccin Frappé
func createCatppuccinFrappeTheme() *Theme {
	return &Theme{
		Name: "catppuccin-frappe",
		Styles: Styles{
			Header: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#8caaee")),

			Watched: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#ef9f76")),

			Border: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626880")),

			Selected: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#ca9ee6")),

			Normal: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#c6d0f5")),

			Error: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#e78284")),

			Success: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#a6d189")),

			Warning: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#e5c890")),

			Footer: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626880")),

			Background: lipgloss.NewStyle().
				Background(lipgloss.Color("#303446")),

			Proto: ProtoStyles{
				TCP:  lipgloss.NewStyle().Foreground(lipgloss.Color("#a6d189")),
				UDP:  lipgloss.NewStyle().Foreground(lipgloss.Color("#ca9ee6")),
				Unix: lipgloss.NewStyle().Foreground(lipgloss.Color("#626880")),
				TCP6: lipgloss.NewStyle().Foreground(lipgloss.Color("#a6d189")),
				UDP6: lipgloss.NewStyle().Foreground(lipgloss.Color("#ca9ee6")),
			},

			State: StateStyles{
				Listen:      lipgloss.NewStyle().Foreground(lipgloss.Color("#a6d189")),
				Established: lipgloss.NewStyle().Foreground(lipgloss.Color("#8caaee")),
				TimeWait:    lipgloss.NewStyle().Foreground(lipgloss.Color("#e5c890")),
				CloseWait:   lipgloss.NewStyle().Foreground(lipgloss.Color("#ef9f76")),
				SynSent:     lipgloss.NewStyle().Foreground(lipgloss.Color("#ca9ee6")),
				SynRecv:     lipgloss.NewStyle().Foreground(lipgloss.Color("#ca9ee6")),
				FinWait1:    lipgloss.NewStyle().Foreground(lipgloss.Color("#e78284")),
				FinWait2:    lipgloss.NewStyle().Foreground(lipgloss.Color("#e78284")),
				Closing:     lipgloss.NewStyle().Foreground(lipgloss.Color("#e78284")),
				LastAck:     lipgloss.NewStyle().Foreground(lipgloss.Color("#e78284")),
				Closed:      lipgloss.NewStyle().Foreground(lipgloss.Color("#626880")),
			},
		},
	}
}

// createCatppuccinMacchiatoTheme creates a theme based on Catppuccin Macchiato
func createCatppuccinMacchiatoTheme() *Theme {
	return &Theme{
		Name: "catppuccin-macchiato",
		Styles: Styles{
			Header: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#8aadf4")),

			Watched: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#f5a97f")),

			Border: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#5b6078")),

			Selected: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#c6a0f6")),

			Normal: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#cad3f5")),

			Error: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ed8796")),

			Success: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#a6da95")),

			Warning: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#eed49f")),

			Footer: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#5b6078")),

			Background: lipgloss.NewStyle().
				Background(lipgloss.Color("#24273a")),

			Proto: ProtoStyles{
				TCP:  lipgloss.NewStyle().Foreground(lipgloss.Color("#a6da95")),
				UDP:  lipgloss.NewStyle().Foreground(lipgloss.Color("#c6a0f6")),
				Unix: lipgloss.NewStyle().Foreground(lipgloss.Color("#5b6078")),
				TCP6: lipgloss.NewStyle().Foreground(lipgloss.Color("#a6da95")),
				UDP6: lipgloss.NewStyle().Foreground(lipgloss.Color("#c6a0f6")),
			},

			State: StateStyles{
				Listen:      lipgloss.NewStyle().Foreground(lipgloss.Color("#a6da95")),
				Established: lipgloss.NewStyle().Foreground(lipgloss.Color("#8aadf4")),
				TimeWait:    lipgloss.NewStyle().Foreground(lipgloss.Color("#eed49f")),
				CloseWait:   lipgloss.NewStyle().Foreground(lipgloss.Color("#f5a97f")),
				SynSent:     lipgloss.NewStyle().Foreground(lipgloss.Color("#c6a0f6")),
				SynRecv:     lipgloss.NewStyle().Foreground(lipgloss.Color("#c6a0f6")),
				FinWait1:    lipgloss.NewStyle().Foreground(lipgloss.Color("#ed8796")),
				FinWait2:    lipgloss.NewStyle().Foreground(lipgloss.Color("#ed8796")),
				Closing:     lipgloss.NewStyle().Foreground(lipgloss.Color("#ed8796")),
				LastAck:     lipgloss.NewStyle().Foreground(lipgloss.Color("#ed8796")),
				Closed:      lipgloss.NewStyle().Foreground(lipgloss.Color("#5b6078")),
			},
		},
	}
}

// createCatppuccinMochaTheme creates a theme based on Catppuccin Mocha (dark)
func createCatppuccinMochaTheme() *Theme {
	return &Theme{
		Name: "catppuccin-mocha",
		Styles: Styles{
			Header: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#89b4fa")),

			Watched: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#fab387")),

			Border: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#585b70")),

			Selected: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#cba6f7")),

			Normal: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#cdd6f4")),

			Error: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#f38ba8")),

			Success: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#a6e3a1")),

			Warning: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#f9e2af")),

			Footer: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#585b70")),

			Background: lipgloss.NewStyle().
				Background(lipgloss.Color("#1e1e2e")),

			Proto: ProtoStyles{
				TCP:  lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")),
				UDP:  lipgloss.NewStyle().Foreground(lipgloss.Color("#cba6f7")),
				Unix: lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")),
				TCP6: lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")),
				UDP6: lipgloss.NewStyle().Foreground(lipgloss.Color("#cba6f7")),
			},

			State: StateStyles{
				Listen:      lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")),
				Established: lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa")),
				TimeWait:    lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af")),
				CloseWait:   lipgloss.NewStyle().Foreground(lipgloss.Color("#fab387")),
				SynSent:     lipgloss.NewStyle().Foreground(lipgloss.Color("#cba6f7")),
				SynRecv:     lipgloss.NewStyle().Foreground(lipgloss.Color("#cba6f7")),
				FinWait1:    lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")),
				FinWait2:    lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")),
				Closing:     lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")),
				LastAck:     lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8")),
				Closed:      lipgloss.NewStyle().Foreground(lipgloss.Color("#585b70")),
			},
		},
	}
}
