package validator

import (
	"testing"
)

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
