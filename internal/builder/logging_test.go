package builder

import (
	"bytes"
	"strings"
	"testing"

	"aibomgen-cra/internal/ui"
)

func TestSetLogger_EnablesAndDisablesLogging(t *testing.T) {
	var buf bytes.Buffer
	SetLogger(&buf)
	t.Cleanup(func() { SetLogger(nil) })

	logf("org/model", "hello")
	if buf.Len() == 0 {
		t.Fatalf("expected log output when logger is set")
	}

	// disable and verify no more output is produced
	buf.Reset()
	SetLogger(nil)
	logf("org/model", "hello")
	if buf.Len() != 0 {
		t.Fatalf("expected no output when logger is nil")
	}
}

func TestLogf_WritesPrefixModelAndMessage(t *testing.T) {
	ui.Init(true) // disable ANSI color for stable assertions

	var buf bytes.Buffer
	SetLogger(&buf)
	t.Cleanup(func() { SetLogger(nil) })

	logf("  org/model  ", "msg %d", 1)

	out := buf.String()
	if !strings.Contains(out, "Build:") {
		t.Fatalf("expected Build: prefix, got %q", out)
	}
	if !strings.Contains(out, "model=org/model") {
		t.Fatalf("expected trimmed model id, got %q", out)
	}
	if !strings.Contains(out, "msg 1") {
		t.Fatalf("expected formatted message, got %q", out)
	}
}

func TestLogf_EmptyModelID_UsesUnknown(t *testing.T) {
	ui.Init(true)

	var buf bytes.Buffer
	SetLogger(&buf)
	t.Cleanup(func() { SetLogger(nil) })

	logf("   ", "x")

	out := buf.String()
	if !strings.Contains(out, "model=(unknown)") {
		t.Fatalf("expected unknown model id, got %q", out)
	}
}
