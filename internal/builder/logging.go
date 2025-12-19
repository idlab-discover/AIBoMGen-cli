package builder

import (
	"io"

	"aibomgen-cra/internal/logging"
	"aibomgen-cra/internal/ui"
)

var logger = &logging.Logger{PrefixText: "Build:", PrefixColor: ui.FgGreen}

// SetLogger sets an optional destination for builder logs.
func SetLogger(w io.Writer) { logger.SetWriter(w) }

func logf(modelID string, format string, args ...any) {
	logger.Logf(modelID, format, args...)
}
