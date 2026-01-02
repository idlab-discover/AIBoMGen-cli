package io

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

// ReadBOM reads a BOM from a file (JSON or XML).
// The format parameter can be "json", "xml", or "auto" (default).
// If "auto", the format is determined from the file extension.
func ReadBOM(path string, format string) (*cdx.BOM, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	actual := strings.ToLower(strings.TrimSpace(format))
	if actual == "" || actual == "auto" {
		if strings.EqualFold(filepath.Ext(path), ".xml") {
			actual = "xml"
		} else {
			actual = "json"
		}
	}

	fileFmt := cdx.BOMFileFormatJSON
	if actual == "xml" {
		fileFmt = cdx.BOMFileFormatXML
	}

	bom := new(cdx.BOM)
	dec := cdx.NewBOMDecoder(f, fileFmt)
	if err := dec.Decode(bom); err != nil {
		return nil, err
	}

	return bom, nil
}

// WriteBOM writes a BOM to a file in the specified format.
// The format parameter can be "json", "xml", or "auto" (default).
// If "auto", the format is determined from the file extension.
// If spec is provided, it encodes with that specific CycloneDX version.
func WriteBOM(bom *cdx.BOM, outputPath string, format string, spec string) error {
	ext := filepath.Ext(outputPath)

	actual := strings.ToLower(strings.TrimSpace(format))
	if actual == "" || actual == "auto" {
		if strings.EqualFold(ext, ".xml") {
			actual = "xml"
		} else {
			actual = "json"
		}
	}

	// Validate extension matches format
	if actual != "auto" {
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

	sv, ok := ParseSpecVersion(spec)
	if !ok {
		return fmt.Errorf("unsupported CycloneDX spec version: %q", spec)
	}
	return encoder.EncodeVersion(bom, sv)
}

// ParseSpecVersion parses a spec version string to a CycloneDX SpecVersion.
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
		return cdx.SpecVersion1_6, false
	}
}
