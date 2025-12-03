package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/gomlx/go-huggingface/hub"
)

// package-level logger to match project convention
var logWriter io.Writer

// SetLogger sets an optional destination for fetch logs
func SetLogger(w io.Writer) { logWriter = w }

func logf(format string, args ...any) {
	if logWriter != nil {
		fmt.Fprintf(logWriter, format, args...)
	}
}

// HuggingFaceFetcher combines HF API, README.md, and common config files.
type HuggingFaceFetcher struct {
	client   *http.Client
	token    string
	cacheDir string
}

func NewHuggingFaceFetcher(timeout time.Duration, token string, cacheDir string) *HuggingFaceFetcher {
	return &HuggingFaceFetcher{client: &http.Client{Timeout: timeout}, token: token, cacheDir: cacheDir}
}

func (h *HuggingFaceFetcher) Get(id string) (*cdx.MLModelCard, error) {
	logf("[hf] start fetch id=%s\n", id)
	api, err := h.fetchAPI(id)
	if err != nil {
		logf("[hf] api error id=%s err=%v\n", id, err)
		return nil, err
	}
	if api == nil {
		logf("[hf] api nil id=%s\n", id)
		return nil, fmt.Errorf("huggingface api returned nil response for %s", id)
	}
	logf("[hf] api ok id=%s task=%s tags=%d\n", id, api.PipelineTag, len(api.Tags))
	readme, rerr := h.fetchREADME(id)
	if rerr != nil {
		logf("[hf] readme miss id=%s err=%v\n", id, rerr)
	} else {
		logf("[hf] readme ok id=%s len=%d\n", id, len(readme))
	}
	cfg, cerr := h.fetchConfigJSON(id)
	if cerr != nil {
		logf("[hf] config miss id=%s err=%v\n", id, cerr)
	} else {
		logf("[hf] config ok id=%s model_type=%s\n", id, cfg.ModelType)
	}

	card := &cdx.MLModelCard{}

	return card, nil
}

type hfModelResp struct {
	PipelineTag string         `json:"pipeline_tag"`
	Tags        []string       `json:"tags"`
	License     string         `json:"license"`
	CardData    map[string]any `json:"cardData"`
}

func (h *HuggingFaceFetcher) fetchAPI(id string) (*hfModelResp, error) {
	logf("[hf] api req id=%s\n", id)
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
	if resp.StatusCode == http.StatusTooManyRequests {
		logf("[hf] api 429 id=%s\n", id)
		return nil, fmt.Errorf("huggingface rate limit (429) for %s", id)
	}
	if resp.StatusCode != http.StatusOK {
		logf("[hf] api status id=%s code=%d\n", id, resp.StatusCode)
		return nil, fmt.Errorf("huggingface api status %d", resp.StatusCode)
	}
	var m hfModelResp
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}
	logf("[hf] api resp id=%s task=%s tags=%d\n", id, m.PipelineTag, len(m.Tags))
	return &m, nil
}

func (h *HuggingFaceFetcher) fetchREADME(id string) (string, error) {
	logf("[hf] readme req id=%s\n", id)
	repo := hub.New(id).WithAuth(h.token)
	if h.cacheDir != "" {
		repo = repo.WithCacheDir(h.cacheDir)
	}
	path, err := repo.DownloadFile("README.md")
	if err != nil {
		logf("[hf] readme miss id=%s err=%v\n", id, err)
		return "", err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		logf("[hf] readme read err id=%s err=%v\n", id, err)
		return "", err
	}
	logf("[hf] readme resp id=%s len=%d\n", id, len(b))
	return string(b), nil
}

type hfConfigJSON struct {
	ModelType string `json:"model_type"`
	// add fields here as needed
}

func (h *HuggingFaceFetcher) fetchConfigJSON(id string) (*hfConfigJSON, error) {
	logf("[hf] config req id=%s\n", id)
	repo := hub.New(id).WithAuth(h.token)
	if h.cacheDir != "" {
		repo = repo.WithCacheDir(h.cacheDir)
	}
	path, err := repo.DownloadFile("config.json")
	if err != nil {
		logf("[hf] config miss id=%s err=%v\n", id, err)
		return nil, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		logf("[hf] config read err id=%s err=%v\n", id, err)
		return nil, err
	}
	var cfg hfConfigJSON
	if err := json.Unmarshal(b, &cfg); err != nil {
		logf("[hf] config parse err id=%s err=%v\n", id, err)
		return nil, err
	}
	logf("[hf] config resp id=%s model_type=%s\n", id, cfg.ModelType)
	return &cfg, nil
}
