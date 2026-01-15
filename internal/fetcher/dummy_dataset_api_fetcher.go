package fetcher

import (
	"context"
)

// DummyDatasetAPIFetcher returns a fixed DatasetAPIResponse for testing/demo purposes
type DummyDatasetAPIFetcher struct{}

// Fetch returns a dummy dataset response
func (f *DummyDatasetAPIFetcher) Fetch(ctx context.Context, datasetID string) (*DatasetAPIResponse, error) {
	return &DatasetAPIResponse{
		ID:          datasetID,
		Author:      "huggingface",
		Tags:        []string{"dataset", "benchmark"},
		Description: "Dummy dataset for testing: " + datasetID,
		Downloads:   100000,
		Likes:       500,
		CardData: map[string]any{
			"language": "en",
			"license":  "cc0-1.0",
		},
	}, nil
}
