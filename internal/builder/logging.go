package builder

import (
	"io"

	"github.com/idlab-discover/AIBoMGen-cli/internal/logging"
	"github.com/idlab-discover/AIBoMGen-cli/internal/ui"
)

var logger = &logging.Logger{PrefixText: "Build:", PrefixColor: ui.FgGreen}

// SetLogger sets an optional destination for builder logs.
func SetLogger(w io.Writer) { logger.SetWriter(w) }

func logf(modelID string, format string, args ...any) {
	logger.Logf(modelID, format, args...)
}
