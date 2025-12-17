package enricher

import (
	"fmt"
	"io"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

// InteractiveCompleteBOM is temporarily not implemented.
func InteractiveCompleteBOM(_ *cdx.BOM, _ bool, _ io.Reader, _ io.Writer) (err error) {
	return fmt.Errorf("model README fetch not implemented")
}
