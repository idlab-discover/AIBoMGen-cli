package cmd

import (
	"fmt"
	"strings"

	"github.com/idlab-discover/AIBoMGen-cli/internal/completeness"
	bomio "github.com/idlab-discover/AIBoMGen-cli/internal/io"
	"github.com/idlab-discover/AIBoMGen-cli/internal/metadata"
	"github.com/idlab-discover/AIBoMGen-cli/internal/validator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	validateInput          string
	validateFormat         string
	validateStrict         bool
	validateMinScore       float64
	validateCheckModelCard bool
	validateLogLevel       string
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate an existing AIBOM file",
	Long:  "Validates that a CycloneDX AIBOM JSON is well-formed and optionally checks for required model card fields in strict mode.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if validateInput == "" {
			return fmt.Errorf("--input is required")
		}

		// Get log level from viper (respects config file)
		level := strings.ToLower(strings.TrimSpace(viper.GetString("log.level")))
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
			validator.SetLogger(lw)
			if level == "debug" {
				metadata.SetLogger(lw)
				completeness.SetLogger(lw)
			}
		}

		// Read BOM
		bom, err := bomio.ReadBOM(validateInput, validateFormat)
		if err != nil {
			return fmt.Errorf("failed to read BOM: %w", err)
		}

		// Validate
		opts := validator.ValidationOptions{
			StrictMode:           validateStrict,
			MinCompletenessScore: validateMinScore,
			CheckModelCard:       validateCheckModelCard,
		}

		result := validator.Validate(bom, opts)
		validator.PrintReport(result)

		if !result.Valid {
			return fmt.Errorf("validation failed")
		}

		return nil
	},
}

func init() {
	validateCmd.Flags().StringVarP(&validateInput, "input", "i", "", "Path to AIBOM file (required)")
	validateCmd.Flags().StringVarP(&validateFormat, "format", "f", "auto", "Input format: json|xml|auto")
	validateCmd.Flags().BoolVar(&validateStrict, "strict", false, "Strict mode: fail on missing required fields")
	validateCmd.Flags().Float64Var(&validateMinScore, "min-score", 0.0, "Minimum completeness score (0.0-1.0)")
	validateCmd.Flags().BoolVar(&validateCheckModelCard, "check-model-card", true, "Validate model card fields")
	validateCmd.Flags().StringVar(&validateLogLevel, "log-level", "standard", "Log level: quiet|standard|debug (default: standard)")

	validateCmd.MarkFlagRequired("input")

	// Bind flags to viper for config file support
	viper.BindPFlag("validate.strict", validateCmd.Flags().Lookup("strict"))
	viper.BindPFlag("validate.minScore", validateCmd.Flags().Lookup("min-score"))
	viper.BindPFlag("validate.checkModelCard", validateCmd.Flags().Lookup("check-model-card"))
	viper.BindPFlag("log.level", validateCmd.Flags().Lookup("log-level"))
}
