package cmd

import (
	"fmt"
	"strings"

	"github.com/idlab-discover/AIBoMGen-cli/internal/completeness"
	"github.com/idlab-discover/AIBoMGen-cli/internal/enricher"
	"github.com/idlab-discover/AIBoMGen-cli/internal/fetcher"
	bomio "github.com/idlab-discover/AIBoMGen-cli/internal/io"
	"github.com/idlab-discover/AIBoMGen-cli/internal/metadata"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// enrichCmd represents the enrich command
var enrichCmd = &cobra.Command{
	Use:   "enrich",
	Short: "Enrich an existing AIBOM with additional metadata",
	Long: `Enrich an existing AIBOM with additional metadata through interactive prompts
or by loading values from a configuration file. Optionally refetch model metadata
from Hugging Face API and README before enrichment.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate strategy
		strategy := strings.ToLower(strings.TrimSpace(enrichStrategy))
		if strategy == "" {
			strategy = "interactive"
		}
		switch strategy {
		case "interactive", "file":
			// ok
		default:
			return fmt.Errorf("invalid --strategy %q (expected interactive|file)", enrichStrategy)
		}

		// Validate log level
		level := strings.ToLower(strings.TrimSpace(enrichLogLevel))
		if level == "" {
			level = "standard"
		}
		switch level {
		case "quiet", "standard", "debug":
			// ok
		default:
			return fmt.Errorf("invalid --log-level %q (expected quiet|standard|debug)", enrichLogLevel)
		}

		// Wire internal package logging
		if level != "quiet" {
			lw := cmd.ErrOrStderr()
			enricher.SetLogger(lw)
			completeness.SetLogger(lw)
			fetcher.SetLogger(lw)
			if level == "debug" {
				metadata.SetLogger(lw)
			}
		}

		// Read existing BOM
		bom, err := bomio.ReadBOM(enrichInput, enrichInputFormat)
		if err != nil {
			return fmt.Errorf("failed to read input BOM: %w", err)
		}

		// Determine output path
		outPath := enrichOutput
		if outPath == "" {
			outPath = enrichInput // overwrite by default
		}

		// Determine spec version (default to same version as input)
		// Note: if not specified, WriteBOM will use the BOM's existing spec version
		specVersion := strings.TrimSpace(enrichSpecVersion)

		// Build enricher configuration
		cfg := enricher.Config{
			Strategy:     strategy,
			ConfigFile:   enrichConfigFile,
			RequiredOnly: enrichRequiredOnly,
			MinWeight:    enrichMinWeight,
			Refetch:      enrichRefetch,
			NoPreview:    enrichNoPreview,
			SpecVersion:  specVersion,
			HFToken:      enrichHFToken,
			HFBaseURL:    enrichHFBaseURL,
			HFTimeout:    enrichHFTimeout,
		}

		// Load config file values if using file strategy
		var fileValues map[string]interface{}
		if strategy == "file" {
			if enrichConfigFile == "" {
				return fmt.Errorf("--file is required when using --strategy file")
			}
			fileValues, err = loadEnrichmentConfig(enrichConfigFile)
			if err != nil {
				return fmt.Errorf("failed to load config file: %w", err)
			}
		}

		// Create enricher
		e := enricher.New(enricher.Options{
			Reader: cmd.InOrStdin(),
			Writer: cmd.OutOrStdout(),
			Config: cfg,
		})

		// Run enrichment
		enriched, err := e.Enrich(bom, fileValues)
		if err != nil {
			return fmt.Errorf("enrichment failed: %w", err)
		}

		// Write output
		if err := bomio.WriteBOM(enriched, outPath, enrichOutputFormat, specVersion); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}

		if level != "quiet" {
			fmt.Fprintf(cmd.OutOrStdout(), "\n✓ Enriched BOM saved to %s\n", outPath)
		}

		return nil
	},
}

var (
	enrichInput        string
	enrichInputFormat  string
	enrichOutput       string
	enrichOutputFormat string
	enrichSpecVersion  string
	enrichStrategy     string
	enrichConfigFile   string
	enrichRequiredOnly bool
	enrichMinWeight    float64
	enrichRefetch      bool
	enrichNoPreview    bool
	enrichLogLevel     string
	enrichHFToken      string
	enrichHFBaseURL    string
	enrichHFTimeout    int
)

func init() {
	enrichCmd.Flags().StringVarP(&enrichInput, "input", "i", "", "Path to existing AIBOM (required)")
	enrichCmd.Flags().StringVarP(&enrichOutput, "output", "o", "", "Output file path (default: overwrite input)")
	enrichCmd.Flags().StringVarP(&enrichInputFormat, "format", "f", "auto", "Input BOM format: json|xml|auto")
	enrichCmd.Flags().StringVar(&enrichOutputFormat, "output-format", "auto", "Output BOM format: json|xml|auto")
	enrichCmd.Flags().StringVar(&enrichSpecVersion, "spec", "", "CycloneDX spec version for output (default: same as input)")

	enrichCmd.Flags().StringVar(&enrichStrategy, "strategy", "interactive", "Enrichment strategy: interactive|file")
	enrichCmd.Flags().StringVar(&enrichConfigFile, "file", "/config/enrichment.yaml", "Path to enrichment config file (YAML)")
	enrichCmd.Flags().BoolVar(&enrichRequiredOnly, "required-only", false, "Only prompt for required fields")
	enrichCmd.Flags().Float64Var(&enrichMinWeight, "min-weight", 0.0, "Only prompt for fields with weight >= this value")
	enrichCmd.Flags().BoolVar(&enrichRefetch, "refetch", false, "Refetch model metadata from Hugging Face before enrichment")
	enrichCmd.Flags().BoolVar(&enrichNoPreview, "no-preview", false, "Skip preview before saving")

	enrichCmd.Flags().StringVar(&enrichLogLevel, "log-level", "standard", "Log level: quiet|standard|debug")
	enrichCmd.Flags().StringVar(&enrichHFToken, "hf-token", "", "Hugging Face API token (for refetch)")
	enrichCmd.Flags().StringVar(&enrichHFBaseURL, "hf-base-url", "", "Hugging Face base URL (for refetch)")
	enrichCmd.Flags().IntVar(&enrichHFTimeout, "hf-timeout", 30, "Hugging Face API timeout in seconds (for refetch)")

	_ = enrichCmd.MarkFlagRequired("input")
}

// loadEnrichmentConfig loads enrichment values from a YAML config file
func loadEnrichmentConfig(path string) (map[string]interface{}, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	return v.AllSettings(), nil
}
