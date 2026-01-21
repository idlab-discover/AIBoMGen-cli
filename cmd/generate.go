package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/idlab-discover/AIBoMGen-cli/internal/generator"
	bomio "github.com/idlab-discover/AIBoMGen-cli/internal/io"
	"github.com/idlab-discover/AIBoMGen-cli/internal/scanner"
	"github.com/idlab-discover/AIBoMGen-cli/internal/ui"
)

var (
	generatePath         string
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
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate an AI-aware BOM (AIBOM)",
	Long:  "Generate BOM from Hugging Face imports: either scan a directory or provide model ID(s) directly.",
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

	inputPath := viper.GetString("generate.input")
	inputPathProvided := cmd.Flags().Changed("input")

	// Determine which mode we're in
	var useModelIDMode bool
	if modelIDFlagProvided && len(cleanModelIDs) > 0 {
		useModelIDMode = true
		if inputPathProvided && inputPath != "" {
			return fmt.Errorf("cannot specify both --model-id and --input (folder scan)")
		}
	} else if inputPathProvided && inputPath != "" {
		useModelIDMode = false
	} else if len(cleanModelIDs) > 0 {
		useModelIDMode = true
		inputPath = ""
	} else {
		useModelIDMode = false
		if inputPath == "" {
			inputPath = "."
		}
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

	if useModelIDMode {
		// Model ID mode
		err = runModelIDMode(ctx, genUI, cleanModelIDs, mode, hfToken, timeout, quiet, &discoveredBOMs)
	} else {
		// Folder scan mode
		err = runScanMode(ctx, genUI, inputPath, mode, hfToken, timeout, quiet, &discoveredBOMs)
	}

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
	written := make([]string, 0, len(discoveredBOMs))
	for _, d := range discoveredBOMs {
		name := bomMetadataComponentName(d.BOM)
		if strings.TrimSpace(name) == "" {
			name = strings.TrimSpace(d.Discovery.Name)
			if name == "" {
				name = strings.TrimSpace(d.Discovery.ID)
			}
			if name == "" {
				name = "model"
			}
		}

		sanitized := sanitizeComponentName(name)
		fileName := fmt.Sprintf("%s_aibom%s", sanitized, fileExt)
		dest := filepath.Join(outputDir, fileName)

		if err := bomio.WriteBOM(d.BOM, dest, fmtChosen, specVersion); err != nil {
			return err
		}
		written = append(written, dest)
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

func runScanMode(ctx context.Context, genUI *ui.GenerateUI, inputPath, mode, hfToken string, timeout time.Duration, quiet bool, results *[]generator.DiscoveredBOM) error {
	absTarget, err := filepath.Abs(inputPath)
	if err != nil {
		return err
	}

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

	// Create workflow (only if not quiet)
	var workflow *ui.Workflow
	var scanTaskIdx, processTaskIdx, writeTaskIdx int

	if !quiet {
		workflow = ui.NewWorkflow(os.Stdout, "")
		scanTaskIdx = workflow.AddTask("Scanning for AI imports")
		processTaskIdx = workflow.AddTask("Processing models")
		writeTaskIdx = workflow.AddTask("Writing output")
		workflow.Start()
	}

	// Step 1: Scan
	if !quiet && workflow != nil {
		workflow.StartTask(scanTaskIdx, ui.Dim.Render(absTarget))
	}

	discoveries, err := scanner.Scan(absTarget)
	if err != nil {
		if !quiet && workflow != nil {
			workflow.FailTask(scanTaskIdx, err.Error())
			workflow.Stop()
		}
		return err
	}

	if !quiet && workflow != nil {
		workflow.CompleteTask(scanTaskIdx, fmt.Sprintf("found %d model(s)", len(discoveries)))
	}

	if len(discoveries) == 0 {
		if !quiet && workflow != nil {
			workflow.SkipTask(processTaskIdx, "no models to process")
			workflow.SkipTask(writeTaskIdx, "no files to write")
			workflow.Stop()
		}
		*results = []generator.DiscoveredBOM{}
		return nil
	}

	totalModels := len(discoveries)
	modelsCompleted := 0

	// Step 2: Process models (fetch + build combined)
	if !quiet && workflow != nil {
		workflow.StartTask(processTaskIdx, ui.Dim.Render(fmt.Sprintf("0/%d", totalModels)))
	}

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

	boms, err := generator.BuildPerDiscoveryWithProgress(ctx, discoveries, opts)
	if err != nil {
		if !quiet && workflow != nil {
			workflow.FailTask(processTaskIdx, err.Error())
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
	generateCmd.Flags().StringVarP(&generatePath, "input", "i", "", "Path to scan (folder scan mode)")
	generateCmd.Flags().StringSliceVarP(&generateModelIDs, "model-id", "m", []string{}, "Hugging Face model ID(s) (e.g., gpt2 or org/model-name) - can be used multiple times or comma-separated")
	generateCmd.Flags().StringVarP(&generateOutput, "output", "o", "", "Output file path (directory is used)")
	generateCmd.Flags().StringVarP(&generateOutputFormat, "format", "f", "", "Output BOM format: json|xml|auto")
	generateCmd.Flags().StringVar(&generateSpecVersion, "spec", "", "CycloneDX spec version for output (e.g., 1.4, 1.5, 1.6)")
	generateCmd.Flags().StringVar(&hfMode, "hf-mode", "", "Hugging Face metadata mode: online|dummy")
	generateCmd.Flags().IntVar(&hfTimeoutSec, "hf-timeout", 0, "HTTP timeout in seconds for Hugging Face API")
	generateCmd.Flags().StringVar(&hfToken, "hf-token", "", "Hugging Face access token")
	generateCmd.Flags().BoolVar(&enrich, "enrich", false, "Prompt for missing fields and compute completeness (deprecated)")
	generateCmd.Flags().StringVar(&generateLogLevel, "log-level", "", "Log level: quiet|standard|debug")

	// Bind all flags to viper for config file support
	viper.BindPFlag("generate.input", generateCmd.Flags().Lookup("input"))
	viper.BindPFlag("generate.model-ids", generateCmd.Flags().Lookup("model-id"))
	viper.BindPFlag("generate.output", generateCmd.Flags().Lookup("output"))
	viper.BindPFlag("generate.format", generateCmd.Flags().Lookup("format"))
	viper.BindPFlag("generate.spec", generateCmd.Flags().Lookup("spec"))
	viper.BindPFlag("generate.hf-mode", generateCmd.Flags().Lookup("hf-mode"))
	viper.BindPFlag("generate.hf-timeout", generateCmd.Flags().Lookup("hf-timeout"))
	viper.BindPFlag("generate.hf-token", generateCmd.Flags().Lookup("hf-token"))
	viper.BindPFlag("generate.enrich", generateCmd.Flags().Lookup("enrich"))
	viper.BindPFlag("generate.log-level", generateCmd.Flags().Lookup("log-level"))
}

func bomMetadataComponentName(bom *cdx.BOM) string {
	if bom == nil || bom.Metadata == nil || bom.Metadata.Component == nil {
		return ""
	}
	return bom.Metadata.Component.Name
}

func sanitizeComponentName(name string) string {
	if name == "" {
		return "model"
	}
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_' || r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	sanitized := b.String()
	if sanitized == "" {
		return "model"
	}
	return sanitized
}
