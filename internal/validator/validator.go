package validator

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"aibomgen-cra/internal/generator"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

// ValidateBOM performs minimal and optional strict validation on a CycloneDX BOM.
func ValidateBOM(bom *cdx.BOM, strict bool) []string {
	var errs []string
	if bom == nil {
		return []string{"BOM is nil"}
	}
	if bom.Components == nil || len(*bom.Components) == 0 {
		errs = append(errs, "BOM has no components")
		return errs
	}
	for i := range *bom.Components {
		comp := &(*bom.Components)[i]
		if comp.Name == "" {
			errs = append(errs, fmt.Sprintf("component[%d]: name is required", i))
		}
		if strict {
			if comp.ModelCard == nil {
				errs = append(errs, fmt.Sprintf("component[%d] %q: missing modelCard (strict)", i, comp.Name))
				continue
			}
			card := comp.ModelCard
			if card.ModelParameters == nil || card.ModelParameters.Task == "" {
				errs = append(errs, fmt.Sprintf("component[%d] %q: missing task (strict)", i, comp.Name))
			}
			if card.ModelParameters == nil || card.ModelParameters.Inputs == nil || len(*card.ModelParameters.Inputs) == 0 {
				errs = append(errs, fmt.Sprintf("component[%d] %q: missing inputs (strict)", i, comp.Name))
			}
			if card.ModelParameters == nil || card.ModelParameters.Outputs == nil || len(*card.ModelParameters.Outputs) == 0 {
				errs = append(errs, fmt.Sprintf("component[%d] %q: missing outputs (strict)", i, comp.Name))
			}
			if card.QuantitativeAnalysis == nil || card.QuantitativeAnalysis.PerformanceMetrics == nil || len(*card.QuantitativeAnalysis.PerformanceMetrics) == 0 {
				errs = append(errs, fmt.Sprintf("component[%d] %q: missing performance metrics (strict)", i, comp.Name))
			}
			if card.Considerations == nil || card.Considerations.Users == nil || len(*card.Considerations.Users) == 0 {
				errs = append(errs, fmt.Sprintf("component[%d] %q: missing intended users (strict)", i, comp.Name))
			}
			if card.Considerations == nil || card.Considerations.UseCases == nil || len(*card.Considerations.UseCases) == 0 {
				errs = append(errs, fmt.Sprintf("component[%d] %q: missing use cases (strict)", i, comp.Name))
			}
		}
	}
	return errs
}

// ValidateFromFile decodes (JSON or XML) then validates a BOM. format can be "json", "xml" or "auto".
func ValidateFromFile(path string, strict bool, format string) (*cdx.BOM, []string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	bom, err := decodeBOM(f, path, format)
	if err != nil {
		return nil, nil, err
	}
	errs := ValidateBOM(bom, strict)
	return bom, errs, nil
}

func decodeBOM(r io.Reader, filename, format string) (*cdx.BOM, error) {
	actual := format
	if actual == "" || actual == "auto" {
		// First try extension
		ext := strings.ToLower(filepath.Ext(filename))
		switch ext {
		case ".xml":
			actual = "xml"
		case ".json":
			actual = "json"
		}
		if actual == "" { // sniff content
			br := bufio.NewReader(r)
			// Peek a reasonable amount
			peek, _ := br.Peek(64)
			s := strings.TrimSpace(string(peek))
			if strings.HasPrefix(s, "<") {
				actual = "xml"
			} else {
				actual = "json"
			}
			r = br
		}
	}
	fileFmt := cdx.BOMFileFormatJSON
	if actual == "xml" {
		fileFmt = cdx.BOMFileFormatXML
	}
	bom := new(cdx.BOM)
	dec := cdx.NewBOMDecoder(r, fileFmt)
	if err := dec.Decode(bom); err != nil {
		return nil, err
	}
	return bom, nil
}

// ValidateSpecVersion ensures the BOM's declared SpecVersion matches expected.
// Returns a slice of errors (empty when OK or when expected is blank).
// ! currently does not yet check fields based on spec version. Only generator will delete unsupported fields. !
func ValidateSpecVersion(bom *cdx.BOM, expected string) []string {
	if expected == "" {
		return nil
	}
	if bom == nil {
		return []string{"BOM is nil"}
	}
	exp, ok := generator.ParseSpecVersion(expected)
	if !ok {
		return []string{fmt.Sprintf("unsupported CycloneDX specVersion: %q", expected)}
	}
	if bom.SpecVersion == 0 {
		return []string{"BOM missing specVersion"}
	}
	if bom.SpecVersion != exp {
		return []string{fmt.Sprintf("specVersion mismatch: expected %s, got %s", exp.String(), bom.SpecVersion.String())}
	}
	return nil
}
