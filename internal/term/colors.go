package term

import "os"

// Hex color constants — single source of truth for the entire app.
const (
	HexPrimary   = "#2DD4BF" // Teal
	HexSecondary = "#5EEAD4" // Light Teal
	HexMuted     = "#78716C" // Gray
	HexSuccess   = "#4ADE80" // Green
	HexDanger    = "#FF6B6B" // Red
	HexWarning   = "#FBBF24" // Yellow
)

// ANSI color codes — disabled when NO_COLOR is set.
var (
	Bold  = "\033[1m"
	Dim   = "\033[2m"
	Reset = "\033[0m"

	// Derived from hex palette
	Primary = "\033[38;2;45;212;191m"
	Muted   = "\033[38;2;120;113;108m"
	Green   = "\033[38;2;74;222;128m"
	Red     = "\033[38;2;255;107;107m"
	Yellow  = "\033[38;2;251;191;36m"
)

func init() {
	if os.Getenv("NO_COLOR") != "" {
		Bold = ""
		Dim = ""
		Reset = ""
		Primary = ""
		Muted = ""
		Green = ""
		Red = ""
		Yellow = ""
	}
}
