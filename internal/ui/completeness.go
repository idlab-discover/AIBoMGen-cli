package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// CompletenessReport mirrors the structure from internal/completeness
// to avoid circular imports
type CompletenessReport struct {
	ModelID         string
	Score           float64
	Passed          int
	Total           int
	MissingRequired []FieldKey
	MissingOptional []FieldKey
	DatasetReports  map[string]DatasetReport
}

// DatasetReport mirrors the dataset report structure
type DatasetReport struct {
	DatasetRef      string
	Score           float64
	Passed          int
	Total           int
	MissingRequired []FieldKey
	MissingOptional []FieldKey
}

// FieldKey represents a field identifier
type FieldKey interface {
	String() string
}

// CompletenessUI provides a rich UI for the completeness command
type CompletenessUI struct {
	writer io.Writer
	quiet  bool
}

// NewCompletenessUI creates a new UI handler for the completeness command
func NewCompletenessUI(w io.Writer, quiet bool) *CompletenessUI {
	return &CompletenessUI{
		writer: w,
		quiet:  quiet,
	}
}

// PrintReport renders a beautiful completeness report
func (c *CompletenessUI) PrintReport(report CompletenessReport) {
	if c.quiet {
		return
	}

	var output strings.Builder

	// Header
	output.WriteString(Success.Bold(true).Render("AIBOM Completeness Report"))
	output.WriteString("\n\n")

	// Model Score Section
	output.WriteString(c.renderModelScore(report))
	output.WriteString("\n\n")

	// Missing Fields Section
	if len(report.MissingRequired) > 0 || len(report.MissingOptional) > 0 {
		output.WriteString(c.renderMissingFields(report))
		output.WriteString("\n\n")
	}

	// Dataset Scores Section
	if len(report.DatasetReports) > 0 {
		output.WriteString(c.renderDatasetScores(report.DatasetReports))
		output.WriteString("\n")
	}

	// Wrap in box
	boxed := SuccessBox.Render(output.String())
	fmt.Fprintln(c.writer, boxed)
}

// renderModelScore creates the model score visualization with progress bar
func (c *CompletenessUI) renderModelScore(report CompletenessReport) string {
	var sb strings.Builder

	sb.WriteString(SectionHeader.Render("Model Component"))
	sb.WriteString("\n")

	// Show model ID if available
	if report.ModelID != "" {
		sb.WriteString(FormatKeyValue("ID", Highlight.Render(report.ModelID)))
		sb.WriteString("\n")
	}

	sb.WriteString(FormatKeyValue("Score", c.renderProgressBar(report.Score, 40)+" "+c.renderScorePercentage(report.Score)))
	sb.WriteString("\n")
	sb.WriteString(Dim.Render(fmt.Sprintf("(%d/%d fields present)", report.Passed, report.Total)))

	return sb.String()
}

// renderMissingFields creates the missing fields section with expandable groups
func (c *CompletenessUI) renderMissingFields(report CompletenessReport) string {
	var sb strings.Builder

	// Required Fields
	if len(report.MissingRequired) > 0 {
		sb.WriteString(Error.Render(fmt.Sprintf("▼ Required Fields (%d missing)", len(report.MissingRequired))))
		sb.WriteString("\n")
		for _, field := range report.MissingRequired {
			sb.WriteString("  ")
			sb.WriteString(CrossMark)
			sb.WriteString(" ")
			sb.WriteString(field.String())
			sb.WriteString("\n")
		}
	}

	// Optional Fields
	if len(report.MissingOptional) > 0 {
		if len(report.MissingRequired) > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(Warning.Render(fmt.Sprintf("▼ Optional Fields (%d missing)", len(report.MissingOptional))))
		sb.WriteString("\n")
		for _, field := range report.MissingOptional {
			sb.WriteString("  ")
			sb.WriteString(WarnMark)
			sb.WriteString(" ")
			sb.WriteString(Dim.Render(field.String()))
			sb.WriteString("\n")
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}

// renderDatasetScores creates the dataset scores section
func (c *CompletenessUI) renderDatasetScores(datasets map[string]DatasetReport) string {
	var sb strings.Builder

	sb.WriteString(SectionHeader.Render("Dataset Components"))
	sb.WriteString("\n")

	for dsName, dsReport := range datasets {
		// Dataset name with label
		sb.WriteString(FormatKeyValue("ID", Highlight.Render(dsName)))
		sb.WriteString("\n")

		// Progress bar with label
		sb.WriteString(FormatKeyValue("Score", c.renderProgressBar(dsReport.Score, 40)+" "+c.renderScorePercentage(dsReport.Score)))
		sb.WriteString("\n")
		sb.WriteString(Dim.Render(fmt.Sprintf("(%d/%d fields present)", dsReport.Passed, dsReport.Total)))
		sb.WriteString("\n")

		// Missing fields for this dataset - show underneath each other like model component
		if len(dsReport.MissingRequired) > 0 {
			sb.WriteString("\n")
			sb.WriteString(Error.Render(fmt.Sprintf("▼ Required Fields (%d missing)", len(dsReport.MissingRequired))))
			sb.WriteString("\n")
			for _, field := range dsReport.MissingRequired {
				sb.WriteString("  ")
				sb.WriteString(CrossMark)
				sb.WriteString(" ")
				sb.WriteString(field.String())
				sb.WriteString("\n")
			}
		}
		if len(dsReport.MissingOptional) > 0 {
			if len(dsReport.MissingRequired) > 0 {
				sb.WriteString("\n")
			} else {
				sb.WriteString("\n")
			}
			sb.WriteString(Warning.Render(fmt.Sprintf("▼ Optional Fields (%d missing)", len(dsReport.MissingOptional))))
			sb.WriteString("\n")
			for _, field := range dsReport.MissingOptional {
				sb.WriteString("  ")
				sb.WriteString(WarnMark)
				sb.WriteString(" ")
				sb.WriteString(Dim.Render(field.String()))
				sb.WriteString("\n")
			}
		}

		sb.WriteString("\n")
	}

	return strings.TrimRight(sb.String(), "\n")
}

// renderProgressBar creates a visual progress bar
func (c *CompletenessUI) renderProgressBar(score float64, width int) string {
	filled := int(score * float64(width))
	empty := width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)

	// Color the bar based on score
	var style lipgloss.Style
	if score >= 0.8 {
		style = lipgloss.NewStyle().Foreground(ColorSuccess)
	} else if score >= 0.5 {
		style = lipgloss.NewStyle().Foreground(ColorWarning)
	} else {
		style = lipgloss.NewStyle().Foreground(ColorError)
	}

	return style.Render(bar)
}

// renderScorePercentage formats the score as a percentage
func (c *CompletenessUI) renderScorePercentage(score float64) string {
	percentage := score * 100
	formatted := fmt.Sprintf("%.1f%%", percentage)

	if score >= 0.8 {
		return Success.Render(formatted)
	} else if score >= 0.5 {
		return Warning.Render(formatted)
	}
	return Error.Render(formatted)
}

// formatFieldKeys formats field keys as a comma-separated string
func (c *CompletenessUI) formatFieldKeys(keys []FieldKey) string {
	if len(keys) == 0 {
		return ""
	}
	names := make([]string, len(keys))
	for i, k := range keys {
		names[i] = k.String()
	}
	return strings.Join(names, ", ")
}

// PrintSimpleReport prints a minimal text report (fallback for quiet mode or issues)
func (c *CompletenessUI) PrintSimpleReport(report CompletenessReport) {
	fmt.Fprintf(c.writer, "Model score: %.1f%% (%d/%d)\n", report.Score*100, report.Passed, report.Total)

	if len(report.MissingRequired) > 0 {
		fmt.Fprintf(c.writer, "Missing required: %s\n", c.formatFieldKeys(report.MissingRequired))
	}
	if len(report.MissingOptional) > 0 {
		fmt.Fprintf(c.writer, "Missing optional: %s\n", c.formatFieldKeys(report.MissingOptional))
	}

	if len(report.DatasetReports) > 0 {
		fmt.Fprintln(c.writer, "\nDatasets:")
		for dsName, dsReport := range report.DatasetReports {
			fmt.Fprintf(c.writer, "  %s: %.1f%% (%d/%d)\n", dsName, dsReport.Score*100, dsReport.Passed, dsReport.Total)
		}
	}
}
