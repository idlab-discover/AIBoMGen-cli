package metadata

import (
	"fmt"
	"io"
	"strings"

	"aibomgen-cra/internal/logging"
	"aibomgen-cra/internal/ui"
)

var logger = &logging.Logger{PrefixText: "Meta:", PrefixColor: ui.FgRed}

// SetLogger sets an optional destination for metadata logs.
func SetLogger(w io.Writer) { logger.SetWriter(w) }

func logf(modelID string, format string, args ...any) {
	logger.Logf(modelID, format, args...)
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
