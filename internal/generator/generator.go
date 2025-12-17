package generator

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aibomgen-cra/internal/builder"
	"aibomgen-cra/internal/fetcher"
	"aibomgen-cra/internal/metadata"
	"aibomgen-cra/internal/scanner"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

type DiscoveredBOM struct {
	Discovery scanner.Discovery
	BOM       *cdx.BOM
}

var logOut io.Writer

// BuildPerDiscovery orchestrates: fetch HF API → map into metadata → build BOM per model.
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

		// per-discovery store
		store := metadata.NewStore()

		// per-discovery fetch
		if modelID != "" {
			logf(modelID, "fetch HF model metadata")
			resp, err := apiFetcher.Fetch(context.Background(), modelID)
			if err != nil {
				logf(modelID, "fetch failed (%v)", err)
				// continue building with empty store (builders should have fallbacks)
			} else {
				logf(modelID, "map HF response into metadata")
				metadata.MapModelAPIToMetadata(modelID, resp, store)
				logf(modelID, "metadata mapped")
			}
		}

		ctx := builder.BuildContext{
			ModelID: modelID,
			Scan:    d,
			Meta:    store.View(modelID),
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

	actual := format
	if actual == "auto" || actual == "" {
		ext := strings.ToLower(filepath.Ext(outputPath))
		if ext == ".xml" {
			actual = "xml"
		} else {
			actual = "json"
		}
	}

	// Enforce extension/format consistency when extension present and format explicitly set
	ext := strings.ToLower(filepath.Ext(outputPath))
	if actual != "auto" && ext != "" {
		switch actual {
		case "xml":
			if ext != ".xml" {
				return fmt.Errorf("output path extension %q does not match format %q", ext, actual)
			}
		case "json":
			if ext != ".json" {
				return fmt.Errorf("output path extension %q does not match format %q", ext, actual)
			}
		}
	}

	fileFmt := cdx.BOMFileFormatJSON
	if actual == "xml" {
		fileFmt = cdx.BOMFileFormatXML
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := cdx.NewBOMEncoder(f, fileFmt)
	encoder.SetPretty(true)

	if spec == "" {
		return encoder.Encode(bom)
	}

	sv, ok := ParseSpecVersion(spec)
	if !ok {
		return fmt.Errorf("unsupported CycloneDX spec version: %q", spec)
	}
	return encoder.EncodeVersion(bom, sv)
}

func ParseSpecVersion(s string) (cdx.SpecVersion, bool) {
	switch s {
	case "1.0":
		return cdx.SpecVersion1_0, true
	case "1.1":
		return cdx.SpecVersion1_1, true
	case "1.2":
		return cdx.SpecVersion1_2, true
	case "1.3":
		return cdx.SpecVersion1_3, true
	case "1.4":
		return cdx.SpecVersion1_4, true
	case "1.5":
		return cdx.SpecVersion1_5, true
	case "1.6":
		return cdx.SpecVersion1_6, true
	default:
		return 0, false
	}
}
