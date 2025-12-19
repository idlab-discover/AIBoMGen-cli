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

		// Resolve effective log level.
		level := strings.ToLower(strings.TrimSpace(completenessLogLevel))
		if level == "" {
			level = "standard"
		}
		switch level {
		case "quiet", "standard", "debug":
			// ok
		default:
			return fmt.Errorf("invalid --log-level %q (expected quiet|standard|debug)", completenessLogLevel)
		}

		// Wire internal package logging based on log level.
		if level != "quiet" {
			lw := cmd.ErrOrStderr()
			completeness.SetLogger(lw)
			if level == "debug" {
				metadata.SetLogger(lw)
			}
		}

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
	inPath               string
	inFormat             string
	completenessLogLevel string
)

func init() {
	rootCmd.AddCommand(completenessCmd)

	completenessCmd.Flags().StringVarP(&inPath, "input", "i", "", "Path to existing AIBOM file (required)")
	completenessCmd.Flags().StringVarP(&inFormat, "format", "f", "auto", "Input BOM format: json|xml|auto (default: auto)")
	completenessCmd.Flags().StringVar(&completenessLogLevel, "log-level", "standard", "Log level: quiet|standard|debug (default: standard)")

	_ = completenessCmd.MarkFlagRequired("input")
}
