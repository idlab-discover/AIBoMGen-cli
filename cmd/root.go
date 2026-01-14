package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/idlab-discover/AIBoMGen-cli/internal/ui"
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "AIBoMGen-cli",
	Short: "BOM Generator for Software Projects using AI {}",
	Long:  longDescription,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initUIAndBanner(cmd)
	},

	// When invoked without a subcommand, show help (with banner) instead of
	// printing a plain usage output.
	RunE: func(cmd *cobra.Command, args []string) error {
		initUIAndBanner(cmd)
		return cmd.Help()
	},
}

var cfgFile string
var noColor bool

// Execute executes the root command.
func Execute() {
	rootCmd.Execute()
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.aibomgen-cli.yaml or ./config/defaults.yaml)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")

	// Ensure `--help` (and help subcommands) show a green banner consistently.
	defaultHelp := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		initUIAndBanner(cmd)
		defaultHelp(cmd, args)
	})

	// Add subcommands
	rootCmd.AddCommand(generateCmd, enrichCmd, validateCmd)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search for config in multiple locations (in order of priority):
		// 1. $HOME/.aibomgen-cli.yaml
		// 2. ./config/defaults.yaml (project local)
		viper.AddConfigPath(home)
		viper.AddConfigPath("./config")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".aibomgen-cli") // for $HOME/.aibomgen-cli.yaml
		viper.SetConfigName("defaults")      // for ./config/defaults.yaml
	}

	// Enable environment variable support (e.g., AIBOMGEN_HUGGINGFACE_TOKEN)
	// Replace dots with underscores: huggingface.token -> AIBOMGEN_HUGGINGFACE_TOKEN
	viper.SetEnvPrefix("AIBOMGEN")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	err := viper.ReadInConfig()

	notFound := &viper.ConfigFileNotFoundError{}
	switch {
	case err != nil && !errors.As(err, notFound):
		cobra.CheckErr(err)
	case err != nil && errors.As(err, notFound):
		// The config file is optional, we shouldn't exit when the config is not found
		break
	default:
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// Simple ASCII banner shown at startup
const bannerASCII = `
  /$$$$$$  /$$$$$$ /$$$$$$$            /$$      /$$  /$$$$$$                                        /$$ /$$
 /$$__  $$|_  $$_/| $$__  $$          | $$$    /$$$ /$$__  $$                                      | $$|__/
| $$  \ $$  | $$  | $$  \ $$  /$$$$$$ | $$$$  /$$$$| $$  \__/  /$$$$$$  /$$$$$$$           /$$$$$$$| $$ /$$
| $$$$$$$$  | $$  | $$$$$$$  /$$__  $$| $$ $$/$$ $$| $$ /$$$$ /$$__  $$| $$__  $$ /$$$$$$ /$$_____/| $$| $$
| $$__  $$  | $$  | $$__  $$| $$  \ $$| $$  $$$| $$| $$|_  $$| $$$$$$$$| $$  \ $$|______/| $$      | $$| $$
| $$  | $$  | $$  | $$  \ $$| $$  | $$| $$\  $ | $$| $$  \ $$| $$_____/| $$  | $$        | $$      | $$| $$
| $$  | $$ /$$$$$$| $$$$$$$/|  $$$$$$/| $$ \/  | $$|  $$$$$$/|  $$$$$$$| $$  | $$        |  $$$$$$$| $$| $$
|__/  |__/|______/|_______/  \______/ |__/     |__/ \______/  \_______/|__/  |__/         \_______/|__/|__/
`
const longDescription = "BOM Generator for Software Projects using AI. Helps PDE manufacturers create accurate Bills of Materials for their AI-based software projects."

func initUIAndBanner(cmd *cobra.Command) {
	ui.Init(noColor)
	if cmd == nil {
		return
	}
	cmd.Root().Long = ui.Color(bannerASCII, ui.FgGreen) + "\n" + longDescription
}
