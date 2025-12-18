package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"aibomgen-cra/internal/completeness"
	"aibomgen-cra/internal/metadata"
)

var completenessCmd = &cobra.Command{
	Use:   "completeness",
	Short: "Compute completeness score for an AIBOM",
	Long:  "Reads an existing CycloneDX AIBOM (json/xml) and scores it against the configured field registry.",
	RunE: func(cmd *cobra.Command, args []string) error {
		bom, err := completeness.ReadBOM(inPath, inFormat)
		if err != nil {
			return err
		}

		res := completeness.Check(bom)

		fmt.Fprintf(cmd.OutOrStdout(), "score=%.1f%% (%d/%d)\n", res.Score*100, res.Passed, res.Total)

		// Optional: show missing keys for debugging
		if len(res.MissingRequired) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "missing required: %s\n", joinKeys(res.MissingRequired))
		}
		if len(res.MissingOptional) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "missing optional: %s\n", joinKeys(res.MissingOptional))
		}
		return nil
	},
}

var (
	inPath   string
	inFormat string
)

func init() {
	rootCmd.AddCommand(completenessCmd)

	completenessCmd.Flags().StringVarP(&inPath, "input", "i", "", "Path to existing AIBOM file (required)")
	completenessCmd.Flags().StringVarP(&inFormat, "format", "f", "auto", "Input BOM format: json|xml|auto (default: auto)")
	_ = completenessCmd.MarkFlagRequired("input")
}

func joinKeys(keys []metadata.Key) string {
	if len(keys) == 0 {
		return ""
	}
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(k.String())
	}
	return b.String()
}
