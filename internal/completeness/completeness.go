package completeness

import (
	"github.com/idlab-discover/AIBoMGen-cli/internal/metadata"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

type Report struct {
	Score float64 // 0..1

	Passed int
	Total  int

	MissingRequired []metadata.Key
	MissingOptional []metadata.Key

	// Dataset-specific tracking
	DatasetReports map[string]DatasetReport // key is dataset name/ref
}

type DatasetReport struct {
	DatasetRef string // Reference to the dataset

	Score  float64 // 0..1
	Passed int
	Total  int

	MissingRequired []metadata.DatasetKey
	MissingOptional []metadata.DatasetKey
}

func Check(bom *cdx.BOM) Report {
	var (
		earned, max float64
		passed      int
		total       int
		missingReq  []metadata.Key
		missingOpt  []metadata.Key
	)

	// Check if datasets are referenced in model
	datasetsReferenced := hasDatasetsReferenced(bom)

	for _, spec := range metadata.Registry() {
		if spec.Weight <= 0 {
			continue
		}

		// Skip dataset field if no datasets are referenced
		if spec.Key == metadata.ModelCardModelParametersDatasets && !datasetsReferenced {
			// Only count as missing if no datasets are referenced
			total++
			max += spec.Weight
			if spec.Required {
				missingReq = append(missingReq, spec.Key)
			} else {
				missingOpt = append(missingOpt, spec.Key)
			}
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

	report := Report{
		Score:           score,
		Passed:          passed,
		Total:           total,
		MissingRequired: missingReq,
		MissingOptional: missingOpt,
		DatasetReports:  make(map[string]DatasetReport),
	}

	// Check dataset components if they exist
	if bom.Components != nil && datasetsReferenced {
		for _, comp := range *bom.Components {
			if comp.Type == cdx.ComponentTypeData {
				dsReport := CheckDataset(&comp)
				report.DatasetReports[comp.Name] = dsReport
			}
		}
	}

	return report
}

// hasDatasetsReferenced checks if the model references any datasets
func hasDatasetsReferenced(bom *cdx.BOM) bool {
	if bom == nil || bom.Metadata == nil || bom.Metadata.Component == nil {
		return false
	}
	comp := bom.Metadata.Component
	if comp.ModelCard == nil || comp.ModelCard.ModelParameters == nil {
		return false
	}
	mp := comp.ModelCard.ModelParameters
	if mp.Datasets == nil || len(*mp.Datasets) == 0 {
		return false
	}
	// Check if any dataset ref is non-empty
	for _, ds := range *mp.Datasets {
		if ds.Ref != "" {
			return true
		}
	}
	return false
}

// CheckDataset checks completeness of a single dataset component
func CheckDataset(comp *cdx.Component) DatasetReport {
	var (
		earned, max float64
		passed      int
		total       int
		missingReq  []metadata.DatasetKey
		missingOpt  []metadata.DatasetKey
	)

	for _, spec := range metadata.DatasetRegistry() {
		if spec.Weight <= 0 {
			continue
		}
		total++
		max += spec.Weight

		ok := false
		if spec.Present != nil {
			ok = spec.Present(comp)
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

	return DatasetReport{
		DatasetRef:      comp.Name,
		Score:           score,
		Passed:          passed,
		Total:           total,
		MissingRequired: missingReq,
		MissingOptional: missingOpt,
	}
}
