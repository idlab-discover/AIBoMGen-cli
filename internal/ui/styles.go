package ui

import "github.com/charmbracelet/lipgloss"

// Color palette for the application
var (
	// Primary colors
	ColorPrimary   = lipgloss.Color("#7C3AED") // Purple
	ColorSecondary = lipgloss.Color("#06B6D4") // Cyan
	ColorSuccess   = lipgloss.Color("#10B981") // Green
	ColorWarning   = lipgloss.Color("#F59E0B") // Amber
	ColorError     = lipgloss.Color("#EF4444") // Red
	ColorMuted     = lipgloss.Color("#6B7280") // Gray
	ColorHighlight = lipgloss.Color("#8B5CF6") // Light purple

	// Text colors
	ColorText     = lipgloss.Color("#F9FAFB") // White
	ColorTextDim  = lipgloss.Color("#9CA3AF") // Light gray
	ColorTextMute = lipgloss.Color("#6B7280") // Muted gray
)

// styleWrapper wraps a lipgloss style
type styleWrapper struct {
	style lipgloss.Style
}

// Render renders the string with the style
func (s styleWrapper) Render(str string) string {
	return s.style.Render(str)
}

// Bold returns a new style with bold enabled
func (s styleWrapper) Bold(v bool) styleWrapper {
	return styleWrapper{s.style.Bold(v)}
}

// Text styles using lipgloss
var (
	// Bold text
	Bold = styleWrapper{lipgloss.NewStyle().Bold(true)}

	// Dimmed text for secondary information
	Dim = styleWrapper{lipgloss.NewStyle().Foreground(ColorTextDim)}

	// Muted text for hints
	Muted = styleWrapper{lipgloss.NewStyle().Foreground(ColorTextMute)}

	// Success text (green)
	Success = styleWrapper{lipgloss.NewStyle().Foreground(ColorSuccess)}

	// Warning text (amber)
	Warning = styleWrapper{lipgloss.NewStyle().Foreground(ColorWarning)}

	// Error text (red)
	Error = styleWrapper{lipgloss.NewStyle().Foreground(ColorError)}

	// Primary accent text (purple)
	Primary = styleWrapper{lipgloss.NewStyle().Foreground(ColorPrimary)}

	// Secondary accent text (cyan)
	Secondary = styleWrapper{lipgloss.NewStyle().Foreground(ColorSecondary)}

	// Highlight text
	Highlight = styleWrapper{lipgloss.NewStyle().Foreground(ColorHighlight).Bold(true)}
)

// Status indicators
var (
	// CheckMark returns a styled check mark
	CheckMark = func() string { return Success.Render("✓") }()

	// CrossMark returns a styled cross mark
	CrossMark = func() string { return Error.Render("✗") }()

	// WarnMark returns a styled warning mark
	WarnMark = func() string { return Warning.Render("⚠") }()

	// InfoMark returns a styled info mark
	InfoMark = func() string { return Secondary.Render("ℹ") }()

	// Bullet returns a styled bullet point
	Bullet = func() string { return Muted.Render("•") }()
)

// GetCheckMark returns a check mark respecting current color settings
func GetCheckMark() string { return Success.Render("✓") }

// GetCrossMark returns a cross mark respecting current color settings
func GetCrossMark() string { return Error.Render("✗") }

// GetWarnMark returns a warning mark respecting current color settings
func GetWarnMark() string { return Warning.Render("⚠") }

// GetInfoMark returns an info mark respecting current color settings
func GetInfoMark() string { return Secondary.Render("ℹ") }

// GetBullet returns a bullet point respecting current color settings
func GetBullet() string { return Muted.Render("•") }

// Box styles for panels and containers
type boxWrapper struct {
	style lipgloss.Style
}

func (b boxWrapper) Render(str string) string {
	return b.style.Render(str)
}

var (
	// Standard box with border
	Box = boxWrapper{lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorMuted).
		Padding(0, 1)}

	// Highlighted box
	HighlightBox = boxWrapper{lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1)}

	// Success box
	SuccessBox = boxWrapper{lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSuccess).
			Padding(0, 1)}

	// Error box
	ErrorBox = boxWrapper{lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorError).
			Padding(0, 1)}
)

// Header styles
var (
	// Main title style
	Title = styleWrapper{lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)}

	// Subtitle style
	Subtitle = styleWrapper{lipgloss.NewStyle().
			Foreground(ColorTextDim).
			Italic(true)}

	// Section header
	SectionHeader = styleWrapper{lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true)}
)

// Progress bar styles
var (
	// Progress bar filled portion
	ProgressFilled = styleWrapper{lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Background(ColorSuccess)}

	// Progress bar empty portion
	ProgressEmpty = styleWrapper{lipgloss.NewStyle().
			Foreground(ColorMuted).
			Background(lipgloss.Color("#374151"))}
)

// Step status styles
var (
	// Pending step (not started)
	StepPending = styleWrapper{lipgloss.NewStyle().Foreground(ColorMuted)}

	// Running step (in progress)
	StepRunning = styleWrapper{lipgloss.NewStyle().Foreground(ColorSecondary)}

	// Completed step
	StepComplete = styleWrapper{lipgloss.NewStyle().Foreground(ColorSuccess)}

	// Failed step
	StepFailed = styleWrapper{lipgloss.NewStyle().Foreground(ColorError)}

	// Skipped step
	StepSkipped = styleWrapper{lipgloss.NewStyle().Foreground(ColorWarning)}
)

// StyledText applies a lipgloss style to a string
func StyledText(s string, style lipgloss.Style) string {
	return style.Render(s)
}

// FormatKeyValue formats a key-value pair with styling
func FormatKeyValue(key, value string) string {
	return Dim.Render(key+": ") + value
}

// FormatStatus formats a status message with an appropriate icon
func FormatStatus(status, message string) string {
	var icon string
	switch status {
	case "success":
		icon = CheckMark
	case "error":
		icon = CrossMark
	case "warning":
		icon = WarnMark
	case "info":
		icon = InfoMark
	default:
		icon = Bullet
	}
	return icon + " " + message
}
