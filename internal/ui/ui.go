package ui

// Basic ANSI color codes (legacy - used by logging package).
// New code should use lipgloss styles from styles.go instead.
const (
	Reset = "\033[0m"
	// LegacyBold is the raw ANSI code for bold text
	LegacyBold = "\033[1m"
	FgCyan     = "\033[36m"
	FgGreen    = "\033[32m"
	FgMagenta  = "\033[35m"
	FgYellow   = "\033[33m"
	FgRed      = "\033[31m"
)

// Color wraps a string with the given ANSI code.
// Deprecated: Use lipgloss styles from styles.go instead.
func Color(s string, code string) string {
	return code + s + Reset
}
