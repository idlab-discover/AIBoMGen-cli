package completeness

import (
	"strings"

	"github.com/idlab-discover/AIBoMGen-cli/internal/metadata"
	"github.com/idlab-discover/AIBoMGen-cli/internal/ui"
)

// ToUIReport converts a completeness.Report to a ui.CompletenessReport
// This avoids circular import issues
func (r Report) ToUIReport() ui.CompletenessReport {
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

	// Convert dataset reports
	datasetReports := make(map[string]ui.DatasetReport)
	for name, dr := range r.DatasetReports {
		missingReqDS := make([]ui.FieldKey, len(dr.MissingRequired))
		for i, k := range dr.MissingRequired {
			missingReqDS[i] = k
		}

		missingOptDS := make([]ui.FieldKey, len(dr.MissingOptional))
		for i, k := range dr.MissingOptional {
			missingOptDS[i] = k
		}

		datasetReports[name] = ui.DatasetReport{
			DatasetRef:      dr.DatasetRef,
			Score:           dr.Score,
			Passed:          dr.Passed,
			Total:           dr.Total,
			MissingRequired: missingReqDS,
			MissingOptional: missingOptDS,
		}
	}

	return ui.CompletenessReport{
		ModelID:         r.ModelID,
		Score:           r.Score,
		Passed:          r.Passed,
		Total:           r.Total,
		MissingRequired: missingReq,
		MissingOptional: missingOpt,
		DatasetReports:  datasetReports,
	}
}

// PrintReport writes the report to the configured logger writer.
// If no logger writer is configured, it produces no output.
func PrintReport(r Report) {
	logf("Model score=%.1f%% (%d/%d)", r.Score*100, r.Passed, r.Total)

	if len(r.MissingRequired) > 0 {
		logf("missing required: %s", joinKeys(r.MissingRequired))
	}
	if len(r.MissingOptional) > 0 {
		logf("missing optional: %s", joinKeys(r.MissingOptional))
	}

	// Print dataset reports if any
	if len(r.DatasetReports) > 0 {
		logf("")
		logf("Dataset Components:")
		for dsName, dsReport := range r.DatasetReports {
			logf("  %s: score=%.1f%% (%d/%d)", dsName, dsReport.Score*100, dsReport.Passed, dsReport.Total)
			if len(dsReport.MissingRequired) > 0 {
				logf("    missing required: %s", joinDatasetKeys(dsReport.MissingRequired))
			}
			if len(dsReport.MissingOptional) > 0 {
				logf("    missing optional: %s", joinDatasetKeys(dsReport.MissingOptional))
			}
		}
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

func joinDatasetKeys(keys []metadata.DatasetKey) string {
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
