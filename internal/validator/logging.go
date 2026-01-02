package validator

import (
	"io"

	"github.com/idlab-discover/AIBoMGen-cli/internal/logging"
	"github.com/idlab-discover/AIBoMGen-cli/internal/ui"
)

var logger = &logging.Logger{PrefixText: "Validation Report:", PrefixColor: ui.FgCyan, OmitModel: true}

// SetLogger sets an optional destination for validator logs.
func SetLogger(w io.Writer) { logger.SetWriter(w) }

func logf(format string, args ...any) {
	logger.Logf("", format, args...)
}
