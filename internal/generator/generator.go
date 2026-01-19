package generator

import (
	"context"
	"fmt"
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
	BuildDataset(builder.DatasetBuildContext) (*cdx.Component, error)
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

	// Extract datasets from model training data and build dataset components
	datasets := extractDatasetsFromModel(apiResp, readme)

	if len(datasets) > 0 {
		dummyDatasetApiFetcher := &fetcher.DummyDatasetAPIFetcher{}
		dummyDatasetReadmeFetcher := &fetcher.DummyDatasetReadmeFetcher{}

		if bom.Components == nil {
			bom.Components = &[]cdx.Component{}
		}

		for _, dsID := range datasets {
			dsApiResp, _ := dummyDatasetApiFetcher.Fetch(context.Background(), dsID)
			dsReadme, _ := dummyDatasetReadmeFetcher.Fetch(context.Background(), dsID)

			dsCtx := builder.DatasetBuildContext{
				DatasetID: dsID,
				Scan:      scanner.Discovery{ID: dsID, Name: dsID, Type: "dataset"},
				HF:        dsApiResp,
				Readme:    dsReadme,
			}

			dsComp, err := bomBuilder.BuildDataset(dsCtx)
			if err == nil && bom.Components != nil {
				*bom.Components = append(*bom.Components, *dsComp)
			}
		}
	}

	return []DiscoveredBOM{
		{
			Discovery: dummyDiscovery,
			BOM:       bom,
		},
	}, nil
}

// BuildPerDiscovery orchestrates: fetch HF API (optional) → build BOM per model via registry-driven builder.
// When building a model, if datasets are referenced in the model's training metadata, build dataset components too.
func BuildPerDiscovery(discoveries []scanner.Discovery, hfToken string, timeout time.Duration) ([]DiscoveredBOM, error) {
	results := make([]DiscoveredBOM, 0, len(discoveries))

	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	httpClient := &http.Client{Timeout: timeout}
	modelApiFetcher := &fetcher.ModelAPIFetcher{Client: httpClient, Token: hfToken}
	modelReadmeFetcher := &fetcher.ModelReadmeFetcher{Client: httpClient, Token: hfToken}
	datasetApiFetcher := &fetcher.DatasetAPIFetcher{Client: httpClient, Token: hfToken}
	datasetReadmeFetcher := &fetcher.DatasetReadmeFetcher{Client: httpClient, Token: hfToken}

	bomBuilder := newBOMBuilder()

	for _, d := range discoveries {
		modelID := strings.TrimSpace(d.ID)
		if modelID == "" {
			modelID = strings.TrimSpace(d.Name)
		}


		var resp *fetcher.ModelAPIResponse
		var readme *fetcher.ModelReadmeCard
		if modelID != "" {
			r, err := modelApiFetcher.Fetch(context.Background(), modelID)
			if err != nil {
			} else {
				resp = r
			}

			c, err := modelReadmeFetcher.Fetch(context.Background(), modelID)
			if err != nil {
			} else {
				readme = c
			}
		}

		ctx := builder.BuildContext{
			ModelID: modelID,
			Scan:    d,
			HF:      resp,
			Readme:  readme,
		}

		bom, err := bomBuilder.Build(ctx)
		if err != nil {
			return nil, err
		}

		// Extract datasets from model training data and build dataset components
		datasets := extractDatasetsFromModel(resp, readme)

		for _, dsID := range datasets {

			var dsResp *fetcher.DatasetAPIResponse
			var dsReadme *fetcher.DatasetReadmeCard

			r, err := datasetApiFetcher.Fetch(context.Background(), dsID)
			if err != nil {
				// In online mode, skip dataset components that don't exist on HuggingFace.
				// The dataset reference is still preserved in the model's modelCard.modelParameters.datasets.
				// Dummy fallbacks are only used when explicitly running in --hf-mode dummy.
				continue
			}
			dsResp = r

			c, err := datasetReadmeFetcher.Fetch(context.Background(), dsID)
			if err != nil {
				// Continue without readme - the dataset exists but may not have a readme
			} else {
				dsReadme = c
			}

			dsBomBuilder := newBOMBuilder()
			dsCtx := builder.DatasetBuildContext{
				DatasetID: dsID,
				Scan:      scanner.Discovery{ID: dsID, Name: dsID, Type: "dataset"},
				HF:        dsResp,
				Readme:    dsReadme,
			}

			dsComp, err := dsBomBuilder.BuildDataset(dsCtx)
			if err != nil {
				continue
			}

			// Only initialize components slice when we have an actual component to add
			if bom.Components == nil {
				bom.Components = &[]cdx.Component{}
			}
			*bom.Components = append(*bom.Components, *dsComp)
		}


		results = append(results, DiscoveredBOM{
			Discovery: d,
			BOM:       bom,
		})
	}

	return results, nil
}

// extractDatasetsFromModel extracts dataset IDs from model's training metadata
func extractDatasetsFromModel(modelResp *fetcher.ModelAPIResponse, readme *fetcher.ModelReadmeCard) []string {
	var datasets []string

	// Check model API response for datasets field
	if modelResp != nil && modelResp.CardData != nil {
		if datasetsVal, ok := modelResp.CardData["datasets"]; ok {
			// Could be a slice or a single value
			switch v := datasetsVal.(type) {
			case []interface{}:
				for _, item := range v {
					if dsID, ok := item.(string); ok && strings.TrimSpace(dsID) != "" {
						datasets = append(datasets, strings.TrimSpace(dsID))
					}
				}
			case string:
				if strings.TrimSpace(v) != "" {
					datasets = append(datasets, strings.TrimSpace(v))
				}
			}
		}
	}

	// Check readme for dataset references
	if readme != nil && readme.Datasets != nil {
		for _, dsID := range readme.Datasets {
			if strings.TrimSpace(dsID) != "" {
				datasets = append(datasets, strings.TrimSpace(dsID))
			}
		}
	}

	// Deduplicate
	if len(datasets) > 0 {
		seen := make(map[string]struct{})
		unique := make([]string, 0)
		for _, ds := range datasets {
			if _, ok := seen[ds]; !ok {
				seen[ds] = struct{}{}
				unique = append(unique, ds)
			}
		}
		return unique
	}

	return nil
}

// BuildFromModelIDs generates BOMs from one or more Hugging Face model IDs without scanning.
// This is used when --model-id is provided instead of --input.
func BuildFromModelIDs(modelIDs []string, hfToken string, timeout time.Duration) ([]DiscoveredBOM, error) {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	results := make([]DiscoveredBOM, 0, len(modelIDs))

	httpClient := &http.Client{Timeout: timeout}
	modelApiFetcher := &fetcher.ModelAPIFetcher{Client: httpClient, Token: hfToken}
	modelReadmeFetcher := &fetcher.ModelReadmeFetcher{Client: httpClient, Token: hfToken}
	datasetApiFetcher := &fetcher.DatasetAPIFetcher{Client: httpClient, Token: hfToken}
	datasetReadmeFetcher := &fetcher.DatasetReadmeFetcher{Client: httpClient, Token: hfToken}

	for _, modelID := range modelIDs {
		modelID = strings.TrimSpace(modelID)
		if modelID == "" {
			continue
		}

		bomBuilder := newBOMBuilder()


		resp, err := modelApiFetcher.Fetch(context.Background(), modelID)
		if err != nil {
			resp = nil
		} else {
		}

		readme, err := modelReadmeFetcher.Fetch(context.Background(), modelID)
		if err != nil {
			readme = nil
		} else {
		}

		// Create a discovery from the model ID
		discovery := scanner.Discovery{
			ID:       modelID,
			Name:     modelID,
			Type:     "huggingface",
			Path:     "",
			Evidence: fmt.Sprintf("from model-id: %s", modelID),
		}

		ctx := builder.BuildContext{
			ModelID: modelID,
			Scan:    discovery,
			HF:      resp,
			Readme:  readme,
		}

		bom, err := bomBuilder.Build(ctx)
		if err != nil {
			continue
		}

		// Extract datasets from model training data and build dataset components
		datasets := extractDatasetsFromModel(resp, readme)

		for _, dsID := range datasets {

			var dsResp *fetcher.DatasetAPIResponse
			var dsReadme *fetcher.DatasetReadmeCard

			r, err := datasetApiFetcher.Fetch(context.Background(), dsID)
			if err != nil {
				continue
			}
			dsResp = r

			c, err := datasetReadmeFetcher.Fetch(context.Background(), dsID)
			if err != nil {
			} else {
				dsReadme = c
			}

			dsBomBuilder := newBOMBuilder()
			dsCtx := builder.DatasetBuildContext{
				DatasetID: dsID,
				Scan:      scanner.Discovery{ID: dsID, Name: dsID, Type: "dataset"},
				HF:        dsResp,
				Readme:    dsReadme,
			}

			dsComp, err := dsBomBuilder.BuildDataset(dsCtx)
			if err != nil {
				continue
			}

			// Only initialize components slice when we have an actual component to add
			if bom.Components == nil {
				bom.Components = &[]cdx.Component{}
			}
			*bom.Components = append(*bom.Components, *dsComp)
		}


		results = append(results, DiscoveredBOM{
			Discovery: discovery,
			BOM:       bom,
		})
	}

	return results, nil
}
