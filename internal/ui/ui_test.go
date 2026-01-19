package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestColorAppliesANSICodes(t *testing.T) {
	got := Color("hello", FgGreen)
	want := FgGreen + "hello" + Reset
	if got != want {
		t.Fatalf("Color() = %q, want %q", got, want)
	}
}

func TestColorWithEmptyString(t *testing.T) {
	got := Color("", FgRed)
	want := FgRed + "" + Reset
	if got != want {
		t.Fatalf("Color(\"\") = %q, want %q", got, want)
	}
}

// mockFieldKey implements FieldKey for testing
type mockFieldKey string

func (m mockFieldKey) String() string { return string(m) }

func TestCompletenessUI_PrintReport(t *testing.T) {
	tests := []struct {
		name   string
		report CompletenessReport
		quiet  bool
		want   []string // Strings that should appear in output
	}{
		{
			name: "complete model with no datasets",
			report: CompletenessReport{
				Score:           1.0,
				Passed:          10,
				Total:           10,
				MissingRequired: []FieldKey{},
				MissingOptional: []FieldKey{},
				DatasetReports:  map[string]DatasetReport{},
			},
			quiet: false,
			want:  []string{"AIBOM Completeness Report", "Model Component", "100.0%", "(10/10 fields present)"},
		},
		{
			name: "partial model with missing fields",
			report: CompletenessReport{
				Score:           0.75,
				Passed:          15,
				Total:           20,
				MissingRequired: []FieldKey{mockFieldKey("field1"), mockFieldKey("field2")},
				MissingOptional: []FieldKey{mockFieldKey("field3")},
				DatasetReports:  map[string]DatasetReport{},
			},
			quiet: false,
			want:  []string{"AIBOM Completeness Report", "Model Component", "75.0%", "(15/20 fields present)", "Required Fields (2 missing)", "field1", "field2", "Optional Fields (1 missing)", "field3"},
		},
		{
			name: "model with datasets",
			report: CompletenessReport{
				Score:           0.8,
				Passed:          16,
				Total:           20,
				MissingRequired: []FieldKey{},
				MissingOptional: []FieldKey{mockFieldKey("optField")},
				DatasetReports: map[string]DatasetReport{
					"wikipedia": {
						DatasetRef:      "wikipedia",
						Score:           0.9,
						Passed:          9,
						Total:           10,
						MissingRequired: []FieldKey{},
						MissingOptional: []FieldKey{mockFieldKey("dsField")},
					},
				},
			},
			quiet: false,
			want:  []string{"AIBOM Completeness Report", "Model Component", "80.0%", "Dataset Components", "wikipedia", "90.0%", "(9/10 fields present)"},
		},
		{
			name: "quiet mode produces no output",
			report: CompletenessReport{
				Score:  0.5,
				Passed: 5,
				Total:  10,
			},
			quiet: true,
			want:  []string{}, // No output expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ui := NewCompletenessUI(&buf, tt.quiet)
			ui.PrintReport(tt.report)

			output := buf.String()

			// In quiet mode, expect empty output
			if tt.quiet {
				if output != "" {
					t.Errorf("Expected no output in quiet mode, got: %q", output)
				}
				return
			}

			// Check for expected strings
			for _, want := range tt.want {
				if !strings.Contains(output, want) {
					t.Errorf("Output missing expected string %q.\nGot:\n%s", want, output)
				}
			}
		})
	}
}

func TestCompletenessUI_PrintSimpleReport(t *testing.T) {
	report := CompletenessReport{
		Score:           0.75,
		Passed:          15,
		Total:           20,
		MissingRequired: []FieldKey{mockFieldKey("req1")},
		MissingOptional: []FieldKey{mockFieldKey("opt1")},
		DatasetReports: map[string]DatasetReport{
			"dataset1": {
				Score:  0.8,
				Passed: 8,
				Total:  10,
			},
		},
	}

	var buf bytes.Buffer
	ui := NewCompletenessUI(&buf, false)
	ui.PrintSimpleReport(report)

	output := buf.String()
	want := []string{
		"Model score: 75.0% (15/20)",
		"Missing required: req1",
		"Missing optional: opt1",
		"Datasets:",
		"dataset1: 80.0% (8/10)",
	}

	for _, w := range want {
		if !strings.Contains(output, w) {
			t.Errorf("Output missing expected string %q.\nGot:\n%s", w, output)
		}
	}
}

func TestRenderProgressBar(t *testing.T) {
	ui := NewCompletenessUI(nil, false)

	tests := []struct {
		name  string
		score float64
		width int
	}{
		{"full", 1.0, 10},
		{"half", 0.5, 10},
		{"empty", 0.0, 10},
		{"partial", 0.75, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ui.renderProgressBar(tt.score, tt.width)
			// Just verify it produces output of reasonable length
			// (actual rendering involves ANSI codes)
			if len(result) < tt.width {
				t.Errorf("Progress bar too short: got length %d, expected at least %d", len(result), tt.width)
			}
		})
	}
}
