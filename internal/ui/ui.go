package ui

// Enabled controls whether color codes are applied.
var Enabled = true

// Basic ANSI color codes.
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	FgCyan    = "\033[36m"
	FgGreen   = "\033[32m"
	FgMagenta = "\033[35m"
	FgYellow  = "\033[33m"
	FgRed     = "\033[31m"
)

// Init sets Enabled based solely on the provided flag.
func Init(noColor bool) {
	if noColor {
		Enabled = false
	} else {
		Enabled = true
	}
}

// Color wraps a string with the given ANSI code when Enabled.
func Color(s string, code string) string {
	if !Enabled {
		return s
	}
	return code + s + Reset
}
