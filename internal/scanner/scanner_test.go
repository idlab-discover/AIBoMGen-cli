package scanner

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/idlab-discover/AIBoMGen-cli/internal/ui"
)

func TestScanDetectsModelsDedupesAndLogs(t *testing.T) {
	ui.Init(true)
	dir := t.TempDir()
	pyPath := filepath.Join(dir, "use_model.py")
	pyContent := "from transformers import AutoModel\n" +
		"AutoModel.from_pretrained(\"bert-base-uncased\")\n" +
		"AutoModel.from_pretrained(\"bert-base-uncased\")\n"
	if err := os.WriteFile(pyPath, []byte(pyContent), 0o644); err != nil {
		t.Fatalf("write python file: %v", err)
	}

	var buf bytes.Buffer
	SetLogger(&buf)
	t.Cleanup(func() { SetLogger(nil) })

	comps, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(comps) != 1 {
		t.Fatalf("expected 1 component, got %d", len(comps))
	}
	if !strings.Contains(comps[0].Evidence, "line 2") || !strings.Contains(comps[0].Evidence, "line 3") {
		t.Fatalf("expected evidence to include both occurrences, got %q", comps[0].Evidence)
	}

	logs := buf.String()
	if !strings.Contains(logs, "found model 'bert-base-uncased'") {
		t.Fatalf("expected detection log, got %q", logs)
	}
	if !strings.Contains(logs, "detected 1 components (models: 1)") {
		t.Fatalf("expected summary log, got %q", logs)
	}
}

func TestScanSkipsUnreadableFiles(t *testing.T) {
	dir := t.TempDir()
	pyPath := filepath.Join(dir, "blocked.py")
	if err := os.WriteFile(pyPath, []byte("AutoModel.from_pretrained(\"bert\")"), 0o644); err != nil {
		t.Fatalf("write python file: %v", err)
	}
	if err := os.Chmod(pyPath, 0o000); err != nil {
		t.Fatalf("chmod file: %v", err)
	}
	defer func() { _ = os.Chmod(pyPath, 0o644) }()

	comps, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(comps) != 0 {
		t.Fatalf("expected no components for unreadable files, got %d", len(comps))
	}
}

func TestScanInvalidRootReturnsError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	if _, err := Scan(missing); err == nil {
		t.Fatalf("expected error for missing root")
	}
}

func TestShouldScanForModelID(t *testing.T) {
	tests := []struct {
		ext  string
		want bool
	}{
		{ext: ".py", want: true},
		{ext: ".ipynb", want: true},
		{ext: ".txt", want: true},
		{ext: ".md", want: false},
	}
	for _, tt := range tests {
		if got := shouldScanForModelID(tt.ext); got != tt.want {
			t.Fatalf("shouldScanForModelID(%q) = %t, want %t", tt.ext, got, tt.want)
		}
	}
}

func TestDedupeMergesEvidence(t *testing.T) {
	components := []Discovery{
		{ID: "bert", Type: "model", Evidence: "line 2"},
		{ID: "bert", Type: "model", Evidence: "line 3"},
		{ID: "bert", Type: "model", Evidence: "line 3"},
		{ID: "other", Type: "model", Evidence: "line 5"},
	}

	deduped := dedupe(components)
	if len(deduped) != 2 {
		t.Fatalf("expected 2 unique components, got %d", len(deduped))
	}

	var merged Discovery
	for _, c := range deduped {
		if c.ID == "bert" {
			merged = c
		}
	}
	if merged.ID == "" {
		t.Fatalf("expected bert component after dedupe")
	}
	if !strings.Contains(merged.Evidence, "line 2") {
		t.Fatalf("expected merged evidence to include line 2, got %q", merged.Evidence)
	}
	if strings.Count(merged.Evidence, "line 3") != 1 {
		t.Fatalf("expected line 3 evidence once, got %q", merged.Evidence)
	}
}
