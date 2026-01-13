package enricher

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/idlab-discover/AIBoMGen-cli/internal/completeness"
	"github.com/idlab-discover/AIBoMGen-cli/internal/fetcher"
	"github.com/idlab-discover/AIBoMGen-cli/internal/metadata"
)

// Config holds enrichment configuration
type Config struct {
	Strategy     string  // "interactive" or "file"
	ConfigFile   string  // path to config file (for file strategy)
	RequiredOnly bool    // only enrich required fields
	MinWeight    float64 // minimum weight threshold
	Refetch      bool    // refetch from Hugging Face
	NoPreview    bool    // skip preview
	SpecVersion  string  // CycloneDX spec version
	HFToken      string  // Hugging Face token
	HFBaseURL    string  // Hugging Face base URL
	HFTimeout    int     // timeout in seconds
}

// Options for creating an Enricher
type Options struct {
	Reader io.Reader
	Writer io.Writer
	Config Config
}

// Enricher handles AIBOM enrichment
type Enricher struct {
	reader io.Reader
	writer io.Writer
	config Config
	scan   *bufio.Scanner
}

// New creates a new Enricher
func New(opts Options) *Enricher {
	return &Enricher{
		reader: opts.Reader,
		writer: opts.Writer,
		config: opts.Config,
		scan:   bufio.NewScanner(opts.Reader),
	}
}

// Enrich enriches a BOM with additional metadata
func (e *Enricher) Enrich(bom *cdx.BOM, fileValues map[string]interface{}) (*cdx.BOM, error) {
	if bom == nil {
		return nil, fmt.Errorf("nil BOM")
	}

	// Get model ID from BOM
	modelID := extractModelID(bom)
	if modelID == "" {
		logf("", "warning: could not extract model ID from BOM")
	}

	// Refetch metadata if requested
	var hfAPI *fetcher.ModelAPIResponse
	var hfReadme *fetcher.ModelReadmeCard
	if e.config.Refetch && modelID != "" {
		logf(modelID, "refetching metadata from Hugging Face...")
		hfAPI, hfReadme = e.refetchMetadata(modelID)
	}

	// Run initial completeness check
	initialReport := completeness.Check(bom)
	logf(modelID, "initial completeness: %.2f (%d/%d fields)", initialReport.Score, initialReport.Passed, initialReport.Total)

	// Collect missing fields based on config
	missingFields := e.collectMissingFields(initialReport)
	if len(missingFields) == 0 {
		logf(modelID, "no fields to enrich")
		return bom, nil
	}

	logf(modelID, "found %d field(s) to enrich", len(missingFields))

	// Prepare enrichment source
	src := metadata.Source{
		ModelID: modelID,
		HF:      hfAPI,
		Readme:  hfReadme,
	}

	// Prepare enrichment target - modify the BOM directly
	tgt := metadata.Target{
		BOM:                bom,
		Component:          bomComponent(bom),
		ModelCard:          bomModelCard(bom),
		HuggingFaceBaseURL: e.config.HFBaseURL,
	}

	// Track changes
	changes := make(map[metadata.Key]string)

	// Enrich each field
	for _, spec := range missingFields {
		var value interface{}
		var err error

		switch e.config.Strategy {
		case "file":
			value, err = e.getValueFromFile(spec, fileValues)
		case "interactive":
			value, err = e.getValueInteractive(spec, src, tgt.BOM)
		default:
			return nil, fmt.Errorf("unknown strategy: %s", e.config.Strategy)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to get value for %s: %w", spec.Key, err)
		}

		if value != nil {
			// Apply the value
			if err := e.applyValue(spec, &src, &tgt, value); err != nil {
				logf(modelID, "failed to apply %s: %v", spec.Key, err)
				continue
			}
			// Only track the change if it was successfully applied
			if spec.SetUserValue != nil {
				changes[spec.Key] = formatValue(value)
			}
		}
	}

	// Show preview if requested
	if !e.config.NoPreview && len(changes) > 0 {
		if !e.showPreviewAndConfirm(initialReport, bom, changes) {
			return nil, fmt.Errorf("enrichment cancelled by user")
		}
	}

	return bom, nil
}

// collectMissingFields returns fields that need enrichment based on config
func (e *Enricher) collectMissingFields(report completeness.Report) []metadata.FieldSpec {
	var result []metadata.FieldSpec

	for _, spec := range metadata.Registry() {
		// Skip if weight is 0 or below threshold
		if spec.Weight <= 0 || spec.Weight < e.config.MinWeight {
			continue
		}

		// Check if field is missing
		isMissing := false
		for _, k := range report.MissingRequired {
			if k == spec.Key {
				isMissing = true
				break
			}
		}
		if !isMissing && !e.config.RequiredOnly {
			for _, k := range report.MissingOptional {
				if k == spec.Key {
					isMissing = true
					break
				}
			}
		}

		if isMissing {
			result = append(result, spec)
		}
	}

	return result
}

// refetchMetadata fetches fresh metadata from Hugging Face
func (e *Enricher) refetchMetadata(modelID string) (*fetcher.ModelAPIResponse, *fetcher.ModelReadmeCard) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(e.config.HFTimeout)*time.Second)
	defer cancel()

	// Fetch API
	apiFetcher := &fetcher.ModelAPIFetcher{
		Token:   e.config.HFToken,
		BaseURL: e.config.HFBaseURL,
	}
	apiResp, err := apiFetcher.Fetch(ctx, modelID)
	if err != nil {
		logf(modelID, "API fetch failed: %v", err)
	}

	// Fetch README
	readmeFetcher := &fetcher.ModelReadmeFetcher{
		Token:   e.config.HFToken,
		BaseURL: e.config.HFBaseURL,
	}
	readme, err := readmeFetcher.Fetch(ctx, modelID)
	if err != nil {
		logf(modelID, "README fetch failed: %v", err)
	}

	return apiResp, readme
}

// getValueFromFile extracts a value from the config file
func (e *Enricher) getValueFromFile(spec metadata.FieldSpec, values map[string]interface{}) (interface{}, error) {
	// Map field spec key to config file key
	// e.g. "BOM.metadata.component.licenses" -> "licenses"
	key := configKeyFromSpec(spec.Key)

	if val, ok := values[key]; ok {
		logf("", "loaded %s from config file", spec.Key)
		return val, nil
	}

	logf("", "no value found for %s in config file", spec.Key)
	return nil, nil
}

// getValueInteractive prompts the user for a value
func (e *Enricher) getValueInteractive(spec metadata.FieldSpec, src metadata.Source, bom *cdx.BOM) (interface{}, error) {
	// Print field info
	required := ""
	if spec.Required {
		required = " [REQUIRED]"
	}
	fmt.Fprintf(e.writer, "\n%s (weight: %.1f)%s\n", spec.Key, spec.Weight, required)

	// Show suggestions if available
	suggestions := e.getSuggestions(spec, src, bom)
	if len(suggestions) > 0 {
		fmt.Fprintf(e.writer, "  Suggestions: %s\n", strings.Join(suggestions, ", "))
	}

	// Prompt for input
	fmt.Fprintf(e.writer, "  Enter value (or press Enter to skip): ")

	if !e.scan.Scan() {
		return nil, e.scan.Err()
	}

	input := strings.TrimSpace(e.scan.Text())
	if input == "" {
		return nil, nil
	}

	return input, nil
}

// getSuggestions returns suggested values for a field
func (e *Enricher) getSuggestions(spec metadata.FieldSpec, src metadata.Source, bom *cdx.BOM) []string {
	// Field-specific suggestions and format hints
	switch spec.Key {
	case "BOM.metadata.component.licenses":
		return []string{"MIT", "Apache-2.0", "BSD-3-Clause", "GPL-3.0", "CC-BY-4.0"}
	case "BOM.metadata.component.modelCard.modelParameters.task":
		return []string{"text-classification", "text-generation", "image-classification", "object-detection"}
	case "BOM.metadata.component.modelCard.considerations.ethicalConsiderations":
		return []string{"Format: 'name: mitigation' or 'name1: mitigation1, name2: mitigation2'"}
	case "BOM.metadata.component.modelCard.considerations.environmentalConsiderations.properties":
		return []string{"Format: 'key:value, key:value' (e.g., 'hardwareType:GPU, carbonEmitted:100kg')"}
	case "BOM.metadata.component.modelCard.quantitativeAnalysis.performanceMetrics":
		return []string{"Format: 'type:value, type:value' (e.g., 'accuracy:0.95, f1:0.92')"}
	case "BOM.metadata.component.tags":
		return []string{"Comma-separated (e.g., 'pytorch, nlp, transformer')"}
	case "BOM.metadata.component.modelCard.modelParameters.datasets":
		return []string{"Comma-separated dataset refs (e.g., 'imagenet, coco')"}
	case "BOM.metadata.component.modelCard.considerations.useCases":
		return []string{"Comma-separated (e.g., 'text classification, sentiment analysis')"}
	case "BOM.metadata.component.modelCard.considerations.technicalLimitations":
		return []string{"Comma-separated (e.g., 'requires GPU, limited to English')"}
	default:
		return nil
	}
}

// applyValue applies a user-provided value to the BOM using the FieldSpec's SetUserValue function
func (e *Enricher) applyValue(spec metadata.FieldSpec, src *metadata.Source, tgt *metadata.Target, value interface{}) error {
	strValue := fmt.Sprintf("%v", value)

	// Use the FieldSpec's SetUserValue if available
	if spec.SetUserValue != nil {
		err := spec.SetUserValue(strValue, *tgt)
		if err != nil {
			return fmt.Errorf("failed to set user value for %s: %w", spec.Key, err)
		}
		logf(src.ModelID, "applied user value for %s", spec.Key)
		return nil
	}

	// Fallback: if no SetUserValue, log a warning
	logf(src.ModelID, "warning: no SetUserValue function for %s, value not applied", spec.Key)
	return nil
}

// showPreviewAndConfirm shows changes and asks for confirmation
func (e *Enricher) showPreviewAndConfirm(initial completeness.Report, enriched *cdx.BOM, changes map[metadata.Key]string) bool {
	fmt.Fprintf(e.writer, "\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Fprintf(e.writer, "Preview Changes\n")
	fmt.Fprintf(e.writer, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	for key, value := range changes {
		fmt.Fprintf(e.writer, "  + %s: %s\n", key, value)
	}

	// Show new completeness score
	newReport := completeness.Check(enriched)
	fmt.Fprintf(e.writer, "\nCompleteness: %.2f → %.2f\n", initial.Score, newReport.Score)
	fmt.Fprintf(e.writer, "Fields: %d/%d → %d/%d\n", initial.Passed, initial.Total, newReport.Passed, newReport.Total)

	fmt.Fprintf(e.writer, "\nSave changes? [Y/n]: ")

	if !e.scan.Scan() {
		return false
	}

	response := strings.ToLower(strings.TrimSpace(e.scan.Text()))
	return response == "" || response == "y" || response == "yes"
}

// Helper functions

func extractModelID(bom *cdx.BOM) string {
	if c := bomComponent(bom); c != nil {
		return c.Name
	}
	return ""
}

func bomComponent(bom *cdx.BOM) *cdx.Component {
	if bom == nil || bom.Metadata == nil || bom.Metadata.Component == nil {
		return nil
	}
	return bom.Metadata.Component
}

func bomModelCard(bom *cdx.BOM) *cdx.MLModelCard {
	c := bomComponent(bom)
	if c == nil || c.ModelCard == nil {
		return nil
	}
	return c.ModelCard
}

func configKeyFromSpec(key metadata.Key) string {
	// Convert "BOM.metadata.component.licenses" to "licenses"
	parts := strings.Split(string(key), ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return string(key)
}

func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int, int64, float64:
		return fmt.Sprintf("%v", val)
	case []string:
		return strings.Join(val, ", ")
	default:
		return fmt.Sprintf("%v", val)
	}
}
