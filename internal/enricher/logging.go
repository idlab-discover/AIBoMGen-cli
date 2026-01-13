package enricher

import (
	"io"

	"github.com/idlab-discover/AIBoMGen-cli/internal/logging"
	"github.com/idlab-discover/AIBoMGen-cli/internal/ui"
)

var logger = &logging.Logger{PrefixText: "Enrich:", PrefixColor: ui.FgRed}

// SetLogger sets an optional destination for fetch logs.
func SetLogger(w io.Writer) { logger.SetWriter(w) }

func logf(modelID string, format string, args ...any) {
	logger.Logf(modelID, format, args...)
}
