package fetcher

type Metadata struct {
	// Basic model info
	ModelID     string
	Name        string
	Version     string
	License     string
	Author      string
	Description string
	SourceURL   string

	// Model Card (CycloneDX MLModelCard)
	Task               string
	ArchitectureFamily string
	ModelArchitecture  string
	ApproachType       string // e.g. "supervised", "unsupervised"
	Datasets           []DatasetInfo
	Inputs             []IOParameter
	Outputs            []IOParameter

	// Quantitative Analysis
	PerformanceMetrics []PerformanceMetric
	Graphics           []GraphicInfo

	// Considerations
	Users                       []string
	UseCases                    []string
	TechnicalLimitations        []string
	PerformanceTradeoffs        []string
	EthicalConsiderations       []EthicalConsideration
	EnvironmentalConsiderations EnvironmentalConsiderations
	FairnessAssessments         []FairnessAssessment

	// Energy/CO2
	EnergyConsumptions []EnergyConsumption

	// External references
	ExternalReferences []ExternalReference
}

// Supporting types
type DatasetInfo struct {
	Name        string
	Ref         string
	Description string
}

type IOParameter struct {
	Format string
}

type PerformanceMetric struct {
	Type               string
	Value              string
	Slice              string
	ConfidenceInterval *ConfidenceInterval
}

type ConfidenceInterval struct {
	LowerBound string
	UpperBound string
}

type GraphicInfo struct {
	URL         string
	Description string
}

type EthicalConsideration struct {
	Name               string
	MitigationStrategy string
}

type EnvironmentalConsiderations struct {
	EnergyConsumptions []EnergyConsumption
	Properties         []Property
}

type FairnessAssessment struct {
	GroupAtRisk        string
	Benefits           string
	Harms              string
	MitigationStrategy string
}

type EnergyConsumption struct {
	Activity           string
	EnergyProviders    []EnergyProvider
	ActivityEnergyCost EnergyMeasure
	CO2CostEquivalent  *CO2Measure
	CO2CostOffset      *CO2Measure
	Properties         []Property
}

type EnergyProvider struct {
	Description        string
	Organization       string
	EnergySource       string
	EnergyProvided     EnergyMeasure
	ExternalReferences []ExternalReference
}

type EnergyMeasure struct {
	Value float32
	Unit  string // e.g. "kWh"
}

type CO2Measure struct {
	Value float32
	Unit  string // e.g. "tCO2eq"
}

type Property struct {
	Name  string
	Value string
}

type ExternalReference struct {
	URL         string
	Description string
	Type        string
}
