package validator

import (
	"fmt"
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/idlab-discover/AIBoMGen-cli/internal/metadata"
)

func TestValidate_NilBOM(t *testing.T) {
	opts := ValidationOptions{}
	result := Validate(nil, opts)

	if result.Valid {
		t.Error("expected validation to fail for nil BOM")
	}

	if len(result.Errors) == 0 {
		t.Error("expected at least one error for nil BOM")
	}
}

func TestValidate_MissingMetadata(t *testing.T) {
	bom := &cdx.BOM{}
	opts := ValidationOptions{}
	result := Validate(bom, opts)

	if result.Valid {
		t.Error("expected validation to fail for BOM without metadata")
	}

	if len(result.Errors) == 0 {
		t.Error("expected at least one error for missing metadata")
	}
}

func TestValidate_ValidBOM(t *testing.T) {
	bom := &cdx.BOM{
		SpecVersion: cdx.SpecVersion1_6,
		Metadata: &cdx.Metadata{
			Component: &cdx.Component{
				Name: "test-model",
			},
		},
	}
	opts := ValidationOptions{
		StrictMode:     false,
		CheckModelCard: false,
	}
	result := Validate(bom, opts)

	// Without strict mode, a basic BOM should pass
	if !result.Valid {
		t.Errorf("expected validation to pass, got errors: %v", result.Errors)
	}
}

func TestValidate_StrictMode(t *testing.T) {
	bom := &cdx.BOM{
		SpecVersion: cdx.SpecVersion1_6,
		Metadata: &cdx.Metadata{
			Component: &cdx.Component{
				Name: "test-model",
			},
		},
	}
	opts := ValidationOptions{
		StrictMode:           true,
		MinCompletenessScore: 0.8,
		CheckModelCard:       false,
	}
	result := Validate(bom, opts)

	// In strict mode with high min score, this simple BOM should fail
	if result.Valid {
		t.Error("expected validation to fail in strict mode with incomplete BOM")
	}
}

func TestValidate_StrictModeReportsMissingRequiredAndScore(t *testing.T) {
	bom := &cdx.BOM{
		SpecVersion: cdx.SpecVersion1_6,
		Metadata: &cdx.Metadata{
			Component: &cdx.Component{},
		},
	}
	opts := ValidationOptions{
		StrictMode:           true,
		MinCompletenessScore: 0.9,
	}
	result := Validate(bom, opts)

	if result.Valid {
		t.Fatal("expected validation to fail")
	}

	missingRequiredMsg := "required field missing: " + metadata.ComponentName.String()
	if !containsString(result.Errors, missingRequiredMsg) {
		t.Fatalf("expected missing required error, got %v", result.Errors)
	}

	scoreMsg := fmt.Sprintf("completeness score %.2f below minimum %.2f", result.CompletenessScore, opts.MinCompletenessScore)
	if !containsString(result.Errors, scoreMsg) {
		t.Fatalf("expected score threshold error, got %v", result.Errors)
	}
}

func TestValidateModelCard_Missing(t *testing.T) {
	bom := &cdx.BOM{
		SpecVersion: cdx.SpecVersion1_6,
		Metadata: &cdx.Metadata{
			Component: &cdx.Component{
				Name: "test-model",
			},
		},
	}
	opts := ValidationOptions{
		StrictMode:     false,
		CheckModelCard: true,
	}
	result := Validate(bom, opts)

	// Should have warnings about missing model card
	if len(result.Warnings) == 0 {
		t.Error("expected warnings about missing model card")
	}
}

func TestValidateModelCard_ParametersMissing(t *testing.T) {
	bom := &cdx.BOM{
		SpecVersion: cdx.SpecVersion1_6,
		Metadata: &cdx.Metadata{
			Component: &cdx.Component{
				Name:      "test-model",
				ModelCard: &cdx.MLModelCard{},
			},
		},
	}
	result := Validate(bom, ValidationOptions{CheckModelCard: true})

	if !containsString(result.Warnings, "model parameters not present") {
		t.Fatalf("expected warning about missing model parameters, got %v", result.Warnings)
	}
}

func TestValidateModelCard_ComponentMissing(t *testing.T) {
	bom := &cdx.BOM{
		SpecVersion: cdx.SpecVersion1_6,
		Metadata:    &cdx.Metadata{},
	}
	result := Validate(bom, ValidationOptions{CheckModelCard: true})

	if !containsString(result.Errors, "BOM missing metadata.component") {
		t.Fatalf("expected structural error, got %v", result.Errors)
	}
	if containsString(result.Warnings, "model card not present") {
		t.Fatalf("expected no model card warning when component is nil, got %v", result.Warnings)
	}
}

func TestValidateSpecVersion_Missing(t *testing.T) {
	bom := &cdx.BOM{
		Metadata: &cdx.Metadata{
			Component: &cdx.Component{
				Name: "test-model",
			},
		},
	}
	opts := ValidationOptions{}
	result := Validate(bom, opts)

	if result.Valid {
		t.Error("expected validation to fail for BOM without spec version")
	}

	foundError := false
	for _, err := range result.Errors {
		if err == "BOM missing spec version" {
			foundError = true
			break
		}
	}
	if !foundError {
		t.Error("expected error about missing spec version")
	}
}

func TestValidateSpecVersion_Invalid(t *testing.T) {
	bom := &cdx.BOM{
		SpecVersion: 999,
		Metadata: &cdx.Metadata{
			Component: &cdx.Component{
				Name: "test-model",
			},
		},
	}
	opts := ValidationOptions{}
	result := Validate(bom, opts)

	if result.Valid {
		t.Error("expected validation to fail for invalid spec version")
	}
}

func TestValidateSpecVersion_Valid(t *testing.T) {
	bom := &cdx.BOM{
		SpecVersion: cdx.SpecVersion1_6,
		Metadata: &cdx.Metadata{
			Component: &cdx.Component{
				Name: "test-model",
			},
		},
	}
	opts := ValidationOptions{
		CheckModelCard: false,
	}
	result := Validate(bom, opts)

	if !result.Valid {
		t.Errorf("expected validation to pass for valid spec version, got errors: %v", result.Errors)
	}
}

func TestValidateSpecVersion_OldVersionWarning(t *testing.T) {
	bom := &cdx.BOM{
		SpecVersion: cdx.SpecVersion1_3,
		Metadata: &cdx.Metadata{
			Component: &cdx.Component{
				Name: "test-model",
			},
		},
	}
	opts := ValidationOptions{
		CheckModelCard: false,
	}
	result := Validate(bom, opts)

	// Should have warning about old spec version
	foundWarning := false
	for _, warn := range result.Warnings {
		if len(warn) > 0 && (warn[0:4] == "spec" || warn[0:4] == "Spec") {
			foundWarning = true
			break
		}
	}
	if !foundWarning {
		t.Error("expected warning about old spec version")
	}
}

func containsString(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}
