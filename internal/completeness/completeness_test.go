package completeness

import (
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
		Readme: &fetcher.ModelReadmeCard{
			BaseModel:                  "bert-base-uncased",
			ModelCardContact:           "contact@example.com",
			DirectUse:                  "Use for classification.",
			OutOfScopeUse:              "Do not use for medical.",
			BiasRisksLimitations:       "May be biased.",
			BiasRecommendations:        "Use with care.",
			EnvironmentalHardwareType:  "NVIDIA A100",
			EnvironmentalHoursUsed:     "10",
			EnvironmentalCloudProvider: "AWS",
			EnvironmentalComputeRegion: "us-east-1",
			EnvironmentalCarbonEmitted: "123g",
			ModelIndexMetrics:          []fetcher.ModelIndexMetric{{Type: "accuracy", Value: "0.91"}},
		},
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
