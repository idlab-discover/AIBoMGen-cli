package logging

import (
	"fmt"
	"io"
	"strings"

	"aibomgen-cra/internal/ui"
)

// Logger is a tiny opt-in logger used across internal packages.
// When Writer is nil, logging is disabled.
//
// The output format is:
//
//	<ColoredPrefix> model=<modelID> <formattedMessage>\n
//
// where <modelID> is trimmed and defaults to "(unknown)".
type Logger struct {
	Writer io.Writer

	PrefixText  string
	PrefixColor string

	// OmitModel controls whether the model ID field is written.
	// When false (default), output includes: "model=<id>".
	OmitModel bool
}

func (l *Logger) SetWriter(w io.Writer) { l.Writer = w }

func (l *Logger) Enabled() bool { return l != nil && l.Writer != nil }

func (l *Logger) Logf(modelID string, format string, args ...any) {
	if l == nil || l.Writer == nil {
		return
	}
	prefix := l.PrefixText
	if prefix == "" {
		prefix = "Log:"
	}
	if l.PrefixColor != "" {
		prefix = ui.Color(prefix, l.PrefixColor)
	}
	msg := fmt.Sprintf(format, args...)
	if l.OmitModel {
		fmt.Fprintf(l.Writer, "%s %s\n", prefix, msg)
		return
	}

	m := strings.TrimSpace(modelID)
	if m == "" {
		m = "(unknown)"
	}
	fmt.Fprintf(l.Writer, "%s model=%s %s\n", prefix, m, msg)
}
