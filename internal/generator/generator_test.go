package generator

import (
	"os"
	"path/filepath"
	"testing"

	"aibomgen-cra/internal/scanner"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

func TestBuildProducesComponentsWithModelCard(t *testing.T) {
	comps := []scanner.Component{{
		ID:       "bert-base-uncased",
		Name:     "bert-base-uncased",
		Type:     "model",
		Path:     "testdata/repo-basic/src/use_model.py",
		Evidence: "from_pretrained()",
	}}

	bom := Build(comps)
	if bom == nil {
		t.Fatalf("expected BOM not nil")
	}
	if bom.Components == nil {
		t.Fatalf("expected components present in BOM")
	}
	if len(*bom.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(*bom.Components))
	}
	comp := (*bom.Components)[0]
	if comp.ModelCard == nil {
		t.Fatalf("expected component to have ModelCard")
	}
	if comp.Type != cdx.ComponentTypeMachineLearningModel {
		t.Errorf("expected component type MachineLearningModel")
	}
	if comp.Properties == nil || len(*comp.Properties) == 0 {
		t.Errorf("expected properties to be set")
	}
}

func TestBuildMLModelCardReturnsCard(t *testing.T) {
	card := BuildMLModelCard("bert-base-uncased")
	if card == nil {
		t.Fatalf("expected MLModelCard not nil")
	}
	// No static defaults: ModelParameters may be empty when offline/non-configured
}

func TestWriteCreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "bom.json")

	comps := []scanner.Component{{
		ID:       "bert-base-uncased",
		Name:     "bert-base-uncased",
		Type:     "model",
		Path:     "testdata/repo-basic/src/use_model.py",
		Evidence: "from_pretrained()",
	}}
	bom := Build(comps)

	if err := Write(outPath, bom); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected file to exist at %s: %v", outPath, err)
	}
}
