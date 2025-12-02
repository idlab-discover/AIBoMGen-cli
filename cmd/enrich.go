package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	cdx "github.com/CycloneDX/cyclonedx-go"

	"aibomgen-cra/internal/enricher"
	"aibomgen-cra/internal/generator"
)

// enrichCmd represents the enrich command
var enrichCmd = &cobra.Command{
	Use:   "enrich",
	Short: "Enrich an existing AIBOM with additional metadata",
	Long:  `Enrich an existing AIBOM with additional metadata by prompting the user for inputs on empty or missing fields.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		input := enrichInput
		if input == "" {
			return fmt.Errorf("--input is required (path to existing AIBOM JSON)")
		}
		// Fail fast on explicit format/extension mismatch before any decoding
		outPath := enrichOutput
		if outPath == "" {
			outPath = input // overwrite implies same file path
		}
		if enrichOutputFormat != "" && enrichOutputFormat != "auto" {
			ext := strings.ToLower(filepath.Ext(outPath))
			if enrichOutputFormat == "xml" && ext == ".json" {
				return fmt.Errorf("output path extension %q does not match format %q", ext, enrichOutputFormat)
			}
			if enrichOutputFormat == "json" && ext == ".xml" {
				return fmt.Errorf("output path extension %q does not match format %q", ext, enrichOutputFormat)
			}
		}
		f, err := os.Open(input)
		if err != nil {
			return err
		}
		defer f.Close()
		// Decode with auto detection (xml/json)
		var bom cdx.BOM
		originalFmt, err := decodeAuto(f, input, &bom)
		if err != nil {
			return fmt.Errorf("failed to decode BOM: %w", err)
		}
		// Always run interactive enrichment for missing fields
		enricher.InteractiveCompleteBOM(&bom, true, cmd.InOrStdin(), cmd.OutOrStdout())
		output := enrichOutput
		if output == "" {
			output = input // overwrite by default
		}
		// Decide output format
		outFmt := enrichOutputFormat
		if outFmt == "auto" || outFmt == "" {
			if output == input { // overwriting original, keep original format
				outFmt = originalFmt
			} else {
				ext := strings.ToLower(filepath.Ext(output))
				if ext == ".xml" {
					outFmt = "xml"
				} else {
					outFmt = originalFmt // default to original
				}
			}
		}
		if err := generator.WriteWithFormatAndSpec(output, &bom, outFmt, enrichSpecVersion); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout())
		action := "written to"
		if output == input {
			action = "overwritten"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "AIBOM %s %s (format: %s)\n", action, output, outFmt)
		return nil
	},
}

var (
	enrichInput        string
	enrichOutput       string
	enrichOutputFormat string
	enrichSpecVersion  string
)

func init() {
	rootCmd.AddCommand(enrichCmd)
	enrichCmd.Flags().StringVarP(&enrichInput, "input", "i", "", "Path to existing AIBOM JSON (required)")
	enrichCmd.Flags().StringVarP(&enrichOutput, "output", "o", "", "Output file path (default: overwrite input)")
	enrichCmd.Flags().StringVarP(&enrichOutputFormat, "format", "f", "auto", "Output BOM format: json|xml|auto (default: auto)")
	enrichCmd.Flags().StringVar(&enrichSpecVersion, "spec", "", "CycloneDX spec version for output (e.g., 1.3, 1.4, 1.5, 1.6)")
	enrichCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-essential logs")
}

// decodeAuto detects format from extension or content and decodes BOM.
func decodeAuto(f *os.File, filename string, bom *cdx.BOM) (string, error) {
	if _, err := f.Seek(0, 0); err != nil {
		return "", err
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	ext := strings.ToLower(filepath.Ext(filename))
	format := ""
	if ext == ".xml" {
		format = "xml"
	} else if ext == ".json" {
		format = "json"
	}
	if format == "" {
		trimmed := strings.TrimSpace(string(data))
		if strings.HasPrefix(trimmed, "<") {
			format = "xml"
		} else {
			format = "json"
		}
	}
	fileFmt := cdx.BOMFileFormatJSON
	if format == "xml" {
		fileFmt = cdx.BOMFileFormatXML
	}
	reader := bytes.NewReader(data)
	dec := cdx.NewBOMDecoder(reader, fileFmt)
	if err := dec.Decode(bom); err != nil {
		return "", err
	}
	return format, nil
}
