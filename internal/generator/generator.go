package generator

import (
	"os"
	"path/filepath"

	"aibomgen-cra/internal/fetcher"
	"aibomgen-cra/internal/scanner"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

// Build constructs a minimal BOM from scanner components.
func Build(components []scanner.Component) *cdx.BOM {
	var cdxComponents []cdx.Component
	for _, comp := range components {
		// Build model card directly via fetcher using CycloneDX types
		modelCard := BuildMLModelCard(comp.Name)
		cdxComp := cdx.Component{
			Type:      cdx.ComponentTypeMachineLearningModel,
			Name:      comp.Name,
			Version:   "", // No version info from scanner
			ModelCard: modelCard,
			Properties: &[]cdx.Property{
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
			},
		}
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
