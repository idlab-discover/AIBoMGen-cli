package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"aibomgen-cra/internal/ui"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "aibomgen-cra",
	Short: "BOM Generator for Software Projects using AI",
	Long:  "BOM Generator for Software Projects using AI. Helps PDE manufacturers create accurate Bills of Materials for their AI-based software projects.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize UI configuration from flags only
		ui.Init(noColor)
		if !noBanner {
			// Print banner to stderr once at startup
			fmt.Fprintln(cmd.ErrOrStderr(), ui.Color(bannerASCII, ui.FgMagenta))
		}
	},
}

var cfgFile string
var noColor bool
var noBanner bool

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.aibomgen-cra.yaml)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().BoolVar(&noBanner, "no-banner", false, "Disable ASCII art banner on startup")
}

// Simple ASCII banner shown at startup
var bannerASCII = `
░█▀█░▀█▀░█▀▄░█▀█░█▄█░█▀▀░█▀▀░█▀█░░░█▀▀░█▀▄░█▀█
░█▀█░░█░░█▀▄░█░█░█░█░█░█░█▀▀░█░█░░░█░░░█▀▄░█▀█
░▀░▀░▀▀▀░▀▀░░▀▀▀░▀░▀░▀▀▀░▀▀▀░▀░▀░░░▀▀▀░▀░▀░▀░▀
`
