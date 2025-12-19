package completeness

import (
	"strings"

	"aibomgen-cra/internal/metadata"
)

// PrintReport writes the report to the configured logger writer.
// If no logger writer is configured, it produces no output.
func PrintReport(r Report) {
	logf("score=%.1f%% (%d/%d)", r.Score*100, r.Passed, r.Total)

	if len(r.MissingRequired) > 0 {
		logf("missing required: %s", joinKeys(r.MissingRequired))
	}
	if len(r.MissingOptional) > 0 {
		logf("missing optional: %s", joinKeys(r.MissingOptional))
	}
}
func joinKeys(keys []metadata.Key) string {
	if len(keys) == 0 {
		return ""
	}
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(k.String())
	}
	return b.String()
}
