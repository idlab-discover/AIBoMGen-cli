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

// BuildPerDiscovery orchestrates: fetch HF API (optional) → build BOM per model via registry-driven builder.
func BuildPerDiscovery(discoveries []scanner.Discovery, hfToken string, timeout time.Duration) ([]DiscoveredBOM, error) {
	results := make([]DiscoveredBOM, 0, len(discoveries))

	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	httpClient := &http.Client{Timeout: timeout}
	apiFetcher := &fetcher.ModelAPIFetcher{Client: httpClient, Token: hfToken}

	bomBuilder := newBOMBuilder()

	for _, d := range discoveries {
		modelID := strings.TrimSpace(d.ID)
		if modelID == "" {
			modelID = strings.TrimSpace(d.Name)
		}

		logf(modelID, "start (scanPath=%s)", strings.TrimSpace(d.Path))

		var resp *fetcher.ModelAPIResponse
		if modelID != "" {
			logf(modelID, "fetch HF model metadata")
			r, err := apiFetcher.Fetch(context.Background(), modelID)
			if err != nil {
				logf(modelID, "fetch failed (%v)", err)
			} else {
				resp = r
				logf(modelID, "metadata fetched")
			}
		}

		ctx := builder.BuildContext{
			ModelID: modelID,
			Scan:    d,
			HF:      resp,
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
