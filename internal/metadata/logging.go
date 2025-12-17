package metadata

import (
	"fmt"
	"io"
	"strings"

	"aibomgen-cra/internal/ui"
)

var logWriter io.Writer

// SetLogger sets an optional destination for metadata logs.
func SetLogger(w io.Writer) { logWriter = w }

func logf(modelID string, format string, args ...any) {
	if logWriter == nil {
		return
	}
	m := strings.TrimSpace(modelID)
	if m == "" {
		m = "(unknown)"
	}
	prefix := ui.Color("Meta:", ui.FgRed)
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(logWriter, "%s model=%s %s\n", prefix, m, msg)
}

func summarizeValue(v any) string {
	if v == nil {
		return "<nil>"
	}
	switch t := v.(type) {
	case string:
		s := strings.TrimSpace(t)
		if len(s) > 80 {
			s = s[:77] + "..."
		}
		return fmt.Sprintf("%q", s)
	case []string:
		return fmt.Sprintf("[]string(len=%d)", len(t))
	case map[string]any:
		return fmt.Sprintf("map(len=%d)", len(t))
	default:
		// avoid dumping huge structs; type is usually enough
		return fmt.Sprintf("%T", v)
	}
}
