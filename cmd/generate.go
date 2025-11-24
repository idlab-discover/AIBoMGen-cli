package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"aibomgen-cra/internal/generator"
	"aibomgen-cra/internal/scanner"
)

var (
	generatePath   string
	generateOutput string
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate an AI-aware BOM (AIBOM)",
	Long:  "Scans the target path for AI artifacts (model IDs, weight files) and produces a minimal CycloneDX-style AIBOM JSON.",
	RunE: func(cmd *cobra.Command, args []string) error {
		target := generatePath
		if target == "" {
			target = "."
		}
		absTarget, err := filepath.Abs(target)
		if err != nil {
			return err
		}
		comps, err := scanner.Scan(absTarget)
		if err != nil {
			return err
		}
		bom := generator.Build(comps)
		output := generateOutput
		if output == "" {
			output = "dist/aibom.json"
		}
		if err := generator.Write(output, bom); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "AIBOM written to %s (components: %d)\n", output, len(comps))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringVarP(&generatePath, "path", "p", "", "Path to scan (default: current directory)")
	generateCmd.Flags().StringVarP(&generateOutput, "output", "o", "", "Output file path (default: dist/aibom.json)")
}
