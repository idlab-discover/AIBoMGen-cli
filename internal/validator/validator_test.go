package validator

import (
	"os"
	"path/filepath"
	"testing"
)

const jsonSample = `{"bomFormat":"CycloneDX","specVersion":"1.5","version":1,"metadata":{"component":{"type":"application","name":"app"}},"components":[{"type":"machine-learning-model","name":"model1"}]}`

const xmlSample = `<?xml version="1.0" encoding="UTF-8"?><bom xmlns="http://cyclonedx.org/schema/bom/1.5" version="1"><metadata><component type="application"><name>app</name></component></metadata><components><component type="machine-learning-model"><name>model1</name></component></components></bom>`

func writeTemp(t *testing.T, ext, content string) string {
	dir := t.TempDir()
	path := filepath.Join(dir, "bom"+ext)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}

func TestValidateFromFile_JSONAuto(t *testing.T) {
	path := writeTemp(t, ".json", jsonSample)
	_, errs, err := ValidateFromFile(path, false, "auto")
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("expected no validation errors, got %v", errs)
	}
}

func TestValidateFromFile_XMLAuto(t *testing.T) {
	path := writeTemp(t, ".xml", xmlSample)
	_, errs, err := ValidateFromFile(path, false, "auto")
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("expected no validation errors, got %v", errs)
	}
}

func TestValidateFromFile_ExplicitFormat(t *testing.T) {
	path := writeTemp(t, ".data", jsonSample)
	_, errs, err := ValidateFromFile(path, false, "json")
	if err != nil {
		t.Fatalf("decode failed explicit json: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("expected no validation errors, got %v", errs)
	}
}
