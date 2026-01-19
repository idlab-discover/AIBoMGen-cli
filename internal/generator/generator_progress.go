package generator

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/idlab-discover/AIBoMGen-cli/internal/builder"
	"github.com/idlab-discover/AIBoMGen-cli/internal/fetcher"
	"github.com/idlab-discover/AIBoMGen-cli/internal/scanner"
)

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

// BuildFromModelIDsWithProgress generates BOMs with progress reporting
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
			logf(modelID, "API fetch failed: %v", err)
			progress(ProgressEvent{Type: EventError, ModelID: modelID, Error: err, Message: "API fetch failed"})
			resp = nil
		} else {
			progress(ProgressEvent{Type: EventFetchAPIComplete, ModelID: modelID})
		}

		// Fetch README
		readme, err := modelReadmeFetcher.Fetch(ctx, modelID)
		if err != nil {
			logf(modelID, "README fetch failed: %v", err)
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
				logf(dsID, "dataset fetch failed, skipping: %v", err)
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
				logf(modelID, "failed to build dataset %s: %v", dsID, err)
				continue
			}

			if bom.Components == nil {
				bom.Components = &[]cdx.Component{}
			}
			*bom.Components = append(*bom.Components, *dsComp)
			datasetCount++

			progress(ProgressEvent{Type: EventDatasetComplete, ModelID: modelID, Message: dsID})
		}

		progress(ProgressEvent{Type: EventModelComplete, ModelID: modelID, Datasets: datasetCount})

		results = append(results, DiscoveredBOM{
			Discovery: discovery,
			BOM:       bom,
		})
	}

	return results, nil
}

// BuildPerDiscoveryWithProgress orchestrates BOM generation with progress reporting
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
				logf(modelID, "fetch failed: %v", err)
			} else {
				resp = r
				progress(ProgressEvent{Type: EventFetchAPIComplete, ModelID: modelID})
			}

			c, err := modelReadmeFetcher.Fetch(ctx, modelID)
			if err != nil {
				logf(modelID, "readme fetch failed: %v", err)
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
				logf(dsID, "dataset fetch failed, skipping: %v", err)
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
				logf(modelID, "failed to build dataset %s: %v", dsID, err)
				continue
			}

			if bom.Components == nil {
				bom.Components = &[]cdx.Component{}
			}
			*bom.Components = append(*bom.Components, *dsComp)
			datasetCount++

			progress(ProgressEvent{Type: EventDatasetComplete, ModelID: modelID, Message: dsID})
		}

		progress(ProgressEvent{Type: EventModelComplete, ModelID: modelID, Datasets: datasetCount})

		results = append(results, DiscoveredBOM{
			Discovery: d,
			BOM:       bom,
		})
	}

	return results, nil
}
