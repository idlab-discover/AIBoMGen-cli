package quantitative

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

type HuggingFaceQuantFetcher struct {
	client *http.Client
	token  string
	logger io.Writer
}

func NewHuggingFaceQuantFetcher(timeout time.Duration, token string) *HuggingFaceQuantFetcher {
	return &HuggingFaceQuantFetcher{client: &http.Client{Timeout: timeout}, token: token}
}

// Partial response focusing on cardData->model-index
type hfModelResp struct {
	CardData map[string]any `json:"cardData"`
}

func (h *HuggingFaceQuantFetcher) Get(id string) (*cdx.MLQuantitativeAnalysis, error) {
	url := fmt.Sprintf("https://huggingface.co/api/models/%s", id)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if h.token != "" {
		req.Header.Set("Authorization", "Bearer "+h.token)
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("huggingface api status %d", resp.StatusCode)
	}
	var m hfModelResp
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}
	qa := mapHFToQuant(&m)
	if h.logger != nil {
		count := 0
		if qa != nil && qa.PerformanceMetrics != nil {
			count = len(*qa.PerformanceMetrics)
		}
		fmt.Fprintf(h.logger, "HF quantitative: %s -> metrics=%d\n", id, count)
	}
	return qa, nil
}

func mapHFToQuant(m *hfModelResp) *cdx.MLQuantitativeAnalysis {
	if m == nil || m.CardData == nil {
		return nil
	}
	mi, ok := m.CardData["model-index"].([]any)
	if !ok || len(mi) == 0 {
		return nil
	}
	// Try to extract first results[].metrics[] entries
	var metrics []cdx.MLPerformanceMetric
	for _, entry := range mi {
		obj, _ := entry.(map[string]any)
		if obj == nil {
			continue
		}
		results, _ := obj["results"].([]any)
		for _, r := range results {
			rm, _ := r.(map[string]any)
			if rm == nil {
				continue
			}
			ms, _ := rm["metrics"].([]any)
			for _, mm := range ms {
				mObj, _ := mm.(map[string]any)
				if mObj == nil {
					continue
				}
				// Try common keys
				t, _ := mObj["type"].(string)
				if t == "" {
					if n, _ := mObj["name"].(string); n != "" {
						t = n
					}
				}
				var v string
				switch vv := mObj["value"].(type) {
				case float64:
					v = fmt.Sprintf("%.4f", vv)
				case string:
					v = vv
				}
				if t != "" && v != "" {
					metrics = append(metrics, cdx.MLPerformanceMetric{Type: t, Value: v})
				}
			}
		}
	}
	if len(metrics) == 0 {
		return nil
	}
	return &cdx.MLQuantitativeAnalysis{PerformanceMetrics: &metrics}
}

// SetLogger sets an optional writer for user-facing logs.
func (h *HuggingFaceQuantFetcher) SetLogger(w io.Writer) { h.logger = w }
