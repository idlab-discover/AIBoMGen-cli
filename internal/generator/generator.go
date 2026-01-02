package generator

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/idlab-discover/AIBoMGen-cli/internal/builder"
	"github.com/idlab-discover/AIBoMGen-cli/internal/fetcher"
	bomio "github.com/idlab-discover/AIBoMGen-cli/internal/io"
	"github.com/idlab-discover/AIBoMGen-cli/internal/scanner"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

type DiscoveredBOM struct {
	Discovery scanner.Discovery
	BOM       *cdx.BOM
}

// BuildPerDiscovery orchestrates: fetch HF API (optional) → build BOM per model via registry-driven builder.
func BuildPerDiscovery(discoveries []scanner.Discovery, hfToken string, timeout time.Duration) ([]DiscoveredBOM, error) {
	results := make([]DiscoveredBOM, 0, len(discoveries))

	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	httpClient := &http.Client{Timeout: timeout}
	apiFetcher := &fetcher.ModelAPIFetcher{Client: httpClient, Token: hfToken}

	bomBuilder := builder.NewBOMBuilder(builder.DefaultOptions())

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

// Write writes the BOM to the given output path, creating directories as needed.
func Write(outputPath string, bom *cdx.BOM) error { return WriteWithFormat(outputPath, bom, "json") }

// WriteWithFormat writes the BOM in the specified format (json|xml). If format is auto, infer from extension.
func WriteWithFormat(outputPath string, bom *cdx.BOM, format string) error {
	return WriteWithFormatAndSpec(outputPath, bom, format, "")
}

// WriteWithFormatAndSpec writes the BOM with the specified file format and optional spec version.
// If spec is non-empty (e.g. "1.3"), EncodeVersion is used.
func WriteWithFormatAndSpec(outputPath string, bom *cdx.BOM, format string, spec string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	return bomio.WriteBOM(bom, outputPath, format, spec)
}

// ParseSpecVersion parses a spec version string.
// Deprecated: Use bomio.ParseSpecVersion instead.
func ParseSpecVersion(s string) (cdx.SpecVersion, bool) {
	return bomio.ParseSpecVersion(s)
}
