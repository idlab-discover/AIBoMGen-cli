package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/idlab-discover/AIBoMGen-cli/internal/generator"
	bomio "github.com/idlab-discover/AIBoMGen-cli/internal/io"
	"github.com/idlab-discover/AIBoMGen-cli/internal/ui"
)

var (
	generateOutput       string
	generateOutputFormat string
	generateSpecVersion  string
	generateModelIDs     []string

	// hfMode controls whether metadata is fetched from Hugging Face.
	// Supported values: online|dummy
	hfMode       string
	hfTimeoutSec int
	hfToken      string

	enrich bool
	// Logging is controlled via generateLogLevel.
	generateLogLevel string

	// interactive enables the interactive model selector
	interactive bool
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate an AI-aware BOM (AIBOM) from Hugging Face model IDs",
	Long:  "Generate BOM from Hugging Face model ID(s). Use --model-id to specify models directly or --interactive for a model selector. Use 'scan' command to scan directories for AI imports.",
	RunE:  runGenerate,
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Resolve effective log level (from config, env, or flag).
	level := strings.ToLower(strings.TrimSpace(viper.GetString("generate.log-level")))
	if level == "" {
		level = "standard"
	}
	switch level {
	case "quiet", "standard", "debug":
		// ok
	default:
		return fmt.Errorf("invalid --log-level %q (expected quiet|standard|debug)", level)
	}

	quiet := level == "quiet"

	// Resolve effective HF mode (from config, env, or flag).
	mode := strings.ToLower(strings.TrimSpace(viper.GetString("generate.hf-mode")))
	if mode == "" {
		mode = "online"
	}
	switch mode {
	case "online", "dummy":
		// ok
	default:
		return fmt.Errorf("invalid --hf-mode %q (expected online|dummy)", mode)
	}

	// Check if --interactive was explicitly provided
	interactiveMode := viper.GetBool("generate.interactive")

	// Check if --model-id was explicitly provided on the command line
	modelIDFlagProvided := cmd.Flags().Changed("model-id")

	// Get model IDs from viper (respects config file and CLI flag)
	modelIDs := viper.GetStringSlice("generate.model-ids")
	// Filter out empty strings
	var cleanModelIDs []string
	for _, id := range modelIDs {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			cleanModelIDs = append(cleanModelIDs, trimmed)
		}
	}

	// Interactive mode validation
	if interactiveMode {
		if modelIDFlagProvided {
			return fmt.Errorf("--interactive cannot be used with --model-id")
		}
	}

	// Validate that we have either model IDs or interactive mode
	if !interactiveMode && len(cleanModelIDs) == 0 {
		return fmt.Errorf("either --model-id or --interactive is required. Use 'scan' command to scan directories")
	}

	// Get format from viper
	outputFormat := viper.GetString("generate.format")
	if outputFormat == "" {
		outputFormat = "auto"
	}

	specVersion := viper.GetString("generate.spec")
	outputPath := viper.GetString("generate.output")

	// Fail fast on format/extension mismatch
	if outputPath != "" && outputFormat != "" && outputFormat != "auto" {
		ext := filepath.Ext(outputPath)
		if outputFormat == "xml" && ext == ".json" {
			return fmt.Errorf("output path extension %q does not match format %q", ext, outputFormat)
		}
		if outputFormat == "json" && ext == ".xml" {
			return fmt.Errorf("output path extension %q does not match format %q", ext, outputFormat)
		}
	}

	// Wire internal package logging for debug mode
	if level == "debug" {
		// Logging removed
	}

	// Get HF settings
	hfToken := viper.GetString("generate.hf-token")
	hfTimeout := viper.GetInt("generate.hf-timeout")
	if hfTimeout <= 0 {
		hfTimeout = 10
	}
	timeout := time.Duration(hfTimeout) * time.Second

	// Create UI handler
	genUI := ui.NewGenerateUI(cmd.OutOrStdout(), quiet)

	var discoveredBOMs []generator.DiscoveredBOM
	var err error

	ctx := context.Background()

	if interactiveMode {
		// Interactive mode: show model selector
		selectedModels, err := ui.RunModelSelector(ui.ModelSelectorConfig{
			HFToken: hfToken,
			Timeout: timeout,
		})
		if err != nil {
			return err
		}
		if len(selectedModels) == 0 {
			return fmt.Errorf("no models selected")
		}
		cleanModelIDs = selectedModels
	}

	// Generate BOMs from model IDs
	err = runModelIDMode(ctx, genUI, cleanModelIDs, mode, hfToken, timeout, quiet, &discoveredBOMs)
	if err != nil {
		return err
	}

	// Deprecation warning
	if viper.GetBool("generate.enrich") {
		fmt.Fprintf(cmd.ErrOrStderr(), "%s --enrich flag is deprecated. Use 'aibomgen-cli enrich' command instead.\n", ui.GetWarnMark())
	}

	// Determine output settings
	output := viper.GetString("generate.output")
	if output == "" {
		if outputFormat == "xml" {
			output = "dist/aibom.xml"
		} else {
			output = "dist/aibom.json"
		}
	}

	fmtChosen := outputFormat
	if fmtChosen == "auto" || fmtChosen == "" {
		ext := filepath.Ext(output)
		if ext == ".xml" {
			fmtChosen = "xml"
		} else {
			fmtChosen = "json"
		}
	}

	outputDir := filepath.Dir(output)
	if outputDir == "" {
		outputDir = "."
	}
	outputDir = filepath.Clean(outputDir)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	fileExt := ".json"
	if fmtChosen == "xml" {
		fileExt = ".xml"
	}

	// Write output files
	written, err := bomio.WriteOutputFiles(discoveredBOMs, outputDir, fileExt, fmtChosen, specVersion)
	if err != nil {
		return err
	}

	// Print summary
	if len(written) == 0 {
		genUI.PrintNoModelsFound()
		return nil
	}

	genUI.PrintSummary(len(written), outputDir, fmtChosen)
	return nil
}

func runModelIDMode(ctx context.Context, genUI *ui.GenerateUI, modelIDs []string, mode, hfToken string, timeout time.Duration, quiet bool, results *[]generator.DiscoveredBOM) error {
	if mode == "dummy" {
		if !quiet {
			genUI.LogStep("info", "Using dummy mode (no API calls)")
		}
		boms, err := generator.BuildDummyBOM()
		if err != nil {
			return err
		}
		*results = boms
		return nil
	}

	// Track completed models to print at the end
	type modelResult struct {
		id       string
		datasets int
		err      string
	}
	var completedModels []modelResult

	// Create workflow with combined processing step
	var workflow *ui.Workflow
	var processTaskIdx, writeTaskIdx int

	if !quiet {
		workflow = ui.NewWorkflow(os.Stdout, "")
		processTaskIdx = workflow.AddTask("Processing models")
		writeTaskIdx = workflow.AddTask("Writing output")
		workflow.Start()
	}

	totalModels := len(modelIDs)
	modelsCompleted := 0

	// Start processing
	if !quiet && workflow != nil {
		workflow.StartTask(processTaskIdx, ui.Dim.Render(fmt.Sprintf("0/%d", totalModels)))
	}

	// Progress callback to update UI
	onProgress := func(evt generator.ProgressEvent) {
		if quiet || workflow == nil {
			return
		}

		switch evt.Type {
		case generator.EventFetchStart:
			workflow.UpdateMessage(processTaskIdx, ui.Dim.Render(fmt.Sprintf("%d/%d: %s (fetching)", modelsCompleted, totalModels, evt.ModelID)))
		case generator.EventBuildStart:
			workflow.UpdateMessage(processTaskIdx, ui.Dim.Render(fmt.Sprintf("%d/%d: %s (building)", modelsCompleted, totalModels, evt.ModelID)))
		case generator.EventDatasetStart:
			workflow.UpdateMessage(processTaskIdx, ui.Dim.Render(fmt.Sprintf("%d/%d: %s → %s", modelsCompleted, totalModels, evt.ModelID, evt.Message)))
		case generator.EventModelComplete:
			modelsCompleted++
			completedModels = append(completedModels, modelResult{id: evt.ModelID, datasets: evt.Datasets})
			if modelsCompleted < totalModels {
				workflow.UpdateMessage(processTaskIdx, ui.Dim.Render(fmt.Sprintf("%d/%d complete", modelsCompleted, totalModels)))
			}
		case generator.EventError:
			completedModels = append(completedModels, modelResult{id: evt.ModelID, err: evt.Message})
		}
	}

	opts := generator.GenerateOptions{
		HFToken:    hfToken,
		Timeout:    timeout,
		OnProgress: onProgress,
	}

	boms, err := generator.BuildFromModelIDsWithProgress(ctx, modelIDs, opts)
	if err != nil {
		if !quiet && workflow != nil {
			workflow.Stop()
		}
		return err
	}

	if !quiet && workflow != nil {
		workflow.CompleteTask(processTaskIdx, fmt.Sprintf("%d model(s)", len(boms)))
		workflow.StartTask(writeTaskIdx, "")
		workflow.CompleteTask(writeTaskIdx, fmt.Sprintf("%d file(s)", len(boms)))
		workflow.Stop()

		// Print individual model results after workflow completes
		fmt.Println()
		for _, m := range completedModels {
			if m.err != "" {
				fmt.Printf("  %s %s %s\n", ui.GetCrossMark(), ui.Highlight.Render(m.id), ui.Error.Render("→ "+m.err))
			} else if m.datasets > 0 {
				fmt.Printf("  %s %s %s\n", ui.GetCheckMark(), ui.Highlight.Render(m.id), ui.Dim.Render(fmt.Sprintf("→ %d dataset(s)", m.datasets)))
			} else {
				fmt.Printf("  %s %s\n", ui.GetCheckMark(), ui.Highlight.Render(m.id))
			}
		}
	}

	*results = boms
	return nil
}

func init() {
	generateCmd.Flags().StringSliceVarP(&generateModelIDs, "model-id", "m", []string{}, "Hugging Face model ID(s) (e.g., gpt2 or org/model-name) - can be used multiple times or comma-separated")
	generateCmd.Flags().StringVarP(&generateOutput, "output", "o", "", "Output file path (directory is used)")
	generateCmd.Flags().StringVarP(&generateOutputFormat, "format", "f", "", "Output BOM format: json|xml|auto")
	generateCmd.Flags().StringVar(&generateSpecVersion, "spec", "", "CycloneDX spec version for output (e.g., 1.4, 1.5, 1.6)")
	generateCmd.Flags().StringVar(&hfMode, "hf-mode", "", "Hugging Face metadata mode: online|dummy")
	generateCmd.Flags().IntVar(&hfTimeoutSec, "hf-timeout", 0, "HTTP timeout in seconds for Hugging Face API")
	generateCmd.Flags().StringVar(&hfToken, "hf-token", "", "Hugging Face access token")
	generateCmd.Flags().BoolVar(&enrich, "enrich", false, "Prompt for missing fields and compute completeness (deprecated)")
	generateCmd.Flags().StringVar(&generateLogLevel, "log-level", "", "Log level: quiet|standard|debug")
	generateCmd.Flags().BoolVar(&interactive, "interactive", false, "Interactive model selector (cannot be used with --model-id)")

	// Bind all flags to viper for config file support
	viper.BindPFlag("generate.model-ids", generateCmd.Flags().Lookup("model-id"))
	viper.BindPFlag("generate.output", generateCmd.Flags().Lookup("output"))
	viper.BindPFlag("generate.format", generateCmd.Flags().Lookup("format"))
	viper.BindPFlag("generate.spec", generateCmd.Flags().Lookup("spec"))
	viper.BindPFlag("generate.hf-mode", generateCmd.Flags().Lookup("hf-mode"))
	viper.BindPFlag("generate.hf-timeout", generateCmd.Flags().Lookup("hf-timeout"))
	viper.BindPFlag("generate.hf-token", generateCmd.Flags().Lookup("hf-token"))
	viper.BindPFlag("generate.enrich", generateCmd.Flags().Lookup("enrich"))
	viper.BindPFlag("generate.log-level", generateCmd.Flags().Lookup("log-level"))
	viper.BindPFlag("generate.interactive", generateCmd.Flags().Lookup("interactive"))
}
