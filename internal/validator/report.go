package validator

import "fmt"

// PrintReport writes the validation report to the configured logger writer.
// If no logger writer is configured, it produces no output.
func PrintReport(r ValidationResult) {
	if r.Valid {
		logf("✅ validation passed")
	} else {
		logf("❌ validation failed")
	}

	if len(r.Errors) > 0 {
		logf("errors (%d):", len(r.Errors))
		for _, err := range r.Errors {
			logf("  • %s", err)
		}
	}

	if len(r.Warnings) > 0 {
		logf("warnings (%d):", len(r.Warnings))
		for _, warn := range r.Warnings {
			logf("  • %s", warn)
		}
	}

	logf("model completeness score: %.1f%% (%d required, %d optional missing)",
		r.CompletenessScore*100,
		len(r.MissingRequired),
		len(r.MissingOptional))

	// Print dataset validation results if any
	if len(r.DatasetResults) > 0 {
		logf("")
		logf("dataset validation:")
		for dsName, dsResult := range r.DatasetResults {
			logf("  %s: %.1f%% (%d required, %d optional missing)",
				dsName,
				dsResult.CompletenessScore*100,
				len(dsResult.MissingRequired),
				len(dsResult.MissingOptional))

			if len(dsResult.Errors) > 0 {
				for _, err := range dsResult.Errors {
					logf("    • error: %s", err)
				}
			}
			if len(dsResult.Warnings) > 0 {
				for _, warn := range dsResult.Warnings {
					logf("    • warning: %s", warn)
				}
			}
		}
	}
}

// FormatSummary returns a formatted summary string of the validation result.
// This is useful for command output.
func FormatSummary(r ValidationResult) string {
	status := "✅ PASSED"
	if !r.Valid {
		status = "❌ FAILED"
	}
	datasetCount := len(r.DatasetResults)
	if datasetCount > 0 {
		return fmt.Sprintf("Validation: %s | Model Score: %.1f%% | Datasets: %d | Errors: %d | Warnings: %d",
			status,
			r.CompletenessScore*100,
			datasetCount,
			len(r.Errors),
			len(r.Warnings))
	}
	return fmt.Sprintf("Validation: %s | Score: %.1f%% | Errors: %d | Warnings: %d",
		status,
		r.CompletenessScore*100,
		len(r.Errors),
		len(r.Warnings))
}
