package generator

import (
	"io"

	"aibomgen-cra/internal/logging"
	"aibomgen-cra/internal/ui"
)

var logger = &logging.Logger{PrefixText: "Generator:", PrefixColor: ui.FgCyan}

// SetLogger sets an optional destination for metadata logs.
func SetLogger(w io.Writer) { logger.SetWriter(w) }

func logf(modelID string, format string, args ...any) {
	logger.Logf(modelID, format, args...)
}
