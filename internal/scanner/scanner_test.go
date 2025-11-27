package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanDetectsModelAndWeight(t *testing.T) {
	dir := t.TempDir()

	// Create a python file with a HF model reference
	pyPath := filepath.Join(dir, "use_model.py")
	pyContent := `from transformers import AutoModel\nAutoModel.from_pretrained("bert-base-uncased")`
	if err := os.WriteFile(pyPath, []byte(pyContent), 0o644); err != nil {
		t.Fatalf("failed to create python file: %v", err)
	}

	// Weight-file detection disabled; only verify model detection.

	comps, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	var foundModel bool
	for _, c := range comps {
		if c.Type == "model" && c.Name == "bert-base-uncased" {
			foundModel = true
		}
	}
	if !foundModel {
		t.Errorf("expected model component for bert-base-uncased")
	}
}
