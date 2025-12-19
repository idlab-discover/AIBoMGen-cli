package cmd

import (
	"github.com/spf13/cobra"

	"aibomgen-cra/internal/completeness"
)

var completenessCmd = &cobra.Command{
	Use:   "completeness",
	Short: "Compute completeness score for an AIBOM",
	Long:  "Reads an existing CycloneDX AIBOM (json/xml) and scores it against the configured field registry.",
	RunE: func(cmd *cobra.Command, args []string) error {
		completeness.SetLogger(cmd.OutOrStdout())
		defer completeness.SetLogger(nil)

		bom, err := completeness.ReadBOM(inPath, inFormat)
		if err != nil {
			return err
		}

		res := completeness.Check(bom)
		completeness.PrintReport(res)
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
