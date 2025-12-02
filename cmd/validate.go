package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"aibomgen-cra/internal/validator"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate an existing AIBOM file",
	Long:  "Validates that a CycloneDX AIBOM JSON is well-formed and optionally checks for required model card fields in strict mode.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if validateInput == "" {
			return fmt.Errorf("--input is required (path to AIBOM JSON)")
		}
		bom, errs, err := validator.ValidateFromFile(validateInput, validateStrict, validateFormat)
		if err != nil {
			return fmt.Errorf("failed to decode BOM: %w", err)
		}
		if sErrs := validator.ValidateSpecVersion(bom, validateSpec); len(sErrs) > 0 {
			for _, e := range sErrs {
				fmt.Fprintf(cmd.ErrOrStderr(), "- %s\n", e)
			}
			return fmt.Errorf("validation failed with %d error(s)", len(sErrs))
		}
		if len(errs) > 0 {
			for _, e := range errs {
				fmt.Fprintf(cmd.ErrOrStderr(), "- %s\n", e)
			}
			return fmt.Errorf("validation failed with %d error(s)", len(errs))
		}
		if !quiet {
			if validateStrict {
				fmt.Fprintln(cmd.OutOrStdout(), "AIBOM is valid (strict mode)")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "AIBOM is valid")
			}
		}
		return nil
	},
}

var (
	validateInput  string
	validateStrict bool
	validateFormat string
	validateSpec   string
)

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringVarP(&validateInput, "input", "i", "", "Path to existing AIBOM JSON (required)")
	validateCmd.Flags().BoolVar(&validateStrict, "strict", false, "Enable strict checks for model card completeness")
	validateCmd.Flags().StringVarP(&validateFormat, "format", "f", "auto", "Input BOM format: json|xml|auto (default: auto)")
	validateCmd.Flags().StringVar(&validateSpec, "spec", "", "Require specific CycloneDX specVersion (e.g., 1.3, 1.4, 1.5, 1.6)")
}
