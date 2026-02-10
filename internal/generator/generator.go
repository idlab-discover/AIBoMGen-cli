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

// ProgressCallback is called during generation to report progress
type ProgressCallback func(event ProgressEvent)

// ProgressEvent represents a progress update
type ProgressEvent struct {
	Type     ProgressEventType
	ModelID  string
	Message  string
	Index    int
	Total    int
	Datasets int
	Error    error
}

// ProgressEventType identifies the type of progress event
type ProgressEventType int

const (
	EventScanStart ProgressEventType = iota
	EventScanComplete
	EventFetchStart
	EventFetchAPIComplete
	EventFetchReadmeComplete
	EventBuildStart
	EventBuildComplete
	EventDatasetStart
	EventDatasetComplete
	EventModelComplete
	EventError
)

// GenerateOptions configures the generation process
type GenerateOptions struct {
	HFToken    string
	Timeout    time.Duration
	OnProgress ProgressCallback
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

	// Add dependencies from model to datasets
	builder.AddDependencies(bom)

	return []DiscoveredBOM{
		{
			Discovery: dummyDiscovery,
			BOM:       bom,
		},
	}, nil
}

// BuildPerDiscovery is a convenience wrapper that generates BOMs without progress reporting.
// For progress updates, use BuildPerDiscoveryWithProgress.
func BuildPerDiscovery(discoveries []scanner.Discovery, hfToken string, timeout time.Duration) ([]DiscoveredBOM, error) {
	return BuildPerDiscoveryWithProgress(context.Background(), discoveries, GenerateOptions{
		HFToken:    hfToken,
		Timeout:    timeout,
		OnProgress: nil,
	})
}

// BuildPerDiscoveryWithProgress orchestrates BOM generation with progress reporting.
// Fetches HF API metadata → builds BOM per model via registry-driven builder.
// When building a model, if datasets are referenced in the model's training metadata, builds dataset components too.
func BuildPerDiscoveryWithProgress(ctx context.Context, discoveries []scanner.Discovery, opts GenerateOptions) ([]DiscoveredBOM, error) {
	if opts.Timeout <= 0 {
		opts.Timeout = 10 * time.Second
	}

	progress := opts.OnProgress
	if progress == nil {
		progress = func(ProgressEvent) {}
	}

	results := make([]DiscoveredBOM, 0, len(discoveries))

	httpClient := &http.Client{Timeout: opts.Timeout}
	modelApiFetcher := &fetcher.ModelAPIFetcher{Client: httpClient, Token: opts.HFToken}
	modelReadmeFetcher := &fetcher.ModelReadmeFetcher{Client: httpClient, Token: opts.HFToken}
	datasetApiFetcher := &fetcher.DatasetAPIFetcher{Client: httpClient, Token: opts.HFToken}
	datasetReadmeFetcher := &fetcher.DatasetReadmeFetcher{Client: httpClient, Token: opts.HFToken}

	bomBuilder := newBOMBuilder()

	for i, d := range discoveries {
		modelID := strings.TrimSpace(d.ID)
		if modelID == "" {
			modelID = strings.TrimSpace(d.Name)
		}

		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		progress(ProgressEvent{Type: EventFetchStart, ModelID: modelID, Index: i, Total: len(discoveries)})

		var resp *fetcher.ModelAPIResponse
		var readme *fetcher.ModelReadmeCard

		if modelID != "" {
			r, err := modelApiFetcher.Fetch(ctx, modelID)
			if err != nil {
			} else {
				resp = r
				progress(ProgressEvent{Type: EventFetchAPIComplete, ModelID: modelID})
			}

			c, err := modelReadmeFetcher.Fetch(ctx, modelID)
			if err != nil {
			} else {
				readme = c
				progress(ProgressEvent{Type: EventFetchReadmeComplete, ModelID: modelID})
			}
		}

		progress(ProgressEvent{Type: EventBuildStart, ModelID: modelID})

		bctx := builder.BuildContext{
			ModelID: modelID,
			Scan:    d,
			HF:      resp,
			Readme:  readme,
		}

		bom, err := bomBuilder.Build(bctx)
		if err != nil {
			progress(ProgressEvent{Type: EventError, ModelID: modelID, Error: err})
			return nil, err
		}

		progress(ProgressEvent{Type: EventBuildComplete, ModelID: modelID})

		// Process datasets
		datasets := extractDatasetsFromModel(resp, readme)
		datasetCount := 0

		for _, dsID := range datasets {
			progress(ProgressEvent{Type: EventDatasetStart, ModelID: modelID, Message: dsID})

			dsResp, err := datasetApiFetcher.Fetch(ctx, dsID)
			if err != nil {
				// In online mode, skip dataset components that don't exist on HuggingFace.
				// The dataset reference is still preserved in the model's modelCard.modelParameters.datasets.
				// Dummy fallbacks are only used when explicitly running in --hf-mode dummy.
				continue
			}

			dsReadme, _ := datasetReadmeFetcher.Fetch(ctx, dsID)

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

			if bom.Components == nil {
				bom.Components = &[]cdx.Component{}
			}
			*bom.Components = append(*bom.Components, *dsComp)
			datasetCount++

			progress(ProgressEvent{Type: EventDatasetComplete, ModelID: modelID, Message: dsID})
		}

		// Add dependencies from model to datasets
		builder.AddDependencies(bom)

		progress(ProgressEvent{Type: EventModelComplete, ModelID: modelID, Datasets: datasetCount})

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

// BuildFromModelIDs is a convenience wrapper that generates BOMs from model IDs without progress reporting.
// For progress updates, use BuildFromModelIDsWithProgress.
func BuildFromModelIDs(modelIDs []string, hfToken string, timeout time.Duration) ([]DiscoveredBOM, error) {
	return BuildFromModelIDsWithProgress(context.Background(), modelIDs, GenerateOptions{
		HFToken:    hfToken,
		Timeout:    timeout,
		OnProgress: nil,
	})
}

// BuildFromModelIDsWithProgress generates BOMs from Hugging Face model IDs with progress reporting.
// This is used when --model-id is provided instead of --input.
func BuildFromModelIDsWithProgress(ctx context.Context, modelIDs []string, opts GenerateOptions) ([]DiscoveredBOM, error) {
	if opts.Timeout <= 0 {
		opts.Timeout = 10 * time.Second
	}

	progress := opts.OnProgress
	if progress == nil {
		progress = func(ProgressEvent) {} // no-op
	}

	results := make([]DiscoveredBOM, 0, len(modelIDs))

	httpClient := &http.Client{Timeout: opts.Timeout}
	modelApiFetcher := &fetcher.ModelAPIFetcher{Client: httpClient, Token: opts.HFToken}
	modelReadmeFetcher := &fetcher.ModelReadmeFetcher{Client: httpClient, Token: opts.HFToken}
	datasetApiFetcher := &fetcher.DatasetAPIFetcher{Client: httpClient, Token: opts.HFToken}
	datasetReadmeFetcher := &fetcher.DatasetReadmeFetcher{Client: httpClient, Token: opts.HFToken}

	for i, modelID := range modelIDs {
		modelID = strings.TrimSpace(modelID)
		if modelID == "" {
			continue
		}

		// Check for cancellation
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		bomBuilder := newBOMBuilder()

		progress(ProgressEvent{Type: EventFetchStart, ModelID: modelID, Index: i, Total: len(modelIDs)})

		// Fetch API metadata
		resp, err := modelApiFetcher.Fetch(ctx, modelID)
		if err != nil {
			progress(ProgressEvent{Type: EventError, ModelID: modelID, Error: err, Message: "API fetch failed"})
			resp = nil
		} else {
			progress(ProgressEvent{Type: EventFetchAPIComplete, ModelID: modelID})
		}

		// Fetch README
		readme, err := modelReadmeFetcher.Fetch(ctx, modelID)
		if err != nil {
			readme = nil
		} else {
			progress(ProgressEvent{Type: EventFetchReadmeComplete, ModelID: modelID})
		}

		// Build BOM
		progress(ProgressEvent{Type: EventBuildStart, ModelID: modelID})

		discovery := scanner.Discovery{
			ID:       modelID,
			Name:     modelID,
			Type:     "huggingface",
			Path:     "",
			Evidence: fmt.Sprintf("from model-id: %s", modelID),
		}

		bctx := builder.BuildContext{
			ModelID: modelID,
			Scan:    discovery,
			HF:      resp,
			Readme:  readme,
		}

		bom, err := bomBuilder.Build(bctx)
		if err != nil {
			progress(ProgressEvent{Type: EventError, ModelID: modelID, Error: err, Message: "BOM build failed"})
			continue
		}

		progress(ProgressEvent{Type: EventBuildComplete, ModelID: modelID})

		// Process datasets
		datasets := extractDatasetsFromModel(resp, readme)
		datasetCount := 0

		for _, dsID := range datasets {
			progress(ProgressEvent{Type: EventDatasetStart, ModelID: modelID, Message: dsID})

			dsResp, err := datasetApiFetcher.Fetch(ctx, dsID)
			if err != nil {
				continue
			}

			dsReadme, _ := datasetReadmeFetcher.Fetch(ctx, dsID)

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

			if bom.Components == nil {
				bom.Components = &[]cdx.Component{}
			}
			*bom.Components = append(*bom.Components, *dsComp)
			datasetCount++

			progress(ProgressEvent{Type: EventDatasetComplete, ModelID: modelID, Message: dsID})
		}

		// Add dependencies from model to datasets
		builder.AddDependencies(bom)

		progress(ProgressEvent{Type: EventModelComplete, ModelID: modelID, Datasets: datasetCount})

		results = append(results, DiscoveredBOM{
			Discovery: discovery,
			BOM:       bom,
		})
	}

	return results, nil
}
