package theme

import (
	"sort"
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

var themes map[string]*Theme

func init() {
	themes = make(map[string]*Theme)

	// ansi theme (default) - inherits from terminal colors
	themes["ansi"] = paletteANSI.ToTheme()

	// catppuccin variants
	themes["catppuccin-mocha"] = paletteCatppuccinMocha.ToTheme()
	themes["catppuccin-macchiato"] = paletteCatppuccinMacchiato.ToTheme()
	themes["catppuccin-frappe"] = paletteCatppuccinFrappe.ToTheme()
	themes["catppuccin-latte"] = paletteCatppuccinLatte.ToTheme()

	// gruvbox variants
	themes["gruvbox-dark"] = paletteGruvboxDark.ToTheme()
	themes["gruvbox-light"] = paletteGruvboxLight.ToTheme()

	// dracula
	themes["dracula"] = paletteDracula.ToTheme()

	// nord
	themes["nord"] = paletteNord.ToTheme()

	// tokyo night variants
	themes["tokyo-night"] = paletteTokyoNight.ToTheme()
	themes["tokyo-night-storm"] = paletteTokyoNightStorm.ToTheme()
	themes["tokyo-night-light"] = paletteTokyoNightLight.ToTheme()

	// solarized variants
	themes["solarized-dark"] = paletteSolarizedDark.ToTheme()
	themes["solarized-light"] = paletteSolarizedLight.ToTheme()

	// one dark
	themes["one-dark"] = paletteOneDark.ToTheme()

	// monochrome (no colors)
	themes["mono"] = createMonoTheme()
}

// DefaultTheme is the theme used when none is specified
const DefaultTheme = "ansi"

// GetTheme returns a theme by name
func GetTheme(name string) *Theme {
	if name == "" || name == "auto" || name == "default" {
		return themes[DefaultTheme]
	}

	if theme, exists := themes[name]; exists {
		return theme
	}

	// fallback to default
	return themes[DefaultTheme]
}

// ListThemes returns available theme names sorted alphabetically
func ListThemes() []string {
	names := make([]string, 0, len(themes))
	for name := range themes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
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
