package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"aibomgen-cra/internal/enricher"
	"aibomgen-cra/internal/fetcher"
	cons "aibomgen-cra/internal/fetcher/considerations"
	params "aibomgen-cra/internal/fetcher/parameters"
	quant "aibomgen-cra/internal/fetcher/quantitative"
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
	hfTokenEnv           string
	enrich               bool
	quiet                bool
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
		// Configure orchestrator and optionally enable Hugging Face for parameters
		var p fetcher.ParametersFetcher
		var q fetcher.QuantitativeFetcher
		var c fetcher.ConsiderationsFetcher
		if hfOnline {
			token := os.Getenv(hfTokenEnv)
			if hfTimeoutSec <= 0 {
				hfTimeoutSec = 10
			}
			timeout := time.Duration(hfTimeoutSec) * time.Second
			pf := params.NewHuggingFaceParametersFetcher(timeout, token)
			qf := quant.NewHuggingFaceQuantFetcher(timeout, token)
			cf := cons.NewHuggingFaceConsiderationsFetcher(timeout, token)
			// Attach command output as logger
			if !quiet {
				lw := cmd.ErrOrStderr()
				pf.SetLogger(lw)
				qf.SetLogger(lw)
				cf.SetLogger(lw)
			}
			p, q, c = pf, qf, cf
		}
		fetcher.SetDefault(fetcher.NewOrchestrator(p, q, c))
		// Wire scanner and generator logging to command output
		if !quiet {
			lw := cmd.ErrOrStderr()
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
		bom := generator.Build(comps)
		// Fetchers now log their findings; no aggregate logging here
		// Optionally prompt user to fill missing fields and annotate completeness
		enricher.InteractiveCompleteBOM(bom, enrich, cmd.InOrStdin(), cmd.OutOrStdout())
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
		if err := generator.WriteWithFormatAndSpec(output, bom, fmtChosen, generateSpecVersion); err != nil {
			return err
		}
		// Add a blank line before the final success message for readability
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintf(cmd.OutOrStdout(), "AIBOM written to %s (format: %s, components: %d)\n", output, fmtChosen, len(comps))
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
	generateCmd.Flags().StringVar(&hfTokenEnv, "hf-token-env", "HUGGINGFACE_TOKEN", "Env var name containing Hugging Face token")
	generateCmd.Flags().BoolVar(&enrich, "enrich", false, "Prompt for missing fields and compute completeness")
	generateCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Suppress scan/fetch logs (errors still shown)")
}
