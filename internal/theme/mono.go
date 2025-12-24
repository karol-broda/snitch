package theme

import "github.com/charmbracelet/lipgloss"

// createMonoTheme creates a monochrome theme (no colors)
// useful for accessibility, piping output, or minimal terminals
func createMonoTheme() *Theme {
	baseStyle := lipgloss.NewStyle()
	boldStyle := lipgloss.NewStyle().Bold(true)

	return &Theme{
		Name: "mono",
		Styles: Styles{
			Header:     boldStyle,
			Border:     baseStyle,
			Selected:   boldStyle,
			Watched:    boldStyle,
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

