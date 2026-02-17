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
	"github.com/idlab-discover/AIBoMGen-cli/internal/scanner"
	"github.com/idlab-discover/AIBoMGen-cli/internal/ui"
)

var (
	scanPath         string
	scanOutput       string
	scanOutputFormat string
	scanSpecVersion  string

	// hfMode controls whether metadata is fetched from Hugging Face.
	// Supported values: online|dummy
	scanHfMode       string
	scanHfTimeoutSec int
	scanHfToken      string

	// Logging is controlled via scanLogLevel.
	scanLogLevel string
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan a directory for AI imports and generate AIBOMs",
	Long:  "Scan a directory or repository for AI-related imports (e.g., Hugging Face models) and generate AI-aware BOMs.",
	RunE:  runScan,
}

func runScan(cmd *cobra.Command, args []string) error {
	// Resolve effective log level (from config, env, or flag).
	level := strings.ToLower(strings.TrimSpace(viper.GetString("scan.log-level")))
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
	mode := strings.ToLower(strings.TrimSpace(viper.GetString("scan.hf-mode")))
	if mode == "" {
		mode = "online"
	}
	switch mode {
	case "online", "dummy":
		// ok
	default:
		return fmt.Errorf("invalid --hf-mode %q (expected online|dummy)", mode)
	}

	inputPath := viper.GetString("scan.input")
	if inputPath == "" {
		inputPath = "."
	}

	// Get format from viper
	outputFormat := viper.GetString("scan.format")
	if outputFormat == "" {
		outputFormat = "auto"
	}

	specVersion := viper.GetString("scan.spec")
	outputPath := viper.GetString("scan.output")

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

	// Get HF settings
	hfToken := viper.GetString("scan.hf-token")
	hfTimeout := viper.GetInt("scan.hf-timeout")
	if hfTimeout <= 0 {
		hfTimeout = 10
	}
	timeout := time.Duration(hfTimeout) * time.Second

	ctx := context.Background()

	// Run the scan
	var discoveredBOMs []generator.DiscoveredBOM
	err := runScanDirectory(ctx, inputPath, mode, hfToken, timeout, quiet, &discoveredBOMs)
	if err != nil {
		return err
	}

	// Determine output settings
	output := viper.GetString("scan.output")
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
		genUI := ui.NewGenerateUI(cmd.OutOrStdout(), quiet)
		genUI.PrintNoModelsFound()
		return nil
	}

	genUI := ui.NewGenerateUI(cmd.OutOrStdout(), quiet)
	genUI.PrintSummary(len(written), outputDir, fmtChosen)
	return nil
}

func runScanDirectory(ctx context.Context, inputPath, mode, hfToken string, timeout time.Duration, quiet bool, results *[]generator.DiscoveredBOM) error {
	absTarget, err := filepath.Abs(inputPath)
	if err != nil {
		return err
	}

	if mode == "dummy" {
		if !quiet {
			genUI := ui.NewGenerateUI(os.Stdout, quiet)
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
				workflow.UpdateMessage(processTaskIdx, ui.Dim.Render(fmt.Sprintf("%d/%d", modelsCompleted, totalModels)))
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
	scanCmd.Flags().StringVarP(&scanPath, "input", "i", "", "Path to scan (defaults to current directory)")
	scanCmd.Flags().StringVarP(&scanOutput, "output", "o", "", "Output file path (directory is used)")
	scanCmd.Flags().StringVarP(&scanOutputFormat, "format", "f", "", "Output BOM format: json|xml|auto")
	scanCmd.Flags().StringVar(&scanSpecVersion, "spec", "", "CycloneDX spec version for output (e.g., 1.4, 1.5, 1.6)")
	scanCmd.Flags().StringVar(&scanHfMode, "hf-mode", "", "Hugging Face metadata mode: online|dummy")
	scanCmd.Flags().IntVar(&scanHfTimeoutSec, "hf-timeout", 0, "HTTP timeout in seconds for Hugging Face API")
	scanCmd.Flags().StringVar(&scanHfToken, "hf-token", "", "Hugging Face access token")
	scanCmd.Flags().StringVar(&scanLogLevel, "log-level", "", "Log level: quiet|standard|debug")

	// Bind all flags to viper for config file support
	viper.BindPFlag("scan.input", scanCmd.Flags().Lookup("input"))
	viper.BindPFlag("scan.output", scanCmd.Flags().Lookup("output"))
	viper.BindPFlag("scan.format", scanCmd.Flags().Lookup("format"))
	viper.BindPFlag("scan.spec", scanCmd.Flags().Lookup("spec"))
	viper.BindPFlag("scan.hf-mode", scanCmd.Flags().Lookup("hf-mode"))
	viper.BindPFlag("scan.hf-timeout", scanCmd.Flags().Lookup("hf-timeout"))
	viper.BindPFlag("scan.hf-token", scanCmd.Flags().Lookup("hf-token"))
	viper.BindPFlag("scan.log-level", scanCmd.Flags().Lookup("log-level"))
}
