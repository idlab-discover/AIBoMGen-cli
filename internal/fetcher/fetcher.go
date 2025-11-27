package fetcher

import cdx "github.com/CycloneDX/cyclonedx-go"

// Fetcher defines behavior for retrieving model cards.
type Fetcher interface {
	Get(id string) (*cdx.MLModelCard, error)
}

// Default is the fetcher used at runtime. It can be overridden by the CLI.
// It is nil until the CLI configures it.
var Default Fetcher

// SetDefault overrides the default fetcher used by generator.
func SetDefault(f Fetcher) {
	if f != nil {
		Default = f
	}
}

// FetchModelCard resolves the model card using the default fetcher.
// If Default is nil or returns an error, an empty MLModelCard is returned.
func FetchModelCard(modelID string) *cdx.MLModelCard {
	if Default != nil {
		if card, err := Default.Get(modelID); err == nil && card != nil {
			return card
		}
	}
	return &cdx.MLModelCard{}
}
