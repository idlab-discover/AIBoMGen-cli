package parameters

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

// HuggingFaceParametersFetcher retrieves model parameters from HF Hub.
type HuggingFaceParametersFetcher struct {
	client *http.Client
	token  string
	logger io.Writer
}

func NewHuggingFaceParametersFetcher(timeout time.Duration, token string) *HuggingFaceParametersFetcher {
	return &HuggingFaceParametersFetcher{
		client: &http.Client{Timeout: timeout},
		token:  token,
	}
}

type hfModelResp struct {
	PipelineTag string   `json:"pipeline_tag"`
	Tags        []string `json:"tags"`
	License     string   `json:"license"`
}

func (h *HuggingFaceParametersFetcher) Get(id string) (*cdx.MLModelParameters, error) {
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
	mp, _ := MapHFToModelParameters(&m)
	if h.logger != nil && mp != nil {
		inN, outN, dsN := 0, 0, 0
		if mp.Inputs != nil {
			inN = len(*mp.Inputs)
		}
		if mp.Outputs != nil {
			outN = len(*mp.Outputs)
		}
		if mp.Datasets != nil {
			dsN = len(*mp.Datasets)
		}
		fmt.Fprintf(h.logger, "HF parameters: %s -> task=%s, datasets=%d, io:%d/%d\n", id, mp.Task, dsN, inN, outN)
	}
	return mp, nil
}

// MapHFToModelParameters exports mapping for testing.
func MapHFToModelParameters(m *hfModelResp) (*cdx.MLModelParameters, []string) {
	if m == nil {
		return nil, nil
	}
	var datasets []cdx.MLDatasetChoice
	var datasetRefs []string
	for _, t := range m.Tags {
		if len(t) > 8 && t[:8] == "dataset:" {
			datasets = append(datasets, cdx.MLDatasetChoice{Ref: t})
			datasetRefs = append(datasetRefs, t)
		}
	}
	inputs, outputs := ioForTask(m.PipelineTag)
	mp := &cdx.MLModelParameters{
		Task:     m.PipelineTag,
		Datasets: &datasets,
		Inputs:   &inputs,
		Outputs:  &outputs,
	}
	return mp, datasetRefs
}

func ioForTask(task string) ([]cdx.MLInputOutputParameters, []cdx.MLInputOutputParameters) {
	in := []cdx.MLInputOutputParameters{{Format: "application/octet-stream"}}
	out := []cdx.MLInputOutputParameters{{Format: "application/octet-stream"}}
	switch task {
	case "text-classification", "sentiment-analysis", "token-classification", "question-answering", "translation", "fill-mask":
		in = []cdx.MLInputOutputParameters{{Format: "text/plain"}}
		out = []cdx.MLInputOutputParameters{{Format: "classification-label"}}
	case "image-classification", "object-detection", "image-segmentation":
		in = []cdx.MLInputOutputParameters{{Format: "image/*"}}
		out = []cdx.MLInputOutputParameters{{Format: "classification-label"}}
	case "audio-classification", "automatic-speech-recognition":
		in = []cdx.MLInputOutputParameters{{Format: "audio/*"}}
		out = []cdx.MLInputOutputParameters{{Format: "text/plain"}}
	case "text-generation", "text2text-generation":
		in = []cdx.MLInputOutputParameters{{Format: "text/plain"}}
		out = []cdx.MLInputOutputParameters{{Format: "text/plain"}}
	}
	return in, out
}

// SetLogger sets an optional writer for user-facing logs.
func (h *HuggingFaceParametersFetcher) SetLogger(w io.Writer) { h.logger = w }
