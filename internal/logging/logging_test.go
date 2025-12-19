package logging

import (
	"bytes"
	"strings"
	"testing"

	"aibomgen-cra/internal/ui"
)

func TestLogger_EnabledAndSetWriter(t *testing.T) {
	var l Logger
	if l.Enabled() {
		t.Fatalf("expected disabled when Writer is nil")
	}

	var buf bytes.Buffer
	l.SetWriter(&buf)
	if !l.Enabled() {
		t.Fatalf("expected enabled after setting Writer")
	}
}

func TestLogger_Logf_WritesPrefixModelAndMessage(t *testing.T) {
	ui.Init(true) // disable ANSI color for stable assertions

	var buf bytes.Buffer
	l := Logger{Writer: &buf, PrefixText: "X:", PrefixColor: ui.FgGreen}
	l.Logf("  org/model  ", "msg %d", 1)

	out := buf.String()
	if !strings.Contains(out, "X:") {
		t.Fatalf("expected prefix, got %q", out)
	}
	if !strings.Contains(out, "model=org/model") {
		t.Fatalf("expected trimmed model id, got %q", out)
	}
	if !strings.Contains(out, "msg 1") {
		t.Fatalf("expected formatted message, got %q", out)
	}
}

func TestLogger_Logf_EmptyModelID_UsesUnknown(t *testing.T) {
	ui.Init(true)

	var buf bytes.Buffer
	l := Logger{Writer: &buf, PrefixText: "X:"}
	l.Logf("   ", "x")

	out := buf.String()
	if !strings.Contains(out, "model=(unknown)") {
		t.Fatalf("expected unknown model id, got %q", out)
	}
}

func TestLogger_Logf_DefaultPrefix(t *testing.T) {
	ui.Init(true)

	var buf bytes.Buffer
	l := Logger{Writer: &buf}
	l.Logf("org/model", "x")

	out := buf.String()
	if !strings.Contains(out, "Log:") {
		t.Fatalf("expected default prefix, got %q", out)
	}
}

func TestLogger_Logf_OmitField(t *testing.T) {
	ui.Init(true)

	var buf bytes.Buffer
	l := Logger{Writer: &buf, PrefixText: "X:", OmitModel: true}
	l.Logf("org/model", "x")

	out := buf.String()
	if out != "X: x\n" {
		t.Fatalf("output = %q, want %q", out, "X: x\\n")
	}
}

func TestLogger_Logf_NilReceiver_NoPanic(t *testing.T) {
	ui.Init(true)

	var l *Logger
	l.Logf("org/model", "x")
}
