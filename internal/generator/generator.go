package generator

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
func Write(outputPath string, bom *cdx.BOM) error { return WriteWithFormat(outputPath, bom, "json") }

// WriteWithFormat writes the BOM in the specified format (json|xml). If format is auto, infer from extension.
// It preserves the BOM's current spec version.
func WriteWithFormat(outputPath string, bom *cdx.BOM, format string) error {
	return WriteWithFormatAndSpec(outputPath, bom, format, "")
}

// WriteWithFormatAndSpec writes the BOM with the specified file format and optional spec version.
// If spec is a non-empty string (e.g., "1.3"), the BOM is encoded using EncodeVersion which may drop
// fields not supported by that spec. If spec is empty, the BOM is encoded as-is.
func WriteWithFormatAndSpec(outputPath string, bom *cdx.BOM, format string, spec string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	actual := format
	if actual == "auto" || actual == "" {
		ext := strings.ToLower(filepath.Ext(outputPath))
		if ext == ".xml" {
			actual = "xml"
		} else {
			actual = "json"
		}
	}
	// Enforce extension/format consistency when extension present and format explicitly set
	ext := strings.ToLower(filepath.Ext(outputPath))
	if actual != "auto" && ext != "" {
		switch actual {
		case "xml":
			if ext != ".xml" {
				return fmt.Errorf("output path extension %q does not match format %q", ext, actual)
			}
		case "json":
			if ext != ".json" {
				return fmt.Errorf("output path extension %q does not match format %q", ext, actual)
			}
		}
	}
	fileFmt := cdx.BOMFileFormatJSON
	if actual == "xml" {
		fileFmt = cdx.BOMFileFormatXML
	}
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	encoder := cdx.NewBOMEncoder(f, fileFmt)
	encoder.SetPretty(true)
	if spec == "" {
		return encoder.Encode(bom)
	}
	// Parse spec string into cyclonedx SpecVersion
	sv, ok := ParseSpecVersion(spec)
	if !ok {
		return fmt.Errorf("unsupported CycloneDX spec version: %q", spec)
	}
	return encoder.EncodeVersion(bom, sv)
}

var logOut io.Writer

// SetLogger sets an optional logger to provide context around fetcher logs.
func SetLogger(w io.Writer) { logOut = w }

func ParseSpecVersion(s string) (cdx.SpecVersion, bool) {
	switch s {
	case "1.0":
		return cdx.SpecVersion1_0, true
	case "1.1":
		return cdx.SpecVersion1_1, true
	case "1.2":
		return cdx.SpecVersion1_2, true
	case "1.3":
		return cdx.SpecVersion1_3, true
	case "1.4":
		return cdx.SpecVersion1_4, true
	case "1.5":
		return cdx.SpecVersion1_5, true
	case "1.6":
		return cdx.SpecVersion1_6, true
	default:
		return 0, false
	}
}
