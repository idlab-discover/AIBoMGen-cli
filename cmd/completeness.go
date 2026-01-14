package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/idlab-discover/AIBoMGen-cli/internal/completeness"
	bomio "github.com/idlab-discover/AIBoMGen-cli/internal/io"
	"github.com/idlab-discover/AIBoMGen-cli/internal/metadata"
)

var completenessCmd = &cobra.Command{
	Use:   "completeness",
	Short: "Compute completeness score for an AIBOM",
	Long:  "Reads an existing CycloneDX AIBOM (json/xml) and scores it against the configured field registry.",
	RunE: func(cmd *cobra.Command, args []string) error {

		// Get log level from viper (respects config file and CLI flag)
		level := strings.ToLower(strings.TrimSpace(viper.GetString("completeness.log-level")))
		if level == "" {
			level = "standard"
		}
		switch level {
		case "quiet", "standard", "debug":
			// ok
		default:
			return fmt.Errorf("invalid --log-level %q (expected quiet|standard|debug)", level)
		}

		// Wire internal package logging based on log level.
		if level != "quiet" {
			lw := cmd.ErrOrStderr()
			completeness.SetLogger(lw)
			if level == "debug" {
				metadata.SetLogger(lw)
			}
		}

		// Get input path and format from viper
		inputPath := viper.GetString("completeness.input")
		inputFormat := viper.GetString("completeness.format")
		if inputFormat == "" {
			inputFormat = "auto"
		}

		bom, err := bomio.ReadBOM(inputPath, inputFormat)
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
	completenessCmd.Flags().StringVarP(&inFormat, "format", "f", "", "Input BOM format: json|xml|auto")
	completenessCmd.Flags().StringVar(&completenessLogLevel, "log-level", "", "Log level: quiet|standard|debug")

	_ = completenessCmd.MarkFlagRequired("input")

	// Bind all flags to viper for config file support
	viper.BindPFlag("completeness.input", completenessCmd.Flags().Lookup("input"))
	viper.BindPFlag("completeness.format", completenessCmd.Flags().Lookup("format"))
	viper.BindPFlag("completeness.log-level", completenessCmd.Flags().Lookup("log-level"))
}
