package completeness

import (
	"io"

	"github.com/idlab-discover/AIBoMGen-cli/internal/logging"
	"github.com/idlab-discover/AIBoMGen-cli/internal/ui"
)

var logger = &logging.Logger{PrefixText: "Completeness Report:", PrefixColor: ui.FgYellow, OmitModel: true}

// SetLogger sets an optional destination for completeness output/logs.
// When set to nil, completeness output/logs are disabled.
func SetLogger(w io.Writer) { logger.SetWriter(w) }

func logf(format string, args ...any) {
	logger.Logf("", format, args...)
}
