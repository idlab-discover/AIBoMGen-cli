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

// ComponentBOM couples the original scanner component with a single-component BOM.
type ComponentBOM struct {
	Component scanner.Component
	BOM       *cdx.BOM
}

// Build constructs a minimal BOM from scanner components.
func Build(components []scanner.Component) *cdx.BOM {
	var cdxComponents []cdx.Component
	for i, comp := range components {
		logComponentContext(i, comp)
		cdxComponents = append(cdxComponents, buildComponent(comp))
	}
	bom := cdx.NewBOM()
	if len(cdxComponents) > 0 {
		bom.Components = &cdxComponents
	}
	return bom
}

// BuildPerComponent produces individual BOMs for each component to support per-model output.
func BuildPerComponent(components []scanner.Component) []ComponentBOM {
	results := make([]ComponentBOM, 0, len(components))
	for i, comp := range components {
		logComponentContext(i, comp)
		component := buildComponent(comp)
		bom := cdx.NewBOM()
		bom.Components = &[]cdx.Component{component}
		results = append(results, ComponentBOM{Component: comp, BOM: bom})
	}
	return results
}

func logComponentContext(idx int, comp scanner.Component) {
	if logOut == nil {
		return
	}
	if idx > 0 {
		fmt.Fprintln(logOut)
	}
	prefix := ui.Color("Model context:", ui.FgCyan)
	fmt.Fprintf(logOut, "%s %s (path: %s)\n", prefix, comp.Name, comp.Path)
}

func buildComponent(comp scanner.Component) cdx.Component {
	modelCard := BuildMLModelCard(comp.Name)
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
		Version:    "",
		ModelCard:  modelCard,
		Properties: &props,
	}
	cdxComp.ExternalReferences = &[]cdx.ExternalReference{{
		Type: cdx.ExternalReferenceType("website"),
		URL:  "https://huggingface.co/" + comp.Name,
	}}
	return cdxComp
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
