package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/idlab-discover/AIBoMGen-cli/internal/builder"
	"github.com/idlab-discover/AIBoMGen-cli/internal/fetcher"
	"github.com/idlab-discover/AIBoMGen-cli/internal/generator"
	bomio "github.com/idlab-discover/AIBoMGen-cli/internal/io"
	"github.com/idlab-discover/AIBoMGen-cli/internal/metadata"
	"github.com/idlab-discover/AIBoMGen-cli/internal/scanner"
)

var (
	generatePath         string
	generateOutput       string
	generateOutputFormat string
	generateSpecVersion  string
	generateModelIDs     []string

	// hfMode controls whether metadata is fetched from Hugging Face.
	// Supported values: online|dummy
	hfMode       string
	hfTimeoutSec int
	hfToken      string

	enrich bool
	// Logging is controlled via generateLogLevel.
	generateLogLevel string
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate an AI-aware BOM (AIBOM)",
	Long:  "Generate BOM from Hugging Face imports: either scan a directory or provide model ID(s) directly.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Resolve effective log level (from config, env, or flag).
		level := strings.ToLower(strings.TrimSpace(viper.GetString("generate.log-level")))
		if level == "" {
			level = "standard"
		}
		switch level {
		case "quiet", "standard", "debug":
			// ok
		default:
			return fmt.Errorf("invalid --log-level %q (expected quiet|standard|debug)", level)
		}

		// Resolve effective HF mode (from config, env, or flag).
		mode := strings.ToLower(strings.TrimSpace(viper.GetString("generate.hf-mode")))
		if mode == "" {
			mode = "online"
		}
		switch mode {
		case "online", "dummy":
			// ok
		default:
			return fmt.Errorf("invalid --hf-mode %q (expected online|dummy)", mode)
		}

		// Check if --model-id was explicitly provided on the command line
		modelIDFlagProvided := cmd.Flags().Changed("model-id")

		// Get model IDs from viper (respects config file and CLI flag)
		modelIDs := viper.GetStringSlice("generate.model-ids")
		// Filter out empty strings
		var cleanModelIDs []string
		for _, id := range modelIDs {
			if trimmed := strings.TrimSpace(id); trimmed != "" {
				cleanModelIDs = append(cleanModelIDs, trimmed)
			}
		}

		inputPath := viper.GetString("generate.input")
		inputPathProvided := cmd.Flags().Changed("input")

		// Determine which mode we're in
		// Priority: explicit --model-id flag > explicit --input flag > defaults
		var useModelIDMode bool
		if modelIDFlagProvided && len(cleanModelIDs) > 0 {
			useModelIDMode = true
			// If both are explicitly provided, that's an error
			if inputPathProvided && inputPath != "" {
				return fmt.Errorf("cannot specify both --model-id and --input (folder scan)")
			}
		} else if inputPathProvided && inputPath != "" {
			// Explicit --input provided
			useModelIDMode = false
		} else if len(cleanModelIDs) > 0 {
			// Model IDs from config, no explicit --input
			useModelIDMode = true
			inputPath = ""
		} else {
			// Use default input path
			useModelIDMode = false
			if inputPath == "" {
				inputPath = "."
			}
		}

		// Get format from viper (respects config file)
		outputFormat := viper.GetString("generate.format")
		if outputFormat == "" {
			outputFormat = "auto"
		}

		// Get spec version from viper
		specVersion := viper.GetString("generate.spec")

		// Get output path from viper for early validation
		outputPath := viper.GetString("generate.output")

		// Fail fast on explicit format/extension mismatch before scanning
		if outputPath != "" && outputFormat != "" && outputFormat != "auto" {
			ext := filepath.Ext(outputPath)
			if outputFormat == "xml" && ext == ".json" {
				return fmt.Errorf("output path extension %q does not match format %q", ext, outputFormat)
			}
			if outputFormat == "json" && ext == ".xml" {
				return fmt.Errorf("output path extension %q does not match format %q", ext, outputFormat)
			}
		}

		// Wire internal package logging based on log level.
		if level != "quiet" {
			lw := cmd.ErrOrStderr()
			if useModelIDMode {
				generator.SetLogger(lw)
			} else {
				scanner.SetLogger(lw)
				generator.SetLogger(lw)
			}
			if level == "debug" {
				fetcher.SetLogger(lw)
				metadata.SetLogger(lw)
				builder.SetLogger(lw)
			}
		}

		// Get HF settings from viper
		hfToken := viper.GetString("generate.hf-token")
		hfTimeout := viper.GetInt("generate.hf-timeout")
		if hfTimeout <= 0 {
			hfTimeout = 10
		}
		timeout := time.Duration(hfTimeout) * time.Second

		var discoveredBOMs []generator.DiscoveredBOM
		var err error

		if useModelIDMode {
			// Model ID mode: generate from one or more model IDs
			if mode == "dummy" {
				discoveredBOMs, err = generator.BuildDummyBOM()
				if err != nil {
					return err
				}
			} else {
				// Online mode: fetch and build BOM for the given model IDs
				discoveredBOMs, err = generator.BuildFromModelIDs(cleanModelIDs, hfToken, timeout)
				if err != nil {
					return err
				}
			}
		} else {
			// Folder scan mode: scan directory for Hugging Face imports
			absTarget, err := filepath.Abs(inputPath)
			if err != nil {
				return err
			}

			if mode == "dummy" {
				discoveredBOMs, err = generator.BuildDummyBOM()
				if err != nil {
					return err
				}
			} else {
				// Scan for AI components
				discoveries, err := scanner.Scan(absTarget)
				if err != nil {
					return err
				}

				// Online mode: per discovery: fetch + map + build
				discoveredBOMs, err = generator.BuildPerDiscovery(discoveries, hfToken, timeout)
				if err != nil {
					return err
				}
			}
		}

		// Optional enrichment: only run when requested.
		// The enricher is now a separate command; this flag is deprecated.
		if viper.GetBool("generate.enrich") {
			fmt.Fprintf(cmd.ErrOrStderr(), "[warn] --enrich flag is deprecated. Use 'aibomgen-cli enrich' command instead.\n")
		}

		// Get output path from viper
		output := viper.GetString("generate.output")
		if output == "" {
			// Default extension based on requested format (json unless explicitly xml)
			if outputFormat == "xml" {
				output = "dist/aibom.xml"
			} else {
				output = "dist/aibom.json"
			}
		}

		fmtChosen := outputFormat
		if fmtChosen == "auto" || fmtChosen == "" {
			ext := filepath.Ext(output)
			if ext == ".xml" {
				fmtChosen = "xml"
			} else {
				fmtChosen = "json"
			}
		}

		outputDir := filepath.Dir(output)
		if outputDir == "" {
			outputDir = "."
		}
		outputDir = filepath.Clean(outputDir)
		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return err
		}

		fileExt := ".json"
		if fmtChosen == "xml" {
			fileExt = ".xml"
		}

		written := make([]string, 0, len(discoveredBOMs))
		for _, d := range discoveredBOMs {
			name := bomMetadataComponentName(d.BOM)
			if strings.TrimSpace(name) == "" {
				// filename fallback
				name = strings.TrimSpace(d.Discovery.Name)
				if name == "" {
					name = strings.TrimSpace(d.Discovery.ID)
				}
				if name == "" {
					name = "model"
				}
			}

			sanitized := sanitizeComponentName(name)
			fileName := fmt.Sprintf("%s_aibom%s", sanitized, fileExt)
			dest := filepath.Join(outputDir, fileName)

			if err := bomio.WriteBOM(d.BOM, dest, fmtChosen, specVersion); err != nil {
				return err
			}
			written = append(written, dest)
		}

		// Add a blank line before the final success message for readability
		fmt.Fprintln(cmd.OutOrStdout())
		if len(written) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No models detected; no AIBOM files written.")
			return nil
		}
		fmt.Fprintf(cmd.OutOrStdout(), "AIBOM files written (%d) under %s (format: %s)\n", len(written), outputDir, fmtChosen)
		return nil
	},
}

func init() {
	generateCmd.Flags().StringVarP(&generatePath, "input", "i", "", "Path to scan (folder scan mode)")
	generateCmd.Flags().StringSliceVarP(&generateModelIDs, "model-id", "m", []string{}, "Hugging Face model ID(s) (e.g., gpt2 or org/model-name) - can be used multiple times or comma-separated")
	generateCmd.Flags().StringVarP(&generateOutput, "output", "o", "", "Output file path (directory is used)")
	generateCmd.Flags().StringVarP(&generateOutputFormat, "format", "f", "", "Output BOM format: json|xml|auto")
	generateCmd.Flags().StringVar(&generateSpecVersion, "spec", "", "CycloneDX spec version for output (e.g., 1.4, 1.5, 1.6)")
	generateCmd.Flags().StringVar(&hfMode, "hf-mode", "", "Hugging Face metadata mode: online|dummy")
	generateCmd.Flags().IntVar(&hfTimeoutSec, "hf-timeout", 0, "HTTP timeout in seconds for Hugging Face API")
	generateCmd.Flags().StringVar(&hfToken, "hf-token", "", "Hugging Face access token")
	generateCmd.Flags().BoolVar(&enrich, "enrich", false, "Prompt for missing fields and compute completeness (deprecated)")

	generateCmd.Flags().StringVar(&generateLogLevel, "log-level", "", "Log level: quiet|standard|debug")

	// Bind all flags to viper for config file support
	viper.BindPFlag("generate.input", generateCmd.Flags().Lookup("input"))
	viper.BindPFlag("generate.model-ids", generateCmd.Flags().Lookup("model-id"))
	viper.BindPFlag("generate.output", generateCmd.Flags().Lookup("output"))
	viper.BindPFlag("generate.format", generateCmd.Flags().Lookup("format"))
	viper.BindPFlag("generate.spec", generateCmd.Flags().Lookup("spec"))
	viper.BindPFlag("generate.hf-mode", generateCmd.Flags().Lookup("hf-mode"))
	viper.BindPFlag("generate.hf-timeout", generateCmd.Flags().Lookup("hf-timeout"))
	viper.BindPFlag("generate.hf-token", generateCmd.Flags().Lookup("hf-token"))
	viper.BindPFlag("generate.enrich", generateCmd.Flags().Lookup("enrich"))
	viper.BindPFlag("generate.log-level", generateCmd.Flags().Lookup("log-level"))
}

func bomMetadataComponentName(bom *cdx.BOM) string {
	if bom == nil || bom.Metadata == nil || bom.Metadata.Component == nil {
		return ""
	}
	return bom.Metadata.Component.Name
}

func sanitizeComponentName(name string) string {
	if name == "" {
		return "model"
	}
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_' || r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	sanitized := b.String()
	if sanitized == "" {
		return "model"
	}
	return sanitized
}
