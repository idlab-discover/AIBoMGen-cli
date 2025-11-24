package fetcher

// Placeholder for future metadata fetching (e.g., Hugging Face Hub API).
// Provides interface definition for future expansion.

// Metadata represents external model metadata.
type Metadata struct {
	ID           string
	License      string
	ModelCardURL string
	Source       string
}

// Fetcher defines behavior for retrieving model metadata.
type Fetcher interface {
	Get(id string) (Metadata, error)
}
