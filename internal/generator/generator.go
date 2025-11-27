package generator

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"aibomgen-cra/internal/fetcher"
	"aibomgen-cra/internal/scanner"
	"aibomgen-cra/internal/ui"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

// Build constructs a minimal BOM from scanner components.
func Build(components []scanner.Component) *cdx.BOM {
	var cdxComponents []cdx.Component
	for i, comp := range components {
		if logOut != nil {
			if i > 0 {
				fmt.Fprintln(logOut)
			}
			prefix := ui.Color("Model context:", ui.FgCyan)
			fmt.Fprintf(logOut, "%s %s (path: %s)\n", prefix, comp.Name, comp.Path)
		}
		// Build model card directly via fetcher using CycloneDX types
		modelCard := BuildMLModelCard(comp.Name)
		// Base properties
		props := []cdx.Property{
			{
				Name:  "aibomgen.type",
				Value: comp.Type,
			},
			{
				Name:  "aibomgen.evidence",
				Value: comp.Evidence,
			},
			{
				Name:  "aibomgen.path",
				Value: comp.Path,
			},
		}
		// Enrich with HF-derived hints for convenience
		if modelCard != nil && modelCard.ModelParameters != nil {
			if modelCard.ModelParameters.Task != "" {
				props = append(props, cdx.Property{Name: "huggingface.pipeline_task", Value: modelCard.ModelParameters.Task})
			}
			if modelCard.ModelParameters.Datasets != nil {
				for _, ds := range *modelCard.ModelParameters.Datasets {
					if ds.Ref != "" {
						props = append(props, cdx.Property{Name: "huggingface.dataset", Value: ds.Ref})
					}
				}
			}
		}
		cdxComp := cdx.Component{
			Type:       cdx.ComponentTypeMachineLearningModel,
			Name:       comp.Name,
			Version:    "", // No version info from scanner
			ModelCard:  modelCard,
			Properties: &props,
		}
		// Add external reference to HF model card page
		cdxComp.ExternalReferences = &[]cdx.ExternalReference{{
			Type: cdx.ExternalReferenceType("website"),
			URL:  "https://huggingface.co/" + comp.Name,
		}}
		cdxComponents = append(cdxComponents, cdxComp)
	}
	bom := cdx.NewBOM()
	bom.Components = &cdxComponents
	return bom
}

// BuildMLModelCard fetches and returns a CycloneDX MLModelCard directly
// using the fetcher without intermediate custom structs.
func BuildMLModelCard(modelID string) *cdx.MLModelCard {
	return fetcher.FetchModelCard(modelID)
}

// Write writes the BOM to the given output path, creating directories as needed.
func Write(outputPath string, bom *cdx.BOM) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	encoder := cdx.NewBOMEncoder(f, cdx.BOMFileFormatJSON)
	encoder.SetPretty(true)
	return encoder.Encode(bom)
}

var logOut io.Writer

// SetLogger sets an optional logger to provide context around fetcher logs.
func SetLogger(w io.Writer) { logOut = w }
