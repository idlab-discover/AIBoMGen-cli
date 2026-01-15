package fetcher

import (
	"context"
)

// DummyDatasetReadmeFetcher returns a fixed DatasetReadmeCard for testing/demo purposes
type DummyDatasetReadmeFetcher struct{}

// Fetch returns a dummy dataset README card
func (f *DummyDatasetReadmeFetcher) Fetch(ctx context.Context, datasetID string) (*DatasetReadmeCard, error) {
	return &DatasetReadmeCard{
		Raw: "# " + datasetID + " Dataset\n\nDummy dataset card for testing.",
		FrontMatter: map[string]any{
			"license": "cc0-1.0",
			"tags":    []string{"dataset", "test"},
		},
		License:            "cc0-1.0",
		Tags:               []string{"dataset", "test"},
		Language:           []string{"en"},
		AnnotationCreators: []string{"huggingface"},
		CuratedBy:          "Dummy Curator",
		FundedBy:           "Test Foundation",
		SharedBy:           "Test Team",
		DatasetDescription: "A dummy dataset for testing dataset component building",
	}, nil
}
