package generator

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
	prefix := ui.Color("Generator:", ui.FgCyan)
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(logWriter, "%s model=%s %s\n", prefix, m, msg)
}
