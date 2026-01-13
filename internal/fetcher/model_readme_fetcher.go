package fetcher

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	yaml "go.yaml.in/yaml/v3"
)

// ModelReadmeCard represents metadata extracted from a Hugging Face model README.
//
// Hugging Face model cards usually contain a YAML front matter block (--- ... ---)
// followed by Markdown sections. We parse both:
// - YAML front matter for structured fields (license, tags, datasets, metrics, base_model, model-index)
// - Markdown sections/bullets using regex (e.g. Direct Use, Bias/Risks, Paper/Demo links)
type ModelReadmeCard struct {
	Raw         string
	FrontMatter map[string]any
	Body        string

	// Common front matter fields
	License   string
	Tags      []string
	Datasets  []string
	Metrics   []string
	BaseModel string

	// Extracted from Markdown body (template-based)
	DevelopedBy          string
	PaperURL             string
	DemoURL              string
	DirectUse            string
	OutOfScopeUse        string
	BiasRisksLimitations string
	BiasRecommendations  string
	ModelCardContact     string

	// Environmental Impact (from Markdown body)
	EnvironmentalHardwareType  string
	EnvironmentalHoursUsed     string
	EnvironmentalCloudProvider string
	EnvironmentalComputeRegion string
	EnvironmentalCarbonEmitted string

	// From model-index (if present)
	TaskType string
	TaskName string
	// Metrics with optional values (best-effort)
	ModelIndexMetrics []ModelIndexMetric

	// Quantitative Analysis sections (from Markdown body)
	TestingMetrics string
	Results        string
}

type ModelIndexMetric struct {
	Type  string
	Value string
}

// ModelReadmeFetcher fetches the README.md (model card) for a model repo.
//
// It uses URLs like:
//
//	GET https://huggingface.co/{modelID}/resolve/main/README.md
//
// and falls back to /resolve/master/README.md.
type ModelReadmeFetcher struct {
	Client  *http.Client
	Token   string
	BaseURL string // optional; defaults to "https://huggingface.co"
}

func (f *ModelReadmeFetcher) Fetch(ctx context.Context, modelID string) (*ModelReadmeCard, error) {
	client := f.Client
	if client == nil {
		client = http.DefaultClient
	}

	trimmedModelID := strings.TrimPrefix(strings.TrimSpace(modelID), "/")
	if trimmedModelID == "" {
		return nil, fmt.Errorf("empty model id")
	}

	baseURL := strings.TrimRight(strings.TrimSpace(f.BaseURL), "/")
	if baseURL == "" {
		baseURL = "https://huggingface.co"
	}

	// Try main then master.
	candidates := []string{
		fmt.Sprintf("%s/%s/resolve/main/README.md", baseURL, trimmedModelID),
		fmt.Sprintf("%s/%s/resolve/master/README.md", baseURL, trimmedModelID),
	}

	var lastErr error
	for _, url := range candidates {
		logf(modelID, "GET %s", strings.TrimPrefix(url, baseURL))
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "text/markdown, text/plain, */*")
		if strings.TrimSpace(f.Token) != "" {
			req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(f.Token))
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		bodyBytes, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("huggingface readme status %d", resp.StatusCode)
			continue
		}

		raw := string(bodyBytes)
		card := parseReadmeCard(raw)
		logf(modelID, "ok")

		return card, nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("unable to fetch README")
	}
	logf(modelID, "readme fetch failed (%v)", lastErr)

	return nil, lastErr
}

func parseReadmeCard(raw string) *ModelReadmeCard {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	card := &ModelReadmeCard{Raw: raw}

	fm, body := splitFrontMatter(raw)
	card.FrontMatter = fm
	card.Body = body

	// Front matter fields (best effort)
	card.License = strings.TrimSpace(stringFromAny(fm["license"]))
	card.Tags = stringSliceFromAny(fm["tags"])
	card.Datasets = stringSliceFromAny(fm["datasets"])
	card.Metrics = stringSliceFromAny(fm["metrics"])
	card.BaseModel = strings.TrimSpace(stringFromAny(fm["base_model"]))

	// model-index task + metrics (best effort)
	if mi, ok := fm["model-index"]; ok {
		parseModelIndex(mi, card)
	}

	// Markdown extraction (template-based)
	card.DevelopedBy = strings.TrimSpace(extractBulletValue(body, "Developed by"))
	card.PaperURL = strings.TrimSpace(extractBulletValue(body, "Paper"))
	card.DemoURL = strings.TrimSpace(extractBulletValue(body, "Demo"))
	card.DirectUse = strings.TrimSpace(extractSection(body, "Direct Use"))
	card.OutOfScopeUse = strings.TrimSpace(extractSection(body, "Out-of-Scope Use"))
	card.BiasRisksLimitations = strings.TrimSpace(extractSection(body, "Bias, Risks, and Limitations"))
	card.BiasRecommendations = strings.TrimSpace(extractSection(body, "Recommendations"))
	card.ModelCardContact = strings.TrimSpace(extractSection(body, "Model Card Contact"))

	// Quantitative Analysis sections
	card.TestingMetrics = strings.TrimSpace(extractSection(body, "Metrics"))
	card.Results = strings.TrimSpace(extractSection(body, "Results"))

	// Environmental Impact
	card.EnvironmentalHardwareType = strings.TrimSpace(extractBulletValue(body, "Hardware Type"))
	card.EnvironmentalHoursUsed = strings.TrimSpace(extractBulletValue(body, "Hours used"))
	card.EnvironmentalCloudProvider = strings.TrimSpace(extractBulletValue(body, "Cloud Provider"))
	card.EnvironmentalComputeRegion = strings.TrimSpace(extractBulletValue(body, "Compute Region"))
	card.EnvironmentalCarbonEmitted = strings.TrimSpace(extractBulletValue(body, "Carbon Emitted"))

	// Note: We keep placeholders in the card structure. (for templates/model-card-example)
	// The fieldspecs layer can decide whether to use them or filter them out.

	return card
}

func splitFrontMatter(raw string) (map[string]any, string) {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "---\n") {
		return nil, raw
	}
	// Find the second '---' marker at start of a line.
	// We only treat it as front matter if it starts at the very beginning.
	rest := strings.TrimPrefix(raw, "---\n")
	idx := strings.Index(rest, "\n---\n")
	if idx < 0 {
		// allow file ending marker
		idx = strings.Index(rest, "\n---")
		if idx < 0 {
			return nil, raw
		}
	}

	y := rest[:idx]
	body := strings.TrimSpace(rest[idx:])
	body = strings.TrimPrefix(body, "\n---\n")
	body = strings.TrimPrefix(body, "\n---")
	body = strings.TrimSpace(body)

	m := map[string]any{}
	dec := yaml.NewDecoder(bytes.NewReader([]byte(y)))
	dec.KnownFields(false)
	if err := dec.Decode(&m); err != nil {
		// If YAML parsing fails, still return body; callers can still regex parse.
		return nil, strings.TrimSpace(raw)
	}
	return m, body
}

func stringFromAny(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	default:
		return fmt.Sprint(t)
	}
}

func stringSliceFromAny(v any) []string {
	switch t := v.(type) {
	case nil:
		return nil
	case []any:
		out := make([]string, 0, len(t))
		for _, x := range t {
			s := strings.TrimSpace(stringFromAny(x))
			if s != "" {
				out = append(out, s)
			}
		}
		return normalizeStrings(out)
	case []string:
		return normalizeStrings(t)
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return nil
		}
		return []string{s}
	default:
		s := strings.TrimSpace(fmt.Sprint(t))
		if s == "" {
			return nil
		}
		return []string{s}
	}
}

func normalizeStrings(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func parseModelIndex(mi any, card *ModelReadmeCard) {
	// The model-index is typically a list of entries.
	// We only take the first result for now.
	list, ok := mi.([]any)
	if !ok || len(list) == 0 {
		return
	}
	first, ok := list[0].(map[string]any)
	if !ok {
		return
	}
	resultsAny, ok := first["results"].([]any)
	if !ok || len(resultsAny) == 0 {
		return
	}
	res, ok := resultsAny[0].(map[string]any)
	if !ok {
		return
	}
	// task
	if taskAny, ok := res["task"].(map[string]any); ok {
		card.TaskType = strings.TrimSpace(stringFromAny(taskAny["type"]))
		card.TaskName = strings.TrimSpace(stringFromAny(taskAny["name"]))
	}
	// metrics
	if metricsAny, ok := res["metrics"].([]any); ok {
		out := make([]ModelIndexMetric, 0, len(metricsAny))
		for _, m := range metricsAny {
			mm, ok := m.(map[string]any)
			if !ok {
				continue
			}
			mt := strings.TrimSpace(stringFromAny(mm["type"]))
			mv := strings.TrimSpace(stringFromAny(mm["value"]))
			if mt == "" && mv == "" {
				continue
			}
			out = append(out, ModelIndexMetric{Type: mt, Value: mv})
		}
		if len(out) > 0 {
			card.ModelIndexMetrics = out
		}
	}
}

func extractSection(markdown string, heading string) string {
	markdown = strings.ReplaceAll(markdown, "\r\n", "\n")
	lines := strings.Split(markdown, "\n")

	// Find a level-2 or level-3 heading with the requested text.
	// Example matches:
	//   "## Bias, Risks, and Limitations"
	//   "### Direct Use"
	headingRe := regexp.MustCompile(fmt.Sprintf(`^#{2,3}\s+%s\s*$`, regexp.QuoteMeta(heading)))
	nextHeadingRe := regexp.MustCompile(`^#+\s+.+$`)

	found := false
	buf := make([]string, 0)
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if !found {
			if headingRe.MatchString(line) {
				found = true
			}
			continue
		}
		// Stop at the next heading (any level)
		if nextHeadingRe.MatchString(line) {
			break
		}
		buf = append(buf, line)
	}
	return strings.TrimSpace(strings.Join(buf, "\n"))
}

func extractBulletValue(markdown string, label string) string {
	// Extract values like:
	// - **Paper [optional]:** https://...
	// - **Developed by:** org
	// - **Carbon Emitted** *(additional text)*: 149.2 kg eq. CO2
	// Supports optional bracketed qualifiers in the label part and text between the label and colon.
	// Pattern handles both: **Label:** (colon inside) and **Label** text: (colon outside)
	pat := fmt.Sprintf(`(?m)^-\s+\*\*%s(?:\s*\[[^\]]+\])?(?::\*\*|\*\*[^:\n]*:)\s*(.+?)\s*$`, regexp.QuoteMeta(label))
	re := regexp.MustCompile(pat)
	m := re.FindStringSubmatch(markdown)
	if len(m) < 2 {
		return ""
	}
	return strings.TrimSpace(m[1])
}

func isPlaceholder(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return true
	}
	// Common placeholder from the template.
	if strings.Contains(s, "[More Information Needed]") {
		return true
	}
	return false
}
