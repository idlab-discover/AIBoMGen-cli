package parameters

import cdx "github.com/CycloneDX/cyclonedx-go"

// DummyParametersFetcher provides placeholder model parameters for testing/demo.
type DummyParametersFetcher struct{}

func NewDummyParametersFetcher() *DummyParametersFetcher { return &DummyParametersFetcher{} }

func (s *DummyParametersFetcher) Get(id string) (*cdx.MLModelParameters, error) {
	datasets := []cdx.MLDatasetChoice{
		{Ref: "dataset:bookcorpus"},
		{Ref: "dataset:wikipedia-en"},
	}
	inputs := []cdx.MLInputOutputParameters{{Format: "text/plain"}}
	outputs := []cdx.MLInputOutputParameters{{Format: "classification-label"}}
	return &cdx.MLModelParameters{
		Task:               "text-classification",
		ArchitectureFamily: "Transformer",
		ModelArchitecture:  "BERT",
		Approach:           &cdx.MLModelParametersApproach{Type: cdx.MLModelParametersApproachType("supervised")},
		Datasets:           &datasets,
		Inputs:             &inputs,
		Outputs:            &outputs,
	}, nil
}
