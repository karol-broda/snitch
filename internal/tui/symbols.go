package tui

// unicode symbols used throughout the TUI
const (
	// indicators
	SymbolSelected   = string('\u25B8') // black right-pointing small triangle
	SymbolWatched    = string('\u2605') // black star
	SymbolWarning    = string('\u26A0') // warning sign
	SymbolSuccess    = string('\u2713') // check mark
	SymbolError      = string('\u2717') // ballot x
	SymbolBullet     = string('\u2022') // bullet
	SymbolArrowRight = string('\u2192') // rightwards arrow
	SymbolArrowLeft  = string('\u2190') // leftwards arrow
	SymbolArrowUp    = string('\u2191') // upwards arrow
	SymbolArrowDown  = string('\u2193') // downwards arrow
	SymbolRefresh    = string('\u21BB') // clockwise open circle arrow
	SymbolEllipsis   = string('\u2026') // horizontal ellipsis
	SymbolDownload   = string('\u21E9') // downwards white arrow

	// box drawing rounded 
	BoxTopLeft     = string('\u256D') // light arc down and right
	BoxTopRight    = string('\u256E') // light arc down and left
	BoxBottomLeft  = string('\u2570') // light arc up and right
	BoxBottomRight = string('\u256F') // light arc up and left
	BoxHorizontal  = string('\u2500') // light horizontal
	BoxVertical    = string('\u2502') // light vertical

	// box drawing connectors
	BoxTeeDown  = string('\u252C') // light down and horizontal
	BoxTeeUp    = string('\u2534') // light up and horizontal
	BoxTeeRight = string('\u251C') // light vertical and right
	BoxTeeLeft  = string('\u2524') // light vertical and left
	BoxCross    = string('\u253C') // light vertical and horizontal

	// misc
	SymbolDash   = string('\u2013') // en dash
	SymbolExport = string('\u21E5') // rightwards arrow to bar
	SymbolPrompt = string('\u276F') // heavy right-pointing angle quotation mark ornament
)

