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

	// hfMode controls whether metadata is fetched from Hugging Face.
	// Supported values: online|offline|dummy
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
	Long:  "Scans the target path for AI Hugginface imports and produces a CycloneDX AIBOM JSON.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get input path from viper (respects config file and CLI flag)
		target := viper.GetString("generate.input")
		if target == "" {
			target = "."
		}
		absTarget, err := filepath.Abs(target)
		if err != nil {
			return err
		}

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
			scanner.SetLogger(lw)
			generator.SetLogger(lw)
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
		if mode == "dummy" {
			// Dummy mode: do not look at the input (Scanner discoveries)
			// Just build one fixed dummy BOM with all fields included.
			// Uses dummy_model_api_fetcher.go and dummy_model_readme_fetcher.go
			discoveredBOMs, err = generator.BuildDummyBOM()
			if err != nil {
				return err
			}
		} else {
			// Scan for AI components (only in non-dummy mode)
			discoveries, err := scanner.Scan(absTarget)
			if err != nil {
				return err
			}

			// Online mode: per discovery: store + fetch + map + build (inside generator).
			discoveredBOMs, err = generator.BuildPerDiscovery(discoveries, hfToken, timeout)
			if err != nil {
				return err
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
	generateCmd.Flags().StringVarP(&generatePath, "input", "i", "", "Path to scan")
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
