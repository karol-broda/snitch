package theme

// ANSI palette uses standard terminal colors (0-15)
// this allows the theme to inherit from the user's terminal color scheme
var paletteANSI = Palette{
	Name: "ansi",

	Fg:       "15", // bright white
	FgMuted:  "7",  // white
	FgSubtle: "8",  // bright black (gray)
	Bg:       "0",  // black
	BgMuted:  "0",  // black
	Border:   "8",  // bright black (gray)

	Red:     "1",  // red
	Green:   "2",  // green
	Yellow:  "3",  // yellow
	Blue:    "4",  // blue
	Magenta: "5",  // magenta
	Cyan:    "6",  // cyan
	Orange:  "3",  // yellow (ansi has no orange, fallback to yellow)
	Gray:    "8",  // bright black
}

