package validator

import (
	"bytes"
	"testing"

	"github.com/idlab-discover/AIBoMGen-cli/internal/metadata"
	"github.com/idlab-discover/AIBoMGen-cli/internal/ui"
)

func TestPrintReport_UsesConfiguredLoggerWriter(t *testing.T) {
	ui.Init(true)

	var buf bytes.Buffer
	SetLogger(&buf)
	t.Cleanup(func() { SetLogger(nil) })

	PrintReport(ValidationResult{
		Valid:             false,
		Errors:            []string{"required field missing"},
		Warnings:          []string{"optional field missing"},
		CompletenessScore: 0.5,
		MissingRequired:   []metadata.Key{metadata.ComponentName},
		MissingOptional:   []metadata.Key{metadata.ComponentTags, metadata.ComponentLicenses},
	})

	got := buf.String()
	want := "Validation Report: ❌ validation failed\n" +
		"Validation Report: errors (1):\n" +
		"Validation Report:   • required field missing\n" +
		"Validation Report: warnings (1):\n" +
		"Validation Report:   • optional field missing\n" +
		"Validation Report: completeness score: 50.0% (1 required, 2 optional missing)\n"

	if got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestPrintReport_NoLoggerWriter_DoesNothing(t *testing.T) {
	ui.Init(true)

	SetLogger(nil)
	PrintReport(ValidationResult{Valid: true})
}

func TestFormatSummary(t *testing.T) {
	tests := []struct {
		name string
		res  ValidationResult
		want string
	}{
		{
			name: "passed",
			res: ValidationResult{
				Valid:             true,
				CompletenessScore: 0.83,
				Errors:            []string{"ignored"},
				Warnings:          []string{"one", "two"},
			},
			want: "Validation: ✅ PASSED | Score: 83.0% | Errors: 1 | Warnings: 2",
		},
		{
			name: "failed",
			res: ValidationResult{
				Valid:             false,
				CompletenessScore: 0.42,
				Errors:            []string{"a", "b"},
				Warnings:          []string{"c"},
			},
			want: "Validation: ❌ FAILED | Score: 42.0% | Errors: 2 | Warnings: 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatSummary(tt.res); got != tt.want {
				t.Fatalf("FormatSummary() = %q, want %q", got, tt.want)
			}
		})
	}
}
