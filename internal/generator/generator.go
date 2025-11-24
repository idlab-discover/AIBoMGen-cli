package generator

import (
	"encoding/json"
	"os"
	"path/filepath"

	"aibomgen-cra/internal/scanner"
)

// SimpleCycloneDX is a minimal structure mimicking CycloneDX for early output.
type SimpleCycloneDX struct {
	BomFormat   string            `json:"bomFormat"`
	SpecVersion string            `json:"specVersion"`
	Components  []SimpleComponent `json:"components"`
}

type SimpleComponent struct {
	Type       string            `json:"type"`
	Name       string            `json:"name"`
	BomRef     string            `json:"bom-ref"`
	Properties map[string]string `json:"properties,omitempty"`
}

// Build constructs a minimal BOM from scanner components.
func Build(components []scanner.Component) SimpleCycloneDX {
	bom := SimpleCycloneDX{
		BomFormat:   "CycloneDX",
		SpecVersion: "1.5", // forward-looking (actual full compliance later)
		Components:  []SimpleComponent{},
	}
	for _, c := range components {
		bom.Components = append(bom.Components, SimpleComponent{
			Type:   "library",
			Name:   c.Name,
			BomRef: c.Type + ":" + c.ID,
			Properties: map[string]string{
				"aibomgen.type":       c.Type,
				"aibomgen.path":       c.Path,
				"aibomgen.evidence":   c.Evidence,
				"aibomgen.confidence": formatFloat(c.Confidence),
			},
		})
	}
	return bom
}

func formatFloat(f float64) string {
	b, _ := json.Marshal(f)
	return string(b)
}

// Write writes the BOM to the given output path, creating directories as needed.
func Write(outputPath string, bom SimpleCycloneDX) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(bom, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, data, 0o644)
}
