package considerations

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

type HuggingFaceConsiderationsFetcher struct {
	client *http.Client
	token  string
	logger io.Writer
}

func NewHuggingFaceConsiderationsFetcher(timeout time.Duration, token string) *HuggingFaceConsiderationsFetcher {
	return &HuggingFaceConsiderationsFetcher{client: &http.Client{Timeout: timeout}, token: token}
}

type hfModelResp struct {
	CardData map[string]any `json:"cardData"`
}

func (h *HuggingFaceConsiderationsFetcher) Get(id string) (*cdx.MLModelCardConsiderations, error) {
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
	cc := mapHFToConsiderations(&m)
	if h.logger != nil {
		uN, ucN, lN, eN := 0, 0, 0, 0
		if cc != nil {
			if cc.Users != nil {
				uN = len(*cc.Users)
			}
			if cc.UseCases != nil {
				ucN = len(*cc.UseCases)
			}
			if cc.TechnicalLimitations != nil {
				lN = len(*cc.TechnicalLimitations)
			}
			if cc.EthicalConsiderations != nil {
				eN = len(*cc.EthicalConsiderations)
			}
		}
		fmt.Fprintf(h.logger, "HF considerations: %s -> users=%d,usecases=%d,limits=%d,ethics=%d\n", id, uN, ucN, lN, eN)
	}
	return cc, nil
}

func mapHFToConsiderations(m *hfModelResp) *cdx.MLModelCardConsiderations {
	if m == nil || m.CardData == nil {
		return nil
	}
	cons := &cdx.MLModelCardConsiderations{}
	// Attempt to extract common cardData keys
	if v, ok := m.CardData["intended_use"]; ok {
		if s := toStringSlice(v); len(s) > 0 {
			cons.UseCases = &s
		}
	}
	if v, ok := m.CardData["intended-uses"]; ok { // alternative key
		if s := toStringSlice(v); len(s) > 0 {
			cons.UseCases = &s
		}
	}
	if v, ok := m.CardData["users"]; ok {
		if s := toStringSlice(v); len(s) > 0 {
			cons.Users = &s
		}
	}
	if v, ok := m.CardData["limitations"]; ok {
		if s := toStringSlice(v); len(s) > 0 {
			cons.TechnicalLimitations = &s
		}
	}
	if v, ok := m.CardData["ethical_considerations"]; ok {
		if s := toStringSlice(v); len(s) > 0 {
			ecs := make([]cdx.MLModelCardEthicalConsideration, 0, len(s))
			for _, name := range s {
				ecs = append(ecs, cdx.MLModelCardEthicalConsideration{Name: name})
			}
			cons.EthicalConsiderations = &ecs
		}
	}
	if cons.Users == nil && cons.UseCases == nil && cons.TechnicalLimitations == nil && cons.EthicalConsiderations == nil {
		return nil
	}
	return cons
}

func toStringSlice(v any) []string {
	switch t := v.(type) {
	case string:
		if t == "" {
			return nil
		}
		return []string{t}
	case []any:
		out := make([]string, 0, len(t))
		for _, e := range t {
			if s, ok := e.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

// SetLogger sets an optional writer for user-facing logs.
func (h *HuggingFaceConsiderationsFetcher) SetLogger(w io.Writer) { h.logger = w }
