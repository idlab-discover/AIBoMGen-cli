package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// DatasetAPIResponse is the decoded response from GET https://huggingface.co/api/datasets/:id
type DatasetAPIResponse struct {
	ID          string         `json:"id"`
	Author      string         `json:"author"`
	SHA         string         `json:"sha"`
	LastMod     string         `json:"lastModified"`
	CreatedAt   string         `json:"createdAt"`
	Private     bool           `json:"private"`
	Gated       BoolOrString   `json:"gated"`
	Disabled    bool           `json:"disabled"`
	Tags        []string       `json:"tags"`
	Description string         `json:"description"`
	Downloads   int            `json:"downloads"`
	Likes       int            `json:"likes"`
	UsedStorage int64          `json:"usedStorage"`
	CardData    map[string]any `json:"cardData"`
}

// DatasetAPIFetcher fetches dataset metadata from the Hugging Face Hub API.
type DatasetAPIFetcher struct {
	Client  *http.Client
	Token   string
	BaseURL string // optional; defaults to "https://huggingface.co"
}

// Fetch fetches dataset metadata for the given datasetID.
func (f *DatasetAPIFetcher) Fetch(ctx context.Context, datasetID string) (*DatasetAPIResponse, error) {
	client := f.Client
	if client == nil {
		client = http.DefaultClient
	}

	trimmedDatasetID := strings.TrimPrefix(strings.TrimSpace(datasetID), "/")
	logf(datasetID, "GET /api/datasets/%s", trimmedDatasetID)

	baseURL := strings.TrimRight(strings.TrimSpace(f.BaseURL), "/")
	if baseURL == "" {
		baseURL = "https://huggingface.co"
	}

	url := fmt.Sprintf("%s/api/datasets/%s", baseURL, trimmedDatasetID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if strings.TrimSpace(f.Token) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(f.Token))
	}

	resp, err := client.Do(req)
	if err != nil {
		logf(datasetID, "request error (%v)", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logf(datasetID, "non-200 status=%d", resp.StatusCode)
		return nil, fmt.Errorf("huggingface api status %d", resp.StatusCode)
	}

	var parsed DatasetAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		logf(datasetID, "decode error (%v)", err)
		return nil, err
	}
	logf(datasetID, "ok")
	return &parsed, nil
}
