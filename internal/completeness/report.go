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

	if len(r.MissingRequired) > 0 {
	}
	if len(r.MissingOptional) > 0 {
	}

	// Print dataset reports if any
	if len(r.DatasetReports) > 0 {
		for _, dsReport := range r.DatasetReports {
			if len(dsReport.MissingRequired) > 0 {
			}
			if len(dsReport.MissingOptional) > 0 {
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
