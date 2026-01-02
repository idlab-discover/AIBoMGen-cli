package completeness

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/idlab-discover/AIBoMGen-cli/internal/metadata"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

type Report struct {
	Score float64 // 0..1

	Passed int
	Total  int

	MissingRequired []metadata.Key
	MissingOptional []metadata.Key
}

func Check(bom *cdx.BOM) Report {
	var (
		earned, max float64
		passed      int
		total       int
		missingReq  []metadata.Key
		missingOpt  []metadata.Key
	)

	for _, spec := range metadata.Registry() {
		if spec.Weight <= 0 {
			continue
		}
		total++
		max += spec.Weight

		ok := false
		if spec.Present != nil {
			ok = spec.Present(bom)
		}

		if ok {
			passed++
			earned += spec.Weight
			continue
		}

		if spec.Required {
			missingReq = append(missingReq, spec.Key)
		} else {
			missingOpt = append(missingOpt, spec.Key)
		}
	}

	score := 0.0
	if max > 0 {
		score = earned / max
	}

	return Report{
		Score:           score,
		Passed:          passed,
		Total:           total,
		MissingRequired: missingReq,
		MissingOptional: missingOpt,
	}
}

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
