package completeness

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

	PrintReport(Report{Score: 0.5, Passed: 1, Total: 2, DatasetReports: make(map[string]DatasetReport)})
	got := buf.String()
	want := "Completeness Report: Model score=50.0% (1/2)\n"
	if got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestPrintReport_NoLoggerWriter_DoesNothing(t *testing.T) {
	ui.Init(true)

	SetLogger(nil)
	// Should not panic; should produce no output.
	PrintReport(Report{Score: 1, Passed: 1, Total: 1})
}

func TestPrintReport_WithMissingKeys(t *testing.T) {
	ui.Init(true)

	var buf bytes.Buffer
	SetLogger(&buf)
	t.Cleanup(func() { SetLogger(nil) })

	PrintReport(Report{
		Score:           0,
		Passed:          0,
		Total:           1,
		MissingRequired: []metadata.Key{metadata.ComponentName},
		MissingOptional: []metadata.Key{metadata.ComponentTags, metadata.ComponentLicenses},
		DatasetReports:  make(map[string]DatasetReport),
	})

	got := buf.String()
	want := "Completeness Report: Model score=0.0% (0/1)\n" +
		"Completeness Report: missing required: BOM.metadata.component.name\n" +
		"Completeness Report: missing optional: BOM.metadata.component.tags, BOM.metadata.component.licenses\n"
	if got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestJoinKeys_Empty(t *testing.T) {
	if got := joinKeys(nil); got != "" {
		t.Fatalf("joinKeys(nil) = %q, want empty", got)
	}
	if got := joinKeys([]metadata.Key{}); got != "" {
		t.Fatalf("joinKeys(empty) = %q, want empty", got)
	}
}
