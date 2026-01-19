package validator

import (
	"fmt"

	"github.com/idlab-discover/AIBoMGen-cli/internal/ui"
)

// ToUIReport converts a validator.ValidationResult to a ui.ValidationReport
// This avoids circular import issues
func (r ValidationResult) ToUIReport(modelID string) ui.ValidationReport {
	// Convert missing required fields
	missingReq := make([]ui.FieldKey, len(r.MissingRequired))
	for i, k := range r.MissingRequired {
		missingReq[i] = k
	}

	// Convert missing optional fields
	missingOpt := make([]ui.FieldKey, len(r.MissingOptional))
	for i, k := range r.MissingOptional {
		missingOpt[i] = k
	}

	// Convert dataset results
	datasetResults := make(map[string]ui.DatasetValidationResult)
	for name, dr := range r.DatasetResults {
		missingReqDS := make([]ui.FieldKey, len(dr.MissingRequired))
		for i, k := range dr.MissingRequired {
			missingReqDS[i] = k
		}

		missingOptDS := make([]ui.FieldKey, len(dr.MissingOptional))
		for i, k := range dr.MissingOptional {
			missingOptDS[i] = k
		}

		datasetResults[name] = ui.DatasetValidationResult{
			DatasetRef:        dr.DatasetRef,
			CompletenessScore: dr.CompletenessScore,
			MissingRequired:   missingReqDS,
			MissingOptional:   missingOptDS,
			Errors:            dr.Errors,
			Warnings:          dr.Warnings,
		}
	}

	return ui.ValidationReport{
		ModelID:           modelID,
		Valid:             r.Valid,
		Errors:            r.Errors,
		Warnings:          r.Warnings,
		CompletenessScore: r.CompletenessScore,
		MissingRequired:   missingReq,
		MissingOptional:   missingOpt,
		DatasetResults:    datasetResults,
	}
}

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
