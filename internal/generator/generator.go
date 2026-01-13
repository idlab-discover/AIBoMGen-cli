package generator

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/idlab-discover/AIBoMGen-cli/internal/builder"
	"github.com/idlab-discover/AIBoMGen-cli/internal/fetcher"
	"github.com/idlab-discover/AIBoMGen-cli/internal/scanner"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

type DiscoveredBOM struct {
	Discovery scanner.Discovery
	BOM       *cdx.BOM
}

type bomBuilder interface {
	Build(builder.BuildContext) (*cdx.BOM, error)
}

var newBOMBuilder = func() bomBuilder {
	return builder.NewBOMBuilder(builder.DefaultOptions())
}

// BuildDummyBOM builds a single comprehensive dummy BOM with all fields populated.
// This is used in dummy mode for testing/demo purposes without scanning or fetching real data.
func BuildDummyBOM() ([]DiscoveredBOM, error) {
	// Create dummy fetchers that return fixed responses
	apiFetcher := &fetcher.DummyModelAPIFetcher{}
	readmeFetcher := &fetcher.DummyModelReadmeFetcher{}

	// Create a dummy discovery
	dummyDiscovery := scanner.Discovery{
		ID:       "dummy-org/dummy-model",
		Name:     "dummy-model",
		Type:     "huggingface",
		Path:     "/dummy/path",
		Evidence: "from_pretrained('dummy-org/dummy-model')",
	}

	// Fetch dummy metadata
	apiResp, err := apiFetcher.Fetch(context.Background(), "dummy-org/dummy-model")
	if err != nil {
		return nil, err
	}

	readme, err := readmeFetcher.Fetch(context.Background(), "dummy-org/dummy-model")
	if err != nil {
		return nil, err
	}

	// Build the BOM with all dummy data
	ctx := builder.BuildContext{
		ModelID: "dummy-org/dummy-model",
		Scan:    dummyDiscovery,
		HF:      apiResp,
		Readme:  readme,
	}

	bomBuilder := newBOMBuilder()
	bom, err := bomBuilder.Build(ctx)
	if err != nil {
		return nil, err
	}

	return []DiscoveredBOM{
		{
			Discovery: dummyDiscovery,
			BOM:       bom,
		},
	}, nil
}

// BuildPerDiscovery orchestrates: fetch HF API (optional) → build BOM per model via registry-driven builder.
func BuildPerDiscovery(discoveries []scanner.Discovery, hfToken string, timeout time.Duration) ([]DiscoveredBOM, error) {
	results := make([]DiscoveredBOM, 0, len(discoveries))

	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	httpClient := &http.Client{Timeout: timeout}
	apiFetcher := &fetcher.ModelAPIFetcher{Client: httpClient, Token: hfToken}
	readmeFetcher := &fetcher.ModelReadmeFetcher{Client: httpClient, Token: hfToken}

	bomBuilder := newBOMBuilder()

	for _, d := range discoveries {
		modelID := strings.TrimSpace(d.ID)
		if modelID == "" {
			modelID = strings.TrimSpace(d.Name)
		}

		logf(modelID, "start (scanPath=%s)", strings.TrimSpace(d.Path))

		var resp *fetcher.ModelAPIResponse
		var readme *fetcher.ModelReadmeCard
		if modelID != "" {
			logf(modelID, "fetch HF model metadata")
			r, err := apiFetcher.Fetch(context.Background(), modelID)
			if err != nil {
				logf(modelID, "fetch failed (%v)", err)
			} else {
				resp = r
				logf(modelID, "metadata fetched")
			}

			logf(modelID, "fetch HF README model card")
			c, err := readmeFetcher.Fetch(context.Background(), modelID)
			if err != nil {
				logf(modelID, "readme fetch failed (%v)", err)
			} else {
				readme = c
				logf(modelID, "readme fetched")
			}
		}

		ctx := builder.BuildContext{
			ModelID: modelID,
			Scan:    d,
			HF:      resp,
			Readme:  readme,
		}

		logf(modelID, "build BOM")
		bom, err := bomBuilder.Build(ctx)
		if err != nil {
			return nil, err
		}
		logf(modelID, "done")

		results = append(results, DiscoveredBOM{
			Discovery: d,
			BOM:       bom,
		})
	}

	return results, nil
}
