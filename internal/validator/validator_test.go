package validator

import (
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
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

func TestValidateModelCard_Missing(t *testing.T) {
	bom := &cdx.BOM{
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
