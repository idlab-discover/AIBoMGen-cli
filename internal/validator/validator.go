package validator

import (
	"fmt"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/idlab-discover/AIBoMGen-cli/internal/completeness"
	"github.com/idlab-discover/AIBoMGen-cli/internal/metadata"
)

type ValidationResult struct {
	Valid    bool
	Errors   []string
	Warnings []string

	// AIBOM-specific metrics
	CompletenessScore float64
	MissingRequired   []metadata.Key
	MissingOptional   []metadata.Key
}

type ValidationOptions struct {
	StrictMode           bool    // Fail if required fields missing
	MinCompletenessScore float64 // Minimum acceptable score (0.0-1.0)
	CheckModelCard       bool    // Validate model card fields
}

func Validate(bom *cdx.BOM, opts ValidationOptions) ValidationResult {

	result := ValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// 1. Basic structural validation
	if bom == nil {
		result.Valid = false
		result.Errors = append(result.Errors, "BOM is nil")
		return result
	}

	// 2. Check metadata component exists
	if bom.Metadata == nil || bom.Metadata.Component == nil {
		result.Valid = false
		result.Errors = append(result.Errors, "BOM missing metadata.component")
	}

	// 3. Run completeness check (leverages existing package)
	report := completeness.Check(bom)
	result.CompletenessScore = report.Score
	result.MissingRequired = report.MissingRequired
	result.MissingOptional = report.MissingOptional

	// 4. Strict mode enforcement
	if opts.StrictMode {
		if len(report.MissingRequired) > 0 {
			result.Valid = false
			for _, key := range report.MissingRequired {
				msg := fmt.Sprintf("required field missing: %s", key)
				result.Errors = append(result.Errors, msg)
			}
		}

		if report.Score < opts.MinCompletenessScore {
			result.Valid = false
			msg := fmt.Sprintf("completeness score %.2f below minimum %.2f", report.Score, opts.MinCompletenessScore)
			result.Errors = append(result.Errors, msg)
		}
	}

	// 5. Add warnings for optional fields
	for _, key := range report.MissingOptional {
		msg := fmt.Sprintf("optional field missing: %s", key)
		result.Warnings = append(result.Warnings, msg)
	}

	// 6. Model card validation
	if opts.CheckModelCard {
		validateModelCard(bom, &result)
	}

	return result
}

func validateModelCard(bom *cdx.BOM, result *ValidationResult) {
	comp := bom.Metadata.Component
	if comp == nil {
		return
	}

	if comp.ModelCard == nil {
		result.Warnings = append(result.Warnings, "model card not present")
		return
	}

	if comp.ModelCard.ModelParameters == nil {
		result.Warnings = append(result.Warnings, "model parameters not present")
	}
}
