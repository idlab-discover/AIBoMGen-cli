package completeness

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/idlab-discover/AIBoMGen-cli/internal/fetcher"
	"github.com/idlab-discover/AIBoMGen-cli/internal/metadata"
	"github.com/idlab-discover/AIBoMGen-cli/internal/scanner"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

// expectedTotalFromRegistry returns the number of metadata specifications
func expectedTotalFromRegistry() int {
	total := 0
	for _, spec := range metadata.Registry() {
		if spec.Weight > 0 {
			total++
		}
	}
	return total
}

func buildFullyPopulatedBOMForRegistry(t *testing.T) *cdx.BOM {
	t.Helper()

	modelID := "org/model"

	// HF response with enough fields to satisfy all Present() checks for weighted specs.
	hf := &fetcher.ModelAPIResponse{}
	hf.ID = modelID
	hf.ModelID = modelID
	hf.Author = "some-author"
	hf.PipelineTag = "text-classification"
	hf.LibraryName = "transformers"
	hf.Tags = []string{"tag1", "license:apache-2.0", "dataset:ds1", "dataset:ds2"}
	hf.License = "mit"
	hf.SHA = "deadbeef"
	hf.Downloads = 123
	hf.Likes = 7
	hf.LastMod = "2020-01-01"
	hf.CreatedAt = "2019-01-01"
	hf.Private = false
	hf.UsedStorage = 42
	hf.CardData = map[string]any{
		"license":  "mit",
		"language": []any{"en", "fr"},
		"datasets": []any{"ds1", "ds3"},
	}
	hf.Config.ModelType = "bert"
	hf.Config.Architectures = []string{"BertForSequenceClassification"}

	bom := cdx.NewBOM()
	comp := &cdx.Component{
		Type:      cdx.ComponentTypeMachineLearningModel,
		Name:      modelID,
		ModelCard: &cdx.MLModelCard{},
	}
	bom.Metadata = &cdx.Metadata{Component: comp}

	src := metadata.Source{
		ModelID: modelID,
		Scan: scanner.Discovery{
			// Keep Name empty so it doesn't override HF/ModelID name logic.
			Type:     "model",
			Path:     "/tmp/x.py",
			Evidence: "from_pretrained() pattern at line 1",
		},
		HF: hf,
	}
	tgt := metadata.Target{
		BOM:                       bom,
		Component:                 comp,
		ModelCard:                 comp.ModelCard,
		IncludeEvidenceProperties: true,
		HuggingFaceBaseURL:        "https://huggingface.co/",
	}

	for _, spec := range metadata.Registry() {
		if spec.Apply != nil {
			spec.Apply(src, tgt)
		}
	}

	return bom
}

func writeBOMFile(t *testing.T, path string, format cdx.BOMFileFormat) {
	t.Helper()

	bom := cdx.NewBOM()
	bom.Metadata = &cdx.Metadata{
		Component: &cdx.Component{
			Type: cdx.ComponentTypeMachineLearningModel,
			Name: "org/model",
		},
	}

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create file: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })

	enc := cdx.NewBOMEncoder(f, format)
	enc.SetPretty(true)

	if err := enc.Encode(bom); err != nil {
		t.Fatalf("encode bom: %v", err)
	}
}

// Test Check function on empty BOM
func TestCheck_EmptyBOM_MissingRequiredMatchesRegistry(t *testing.T) {
	r := Check(&cdx.BOM{})

	if r.Total != expectedTotalFromRegistry() {
		t.Fatalf("Total = %d, want %d", r.Total, expectedTotalFromRegistry())
	}

	// Current registry has ComponentName as Required+Weight>0; assert it's missing.
	found := false
	for _, k := range r.MissingRequired {
		if k == metadata.ComponentName {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected MissingRequired to include %q; got %#v", metadata.ComponentName, r.MissingRequired)
	}

	if r.Passed != 0 {
		t.Fatalf("Passed = %d, want 0", r.Passed)
	}
	if r.Score != 0 {
		t.Fatalf("Score = %v, want 0", r.Score)
	}
}

// Test Check function on fully populated BOM
func TestCheck_FullyPopulatedBOM_ScoreIsOne(t *testing.T) {
	bom := buildFullyPopulatedBOMForRegistry(t)
	r := Check(bom)

	if r.Total != expectedTotalFromRegistry() {
		t.Fatalf("Total = %d, want %d", r.Total, expectedTotalFromRegistry())
	}
	if r.Passed != r.Total {
		t.Fatalf("Passed = %d, want %d", r.Passed, r.Total)
	}
	if len(r.MissingRequired) != 0 {
		t.Fatalf("MissingRequired = %#v, want empty", r.MissingRequired)
	}
	if len(r.MissingOptional) != 0 {
		t.Fatalf("MissingOptional = %#v, want empty", r.MissingOptional)
	}

	// Float-safe compare
	if r.Score < 0.999999 {
		t.Fatalf("Score = %v, want ~1.0", r.Score)
	}
}

// Test ReadBOM function with various formats and error cases
func TestReadBOM_JSON_ExplicitFormat(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bom.json")
	writeBOMFile(t, p, cdx.BOMFileFormatJSON)

	b, err := ReadBOM(p, "json")
	if err != nil {
		t.Fatalf("ReadBOM err = %v", err)
	}
	if b == nil || b.Metadata == nil || b.Metadata.Component == nil {
		t.Fatalf("decoded bom missing metadata/component")
	}
}
func TestReadBOM_XML_ExplicitFormat(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bom.xml")
	writeBOMFile(t, p, cdx.BOMFileFormatXML)

	b, err := ReadBOM(p, "xml")
	if err != nil {
		t.Fatalf("ReadBOM err = %v", err)
	}
	if b == nil || b.Metadata == nil || b.Metadata.Component == nil {
		t.Fatalf("decoded bom missing metadata/component")
	}
}
func TestReadBOM_AutoDetectsByExtension(t *testing.T) {
	dir := t.TempDir()

	pJSON := filepath.Join(dir, "bom.json")
	writeBOMFile(t, pJSON, cdx.BOMFileFormatJSON)
	if _, err := ReadBOM(pJSON, "auto"); err != nil {
		t.Fatalf("ReadBOM(json, auto) err = %v", err)
	}

	pXML := filepath.Join(dir, "bom.xml")
	writeBOMFile(t, pXML, cdx.BOMFileFormatXML)
	if _, err := ReadBOM(pXML, ""); err != nil { // empty behaves like auto
		t.Fatalf("ReadBOM(xml, empty) err = %v", err)
	}
}

// Test ReadBOM function error cases
func TestReadBOM_InvalidPath_ReturnsError(t *testing.T) {
	if _, err := ReadBOM("/definitely/does/not/exist/lol.json", "json"); err == nil {
		t.Fatalf("expected error for missing file")
	}
}

// Test ReadBOM function error cases
func TestReadBOM_InvalidContent_ReturnsDecodeError(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(p, []byte("{not valid json"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if _, err := ReadBOM(p, "json"); err == nil {
		t.Fatalf("expected decode error")
	}
}

func TestReadBOM_FormatMismatch_ReturnsDecodeError(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bom.json")
	writeBOMFile(t, p, cdx.BOMFileFormatJSON)

	if _, err := ReadBOM(p, "xml"); err == nil {
		t.Fatalf("expected decode error when decoding json as xml")
	}
}
