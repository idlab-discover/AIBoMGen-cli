package enricher

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

// InteractiveCompleteBOM optionally prompts to fill missing model card fields, then annotates completeness.
func InteractiveCompleteBOM(bom *cdx.BOM, interactive bool, in io.Reader, out io.Writer) {
	if bom == nil || bom.Components == nil {
		return
	}
	// Count ML model components for clearer prompts
	total := 0
	for i := range *bom.Components {
		if (*bom.Components)[i].Type == cdx.ComponentTypeMachineLearningModel {
			total++
		}
	}
	idx := 0
	for i := range *bom.Components {
		comp := &(*bom.Components)[i]
		if comp.Type != cdx.ComponentTypeMachineLearningModel {
			continue
		}
		idx++
		if comp.ModelCard == nil {
			comp.ModelCard = &cdx.MLModelCard{}
		}
		if interactive {
			// Show context so the user knows which model they are enriching
			path := getProperty(comp, "aibomgen.path")
			// Add a separating blank line before enrichment header
			fmt.Fprintln(out)
			if path != "" {
				fmt.Fprintf(out, "Enriching model [%d/%d]: %s (path: %s)\n", idx, total, comp.Name, path)
			} else {
				fmt.Fprintf(out, "Enriching model [%d/%d]: %s\n", idx, total, comp.Name)
			}
			interactiveFill(comp, in, out)
		}
		annotateCompleteness(comp)
	}
}

func annotateCompleteness(comp *cdx.Component) {
	filled, total := completenessCounts(comp)
	percent := 0
	if total > 0 {
		percent = int(float64(filled) / float64(total) * 100.0)
	}
	prop := cdx.Property{Name: "aibomgen.completeness", Value: fmt.Sprintf("%d%%", percent)}
	if comp.Properties == nil {
		comp.Properties = &[]cdx.Property{prop}
		return
	}
	props := append(*comp.Properties, prop)
	comp.Properties = &props
}

func completenessCounts(comp *cdx.Component) (filled, total int) {
	total = 7 // task, datasets, inputs, outputs, perf metrics, users, usecases
	card := comp.ModelCard
	// Task
	if card != nil && card.ModelParameters != nil && card.ModelParameters.Task != "" {
		filled++
	}
	// Datasets
	if card != nil && card.ModelParameters != nil && card.ModelParameters.Datasets != nil && len(*card.ModelParameters.Datasets) > 0 {
		filled++
	}
	// Inputs
	if card != nil && card.ModelParameters != nil && card.ModelParameters.Inputs != nil && len(*card.ModelParameters.Inputs) > 0 {
		filled++
	}
	// Outputs
	if card != nil && card.ModelParameters != nil && card.ModelParameters.Outputs != nil && len(*card.ModelParameters.Outputs) > 0 {
		filled++
	}
	// Performance metrics
	if card != nil && card.QuantitativeAnalysis != nil && card.QuantitativeAnalysis.PerformanceMetrics != nil && len(*card.QuantitativeAnalysis.PerformanceMetrics) > 0 {
		filled++
	}
	// Users
	if card != nil && card.Considerations != nil && card.Considerations.Users != nil && len(*card.Considerations.Users) > 0 {
		filled++
	}
	// UseCases
	if card != nil && card.Considerations != nil && card.Considerations.UseCases != nil && len(*card.Considerations.UseCases) > 0 {
		filled++
	}
	return
}

func interactiveFill(comp *cdx.Component, in io.Reader, out io.Writer) {
	r := bufio.NewReader(in)
	card := comp.ModelCard
	if card.ModelParameters == nil {
		card.ModelParameters = &cdx.MLModelParameters{}
	}

	// Task
	if card.ModelParameters.Task == "" {
		fmt.Fprint(out, "Task (e.g., text-classification): ")
		if s := readLine(r); s != "" {
			card.ModelParameters.Task = s
		}
	}
	// Datasets
	if card.ModelParameters.Datasets == nil || len(*card.ModelParameters.Datasets) == 0 {
		fmt.Fprint(out, "Datasets (comma-separated refs like dataset:imdb): ")
		if s := readLine(r); s != "" {
			refs := splitCSV(s)
			var ds []cdx.MLDatasetChoice
			for _, ref := range refs {
				if ref != "" {
					ds = append(ds, cdx.MLDatasetChoice{Ref: ref})
				}
			}
			card.ModelParameters.Datasets = &ds
		}
	}
	// Inputs
	if card.ModelParameters.Inputs == nil || len(*card.ModelParameters.Inputs) == 0 {
		fmt.Fprint(out, "Input formats (comma-separated MIME, e.g., text/plain): ")
		if s := readLine(r); s != "" {
			fmts := splitCSV(s)
			var inParams []cdx.MLInputOutputParameters
			for _, f := range fmts {
				if f != "" {
					inParams = append(inParams, cdx.MLInputOutputParameters{Format: f})
				}
			}
			card.ModelParameters.Inputs = &inParams
		}
	}
	// Outputs
	if card.ModelParameters.Outputs == nil || len(*card.ModelParameters.Outputs) == 0 {
		fmt.Fprint(out, "Output formats (comma-separated MIME, e.g., classification-label): ")
		if s := readLine(r); s != "" {
			fmts := splitCSV(s)
			var outParams []cdx.MLInputOutputParameters
			for _, f := range fmts {
				if f != "" {
					outParams = append(outParams, cdx.MLInputOutputParameters{Format: f})
				}
			}
			card.ModelParameters.Outputs = &outParams
		}
	}
	// Performance metric
	if card.QuantitativeAnalysis == nil || card.QuantitativeAnalysis.PerformanceMetrics == nil || len(*card.QuantitativeAnalysis.PerformanceMetrics) == 0 {
		fmt.Fprint(out, "Performance metric type:value (e.g., accuracy:0.84), blank to skip: ")
		if s := readLine(r); s != "" {
			t, v := splitPair(s, ":")
			perf := []cdx.MLPerformanceMetric{{Type: t, Value: v}}
			card.QuantitativeAnalysis = &cdx.MLQuantitativeAnalysis{PerformanceMetrics: &perf}
		}
	}
	// Users
	if card.Considerations == nil {
		card.Considerations = &cdx.MLModelCardConsiderations{}
	}
	if card.Considerations.Users == nil || len(*card.Considerations.Users) == 0 {
		fmt.Fprint(out, "Intended users (comma-separated): ")
		if s := readLine(r); s != "" {
			users := splitCSV(s)
			card.Considerations.Users = &users
		}
	}
	// Use cases
	if card.Considerations.UseCases == nil || len(*card.Considerations.UseCases) == 0 {
		fmt.Fprint(out, "Use cases (comma-separated): ")
		if s := readLine(r); s != "" {
			usecases := splitCSV(s)
			card.Considerations.UseCases = &usecases
		}
	}
}

func readLine(r *bufio.Reader) string {
	line, _ := r.ReadString('\n')
	return strings.TrimSpace(line)
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		out = append(out, v)
	}
	return out
}

func splitPair(s, sep string) (string, string) {
	parts := strings.SplitN(s, sep, 2)
	if len(parts) < 2 {
		return strings.TrimSpace(parts[0]), ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

func getProperty(comp *cdx.Component, name string) string {
	if comp == nil || comp.Properties == nil {
		return ""
	}
	for _, p := range *comp.Properties {
		if p.Name == name {
			return p.Value
		}
	}
	return ""
}
