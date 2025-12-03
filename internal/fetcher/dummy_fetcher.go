package fetcher

import (
	cdx "github.com/CycloneDX/cyclonedx-go"
)

// DummyFetcher returns an empty-but-valid MLModelCard with placeholders.
type DummyFetcher struct{}

func NewDummyFetcher() *DummyFetcher { return &DummyFetcher{} }

func (d *DummyFetcher) Get(id string) (*cdx.MLModelCard, error) {
	// Full placeholder card: populate all fields with sensible defaults
	inputs := []cdx.MLInputOutputParameters{{Format: "application/octet-stream"}}
	outputs := []cdx.MLInputOutputParameters{{Format: "application/octet-stream"}}
	datasets := []cdx.MLDatasetChoice{{Ref: "dataset:unknown"}}

	mp := &cdx.MLModelParameters{
		Approach:           &cdx.MLModelParametersApproach{Type: cdx.MLModelParametersApproachTypeSupervised},
		Task:               "unknown",
		ArchitectureFamily: "unknown",
		ModelArchitecture:  "unknown",
		Datasets:           &datasets,
		Inputs:             &inputs,
		Outputs:            &outputs,
	}

	// QuantitativeAnalysis with placeholder metrics and empty graphics
	metrics := []cdx.MLPerformanceMetric{
		{Type: "accuracy", Value: "0.0", Slice: "overall", ConfidenceInterval: &cdx.MLPerformanceMetricConfidenceInterval{LowerBound: "0.0", UpperBound: "0.0"}},
		{Type: "f1", Value: "0.0", Slice: "overall", ConfidenceInterval: &cdx.MLPerformanceMetricConfidenceInterval{LowerBound: "0.0", UpperBound: "0.0"}},
	}
	qa := &cdx.MLQuantitativeAnalysis{
		PerformanceMetrics: &metrics,
		Graphics:           &cdx.ComponentDataGraphics{Description: "placeholder"},
	}

	// Considerations: populate all subfields
	users := []string{"unknown-user"}
	useCases := []string{"unknown-use-case"}
	techLimits := []string{"unknown-limitation"}
	perfTradeoffs := []string{"unknown-tradeoff"}
	ethics := []cdx.MLModelCardEthicalConsideration{{Name: "unknown-ethics", MitigationStrategy: "unknown-mitigation"}}
	energyConsumptions := []cdx.MLModelEnergyConsumption{{
		Activity:           cdx.MLModelEnergyConsumptionActivityOther,
		EnergyProviders:    &[]cdx.MLModelEnergyProvider{},
		ActivityEnergyCost: cdx.MLModelEnergyMeasure{Value: 0.0, Unit: cdx.MLModelEnergyUnitKWH},
		CO2CostEquivalent:  &cdx.MLModelCO2Measure{Value: 0.0, Unit: cdx.MLModelCO2UnitTCO2Eq},
		CO2CostOffset:      &cdx.MLModelCO2Measure{Value: 0.0, Unit: cdx.MLModelCO2UnitTCO2Eq},
		Properties:         &[]cdx.Property{{Name: "placeholder", Value: "true"}},
	}}
	env := &cdx.MLModelCardEnvironmentalConsiderations{
		EnergyConsumptions: &energyConsumptions,
		Properties:         &[]cdx.Property{{Name: "placeholder", Value: "true"}},
	}
	fairness := []cdx.MLModelCardFairnessAssessment{{
		GroupAtRisk:        "unknown-group",
		Benefits:           "unknown",
		Harms:              "unknown",
		MitigationStrategy: "unknown",
	}}

	cons := &cdx.MLModelCardConsiderations{
		Users:                       &users,
		UseCases:                    &useCases,
		TechnicalLimitations:        &techLimits,
		PerformanceTradeoffs:        &perfTradeoffs,
		EthicalConsiderations:       &ethics,
		EnvironmentalConsiderations: env,
		FairnessAssessments:         &fairness,
	}

	card := &cdx.MLModelCard{
		ModelParameters:      mp,
		QuantitativeAnalysis: qa,
		Considerations:       cons,
	}
	return card, nil
}
