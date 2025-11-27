package fetcher

import cdx "github.com/CycloneDX/cyclonedx-go"

// Placeholder for future metadata fetching (e.g., Hugging Face Hub API).
// Provides interface definition for future expansion.

// Fetcher defines behavior for retrieving model cards.
type Fetcher interface {
	Get(id string) (*cdx.MLModelCard, error)
}

// FetchModelCard returns a dummy CycloneDX MLModelCard directly, avoiding custom metadata structs.
func FetchModelCard(modelID string) *cdx.MLModelCard {
	datasets := []cdx.MLDatasetChoice{
		{
			Ref: "dataset:bookcorpus",
			ComponentData: &cdx.ComponentData{
				Name:        "BookCorpus",
				Description: "Large dataset of books.",
			},
		},
		{
			Ref: "dataset:wikipedia-en",
			ComponentData: &cdx.ComponentData{
				Name:        "Wikipedia",
				Description: "English Wikipedia dump.",
			},
		},
	}
	inputs := []cdx.MLInputOutputParameters{{Format: "text/plain"}}
	outputs := []cdx.MLInputOutputParameters{{Format: "classification-label"}}
	perf := []cdx.MLPerformanceMetric{{
		Type:  "accuracy",
		Value: "0.84",
		Slice: "dev",
		ConfidenceInterval: &cdx.MLPerformanceMetricConfidenceInterval{
			LowerBound: "0.82",
			UpperBound: "0.86",
		},
	}}

	return &cdx.MLModelCard{
		ModelParameters: &cdx.MLModelParameters{
			Task:               "text-classification",
			ArchitectureFamily: "Transformer",
			ModelArchitecture:  "BERT",
			Approach: &cdx.MLModelParametersApproach{
				Type: cdx.MLModelParametersApproachType("supervised"),
			},
			Datasets: &datasets,
			Inputs:   &inputs,
			Outputs:  &outputs,
		},
		QuantitativeAnalysis: &cdx.MLQuantitativeAnalysis{
			PerformanceMetrics: &perf,
		},
		Considerations: &cdx.MLModelCardConsiderations{
			Users:                &[]string{"NLP researchers", "Developers"},
			UseCases:             &[]string{"Sentiment analysis", "Intent classification"},
			TechnicalLimitations: &[]string{"Not suitable for non-English text"},
			PerformanceTradeoffs: &[]string{"Large model size increases inference time"},
			EthicalConsiderations: &[]cdx.MLModelCardEthicalConsideration{{
				Name:               "Bias in training data",
				MitigationStrategy: "Careful dataset curation",
			}},
		},
		// Optional: attach external references via component properties in generator
	}
}
