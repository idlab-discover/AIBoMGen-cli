package builder

import (
	"aibomgen-cra/internal/scanner"
	"strings"
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

// Helper to quickly find properties by name
func findProperty(props *[]cdx.Property, name string) (string, bool) {
	if props == nil {
		return "", false
	}
	for _, p := range *props {
		if strings.TrimSpace(p.Name) == name && strings.TrimSpace(p.Value) != "" {
			return p.Value, true
		}
	}
	return "", false
}

// Test DefaultOptions function
func TestDefaultOptions_ReturnsExpectedDefaults(t *testing.T) {
	opts := DefaultOptions()
	if opts.IncludeEvidenceProperties != true {
		t.Errorf("Expected IncludeEvidenceProperties to be true by default, got %v", opts.IncludeEvidenceProperties)
	}
	// if url is not one of the two accepted forms
	if opts.HuggingFaceBaseURL != "https://huggingface.co" && opts.HuggingFaceBaseURL != "https://huggingface.co/" {
		t.Errorf("Expected HuggingFaceBaseURL to be 'https://huggingface.co' by default, got %s", opts.HuggingFaceBaseURL)
	}
}

// Test BuildMetadataComponent function name fallbacks
func TestBuildMetadataComponent_NamePrefersModelID(t *testing.T) {
	ctx := BuildContext{
		ModelID: "test-model-id",
		Scan: scanner.Discovery{
			Name: "scan-name",
		},
	}
	comp := buildMetadataComponent(ctx)

	if comp.Name != "test-model-id" {
		t.Fatalf("component name = %q, want %q", comp.Name, "org/model")
	}
}
func TestBuildMetadataComponent_NameFallsBackToScanName(t *testing.T) {
	ctx := BuildContext{
		ModelID: "",
		Scan: scanner.Discovery{
			Name: "scan-name",
		},
	}
	comp := buildMetadataComponent(ctx)

	if comp.Name != "scan-name" {
		t.Fatalf("component name = %q, want %q", comp.Name, "scan-name")
	}
}
func TestBuildMetadataComponent_NameFallsBackToLiteralModel(t *testing.T) {
	ctx := BuildContext{
		ModelID: "",
		Scan:    scanner.Discovery{},
	}
	comp := buildMetadataComponent(ctx)

	if comp.Name != "model" {
		t.Fatalf("component name = %q, want %q", comp.Name, "model")
	}
}

// Test BuildMetadataComponent sets type and non-nil ModelCard
func TestBuildMetadataComponent_SetsTypeAndNonNilModelCard(t *testing.T) {
	ctx := BuildContext{
		ModelID: "org/model",
		Scan:    scanner.Discovery{},
	}

	comp := buildMetadataComponent(ctx)
	if comp == nil {
		t.Fatalf("buildMetadataComponent() returned nil")
	}
	if comp.Type != cdx.ComponentTypeMachineLearningModel {
		t.Fatalf("Type = %q, want %q", comp.Type, cdx.ComponentTypeMachineLearningModel)
	}
	if comp.ModelCard == nil {
		t.Fatalf("ModelCard is nil, want non-nil")
	}
}

// Test Build function creates BOM with metadata component
func TestBOMBuilder_Build_CreatesBOMWithMetadataComponent(t *testing.T) {
	b := NewBOMBuilder(DefaultOptions())

	// Keep Scan.Name empty so registry won't override the name.
	ctx := BuildContext{
		ModelID: "org/model",
		Scan:    scanner.Discovery{},
		HF:      nil,
	}

	bom, err := b.Build(ctx)
	if err != nil {
		t.Fatalf("Build() err = %v", err)
	}
	if bom == nil {
		t.Fatalf("Build() bom is nil")
	}
	if bom.Metadata == nil {
		t.Fatalf("bom.Metadata is nil")
	}
	if bom.Metadata.Component == nil {
		t.Fatalf("bom.Metadata.Component is nil")
	}

	comp := bom.Metadata.Component
	if comp.Type != cdx.ComponentTypeMachineLearningModel {
		t.Fatalf("component.Type = %q, want %q", comp.Type, cdx.ComponentTypeMachineLearningModel)
	}
	if strings.TrimSpace(comp.Name) == "" {
		t.Fatalf("component.Name is empty; want non-empty")
	}
	if comp.ModelCard == nil {
		t.Fatalf("component.ModelCard is nil; want non-nil")
	}
}

// Test Build function passes HuggingFaceBaseURL into registry
func TestBOMBuilder_Build_PassesHuggingFaceBaseURLIntoRegistry(t *testing.T) {
	opts := DefaultOptions()
	opts.HuggingFaceBaseURL = "https://example.com" // no trailing slash on purpose

	b := NewBOMBuilder(opts)
	bom, err := b.Build(BuildContext{
		ModelID: "org/model",
		Scan:    scanner.Discovery{Name: "ignored"},
	})
	if err != nil {
		t.Fatalf("Build() err = %v", err)
	}

	comp := bom.Metadata.Component
	if comp.ExternalReferences == nil || len(*comp.ExternalReferences) == 0 {
		t.Fatalf("expected external references to be set")
	}

	got := (*comp.ExternalReferences)[0].URL
	want := "https://example.com/org/model"
	if got != want {
		t.Fatalf("external reference URL = %q, want %q", got, want)
	}
}

// Test Build function includes or excludes evidence properties when enabled
func TestBOMBuilder_Build_IncludesEvidenceProperties_WhenEnabled(t *testing.T) {
	opts := DefaultOptions()
	opts.IncludeEvidenceProperties = true

	b := NewBOMBuilder(opts)
	bom, err := b.Build(BuildContext{
		ModelID: "org/model",
		Scan: scanner.Discovery{
			Type:     "model",
			Path:     "/tmp/x.py",
			Evidence: "found from_pretrained(...)",
		},
	})
	if err != nil {
		t.Fatalf("Build() err = %v", err)
	}

	props := bom.Metadata.Component.Properties

	// Verify properties exist
	if v, ok := findProperty(props, "aibomgen.type"); !ok {
		t.Fatalf("missing/incorrect aibomgen.type: ok=%v value=%q", ok, v)
	}
	if v, ok := findProperty(props, "aibomgen.path"); !ok {
		t.Fatalf("missing/incorrect aibomgen.path: ok=%v value=%q", ok, v)
	}
	if v, ok := findProperty(props, "aibomgen.evidence"); !ok {
		t.Fatalf("missing/incorrect aibomgen.evidence: ok=%v value=%q", ok, v)
	}
}

// Test Build function excludes evidence properties when disabled
func TestBOMBuilder_Build_DoesNotIncludeEvidenceProperties_WhenDisabled(t *testing.T) {
	opts := DefaultOptions()
	opts.IncludeEvidenceProperties = false

	b := NewBOMBuilder(opts)
	bom, err := b.Build(BuildContext{
		ModelID: "org/model",
		Scan: scanner.Discovery{
			Type:     "model",
			Path:     "/tmp/x.py",
			Evidence: "found from_pretrained(...)",
		},
	})
	if err != nil {
		t.Fatalf("Build() err = %v", err)
	}

	props := bom.Metadata.Component.Properties

	if v, ok := findProperty(props, "aibomgen.type"); ok {
		t.Fatalf("missing/incorrect aibomgen.type: ok=%v value=%q", ok, v)
	}
	if v, ok := findProperty(props, "aibomgen.path"); ok {
		t.Fatalf("missing/incorrect aibomgen.path: ok=%v value=%q", ok, v)
	}
	if v, ok := findProperty(props, "aibomgen.evidence"); ok {
		t.Fatalf("missing/incorrect aibomgen.evidence: ok=%v value=%q", ok, v)
	}
}

// Test pruneEmptyModelParameters function
func TestPruneEmptyModelParameters_RemovesWhenAllFieldsEmpty(t *testing.T) {
	comp := &cdx.Component{
		ModelCard: &cdx.MLModelCard{
			ModelParameters: &cdx.MLModelParameters{
				// all empty
				Task:               "",
				ArchitectureFamily: "",
				ModelArchitecture:  "",
				Datasets:           nil, // also try empty slice if you want
			},
		},
	}

	pruneEmptyModelParameters(comp)

	if comp.ModelCard == nil {
		t.Fatalf("ModelCard unexpectedly nil")
	}
	if comp.ModelCard.ModelParameters != nil {
		t.Fatalf("ModelParameters = %#v, want nil (pruned)", comp.ModelCard.ModelParameters)
	}
}

// Test pruneEmptyModelParameters keeps the parameters when any field is present
func TestPruneEmptyModelParameters_KeepsWhenAnyFieldPresent(t *testing.T) {
	comp := &cdx.Component{
		ModelCard: &cdx.MLModelCard{
			ModelParameters: &cdx.MLModelParameters{
				Task: "text-classification", // makes it non-empty
			},
		},
	}

	pruneEmptyModelParameters(comp)

	if comp.ModelCard == nil || comp.ModelCard.ModelParameters == nil {
		t.Fatalf("ModelParameters was pruned, want it kept")
	}
	if comp.ModelCard.ModelParameters.Task != "text-classification" {
		t.Fatalf("Task = %q, want %q", comp.ModelCard.ModelParameters.Task, "text-classification")
	}
}
