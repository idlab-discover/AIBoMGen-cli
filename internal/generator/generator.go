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
		cdxComp := cdx.Component{
			Type:      cdx.ComponentTypeMachineLearningModel,
			Name:      comp.Name,
			Version:   "", // No version info from scanner
			ModelCard: BuildMLModelCard(comp.Name),
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

// Example: fill MLModelCard for a detected model
func BuildMLModelCard(modelID string) *cdx.MLModelCard {
	// Fetch metadata from Hugging Face or local cache
	meta := fetcher.FetchModelMetadata(modelID)

	// Map Metadata fields to CycloneDX MLModelCard
	// Datasets
	var datasets []cdx.MLDatasetChoice
	for _, ds := range meta.Datasets {
		datasets = append(datasets, cdx.MLDatasetChoice{
			Ref: ds.Ref,
			ComponentData: &cdx.ComponentData{
				Name:        ds.Name,
				Description: ds.Description,
			},
		})
	}
	// Inputs/Outputs
	var inputs, outputs []cdx.MLInputOutputParameters
	for _, in := range meta.Inputs {
		inputs = append(inputs, cdx.MLInputOutputParameters{Format: in.Format})
	}
	for _, out := range meta.Outputs {
		outputs = append(outputs, cdx.MLInputOutputParameters{Format: out.Format})
	}
	// Performance metrics
	var perfMetrics []cdx.MLPerformanceMetric
	for _, pm := range meta.PerformanceMetrics {
		var ci *cdx.MLPerformanceMetricConfidenceInterval
		if pm.ConfidenceInterval != nil {
			ci = &cdx.MLPerformanceMetricConfidenceInterval{
				LowerBound: pm.ConfidenceInterval.LowerBound,
				UpperBound: pm.ConfidenceInterval.UpperBound,
			}
		}
		perfMetrics = append(perfMetrics, cdx.MLPerformanceMetric{
			Type:               pm.Type,
			Value:              pm.Value,
			Slice:              pm.Slice,
			ConfidenceInterval: ci,
		})
	}
	// Considerations
	var ethicals []cdx.MLModelCardEthicalConsideration
	for _, e := range meta.EthicalConsiderations {
		ethicals = append(ethicals, cdx.MLModelCardEthicalConsideration{
			Name:               e.Name,
			MitigationStrategy: e.MitigationStrategy,
		})
	}
	var fairness []cdx.MLModelCardFairnessAssessment
	for _, f := range meta.FairnessAssessments {
		fairness = append(fairness, cdx.MLModelCardFairnessAssessment{
			GroupAtRisk:        f.GroupAtRisk,
			Benefits:           f.Benefits,
			Harms:              f.Harms,
			MitigationStrategy: f.MitigationStrategy,
		})
	}
	// Environmental considerations
	var env *cdx.MLModelCardEnvironmentalConsiderations
	if len(meta.EnvironmentalConsiderations.EnergyConsumptions) > 0 || len(meta.EnvironmentalConsiderations.Properties) > 0 {
		env = &cdx.MLModelCardEnvironmentalConsiderations{}
		// EnergyConsumptions and Properties mapping omitted for brevity
	}

	return &cdx.MLModelCard{
		ModelParameters: &cdx.MLModelParameters{
			Task:               meta.Task,
			ArchitectureFamily: meta.ArchitectureFamily,
			ModelArchitecture:  meta.ModelArchitecture,
			Approach: &cdx.MLModelParametersApproach{
				Type: cdx.MLModelParametersApproachType(meta.ApproachType),
			},
			Datasets: &datasets,
			Inputs:   &inputs,
			Outputs:  &outputs,
		},
		QuantitativeAnalysis: &cdx.MLQuantitativeAnalysis{
			PerformanceMetrics: &perfMetrics,
			// Graphics mapping omitted for brevity
		},
		Considerations: &cdx.MLModelCardConsiderations{
			Users:                       &meta.Users,
			UseCases:                    &meta.UseCases,
			TechnicalLimitations:        &meta.TechnicalLimitations,
			PerformanceTradeoffs:        &meta.PerformanceTradeoffs,
			EthicalConsiderations:       &ethicals,
			EnvironmentalConsiderations: env,
			FairnessAssessments:         &fairness,
		},
	}
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
