package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanDetectsModelAndWeight(t *testing.T) {
	dir := t.TempDir()
	// Create a pseudo weight file
	weightPath := filepath.Join(dir, "model.bin")
	if err := os.WriteFile(weightPath, []byte("binarydata"), 0o644); err != nil {
		t.Fatalf("failed to create weight file: %v", err)
	}
	// Create a python file with a HF model reference
	pyPath := filepath.Join(dir, "use_model.py")
	pyContent := `from transformers import AutoModel\nAutoModel.from_pretrained("bert-base-uncased")`
	if err := os.WriteFile(pyPath, []byte(pyContent), 0o644); err != nil {
		t.Fatalf("failed to create python file: %v", err)
	}

	comps, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	var foundWeight, foundModel bool
	for _, c := range comps {
		if c.Type == "weight-file" && c.Path == weightPath {
			foundWeight = true
		}
		if c.Type == "model" && c.Name == "bert-base-uncased" {
			foundModel = true
		}
	}
	if !foundWeight {
		t.Errorf("expected weight file component for %s", weightPath)
	}
	if !foundModel {
		t.Errorf("expected model component for bert-base-uncased")
	}
}
