package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"aibomgen-cra/internal/enricher"
	"aibomgen-cra/internal/fetcher"
	"aibomgen-cra/internal/generator"
	"aibomgen-cra/internal/scanner"
)

var (
	generatePath         string
	generateOutput       string
	generateOutputFormat string
	generateSpecVersion  string
	hfOnline             bool
	hfTimeoutSec         int
	hfToken              string
	hfCacheDir           string
	enrich               bool
	quiet                bool
	dummy                bool
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate an AI-aware BOM (AIBOM)",
	Long:  "Scans the target path for AI Hugginface imports and produces a CycloneDX-style AIBOM JSON.",
	RunE: func(cmd *cobra.Command, args []string) error {
		target := generatePath
		if target == "" {
			target = "."
		}
		absTarget, err := filepath.Abs(target)
		if err != nil {
			return err
		}
		// Fail fast on explicit format/extension mismatch before scanning
		if generateOutput != "" && generateOutputFormat != "" && generateOutputFormat != "auto" {
			ext := filepath.Ext(generateOutput)
			if generateOutputFormat == "xml" && ext == ".json" {
				return fmt.Errorf("output path extension %q does not match format %q", ext, generateOutputFormat)
			}
			if generateOutputFormat == "json" && ext == ".xml" {
				return fmt.Errorf("output path extension %q does not match format %q", ext, generateOutputFormat)
			}
		}
		// Configure default fetcher: HuggingFace (online) or Dummy (offline)
		if dummy || !hfOnline {
			fetcher.SetDefault(fetcher.NewDummyFetcher())
		} else {
			token := hfToken
			if hfTimeoutSec <= 0 {
				hfTimeoutSec = 10
			}
			timeout := time.Duration(hfTimeoutSec) * time.Second
			hf := fetcher.NewHuggingFaceFetcher(timeout, token, hfCacheDir)
			fetcher.SetDefault(hf)
		}
		// Wire scanner and generator logging to command output
		if !quiet {
			lw := cmd.ErrOrStderr()
			// Configure package-level fetcher logger per project convention
			fetcher.SetLogger(lw)
			scanner.SetLogger(lw)
			generator.SetLogger(lw)
		}
		comps, err := scanner.Scan(absTarget)
		if err != nil {
			return err
		}
		// Add a blank line between scan summary and fetch/model context logs
		if !quiet {
			fmt.Fprintln(cmd.ErrOrStderr())
		}
		componentBOMs := generator.BuildPerComponent(comps)
		// Fetchers now log their findings; no aggregate logging here
		// Optionally prompt user to fill missing fields and annotate completeness
		for _, compBOM := range componentBOMs {
			enricher.InteractiveCompleteBOM(compBOM.BOM, enrich, cmd.InOrStdin(), cmd.OutOrStdout())
		}
		output := generateOutput
		if output == "" {
			// Default extension based on requested format (json unless explicitly xml)
			if generateOutputFormat == "xml" {
				output = "dist/aibom.xml"
			} else {
				output = "dist/aibom.json"
			}
		}
		fmtChosen := generateOutputFormat
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
		fileExt := ".json"
		if fmtChosen == "xml" {
			fileExt = ".xml"
		}
		written := make([]string, 0, len(componentBOMs))
		for _, compBOM := range componentBOMs {
			sanitized := sanitizeComponentName(compBOM.Component.Name)
			fileName := fmt.Sprintf("%s_aibom%s", sanitized, fileExt)
			dest := filepath.Join(outputDir, fileName)
			if err := generator.WriteWithFormatAndSpec(dest, compBOM.BOM, fmtChosen, generateSpecVersion); err != nil {
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
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringVarP(&generatePath, "path", "p", "", "Path to scan (default: current directory)")
	generateCmd.Flags().StringVarP(&generateOutput, "output", "o", "", "Output file path (default: dist/aibom.json)")
	generateCmd.Flags().StringVarP(&generateOutputFormat, "format", "f", "auto", "Output BOM format: json|xml|auto (default: auto)")
	generateCmd.Flags().StringVar(&generateSpecVersion, "spec", "", "CycloneDX spec version for output (e.g., 1.3, 1.4, 1.5, 1.6)")
	generateCmd.Flags().BoolVar(&hfOnline, "hf-online", true, "Enable Hugging Face API for model metadata")
	generateCmd.Flags().IntVar(&hfTimeoutSec, "hf-timeout", 10, "HTTP timeout in seconds for Hugging Face API")
	generateCmd.Flags().StringVar(&hfToken, "hf-token", "", "Hugging Face access token (string)")
	generateCmd.Flags().StringVar(&hfCacheDir, "hf-cache-dir", "", "Cache directory for Hugging Face downloads (optional)")
	generateCmd.Flags().BoolVar(&enrich, "enrich", false, "Prompt for missing fields and compute completeness")
	generateCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Suppress scan/fetch logs (errors still shown)")
	generateCmd.Flags().BoolVar(&dummy, "dummy", false, "Use dummy fetcher instead of Hugging Face API")
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
