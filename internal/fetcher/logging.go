package fetcher

import (
	"fmt"
	"io"
	"strings"

	"aibomgen-cra/internal/ui"
)

var logWriter io.Writer

// SetLogger sets an optional destination for fetch logs.
func SetLogger(w io.Writer) { logWriter = w }

func logf(modelID string, format string, args ...any) {
	if logWriter == nil {
		return
	}
	m := strings.TrimSpace(modelID)
	if m == "" {
		m = "(unknown)"
	}
	prefix := ui.Color("Fetch:", ui.FgMagenta)
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(logWriter, "%s model=%s %s\n", prefix, m, msg)
}
