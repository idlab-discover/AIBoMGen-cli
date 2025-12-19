package fetcher

import (
	"io"

	"aibomgen-cra/internal/logging"
	"aibomgen-cra/internal/ui"
)

var logger = &logging.Logger{PrefixText: "Fetch:", PrefixColor: ui.FgMagenta}

// SetLogger sets an optional destination for fetch logs.
func SetLogger(w io.Writer) { logger.SetWriter(w) }

func logf(modelID string, format string, args ...any) {
	logger.Logf(modelID, format, args...)
}
