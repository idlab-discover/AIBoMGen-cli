package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Component represents an AI-related artifact detected in a project.
type Component struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	Path       string  `json:"path"`
	Evidence   string  `json:"evidence"`
	Confidence float64 `json:"confidence"`
}

var weightExtensions = map[string]struct{}{
	".pt":          {},
	".pth":         {},
	".bin":         {},
	".safetensors": {},
	".ckpt":        {},
	".onnx":        {},
	".tflite":      {},
}

// hfModelPattern matches Hugging Face model IDs inside from_pretrained("...") calls.
// Supports both single segment IDs (e.g., bert-base-uncased) and org/model forms (e.g., facebook/opt-1.3b).
var hfModelPattern = regexp.MustCompile(`from_pretrained\("([A-Za-z0-9_.-]+(?:/[A-Za-z0-9_.-]+)?)"\)`)

// Scan walks the target path and returns detected AI components.
func Scan(root string) ([]Component, error) {
	var results []Component

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(info.Name()))
		if _, ok := weightExtensions[ext]; ok {
			results = append(results, Component{
				ID:         path,
				Name:       info.Name(),
				Type:       "weight-file",
				Path:       path,
				Evidence:   "extension:" + ext,
				Confidence: 0.9,
			})
		}

		// Lightweight content scan for HF model IDs.
		if shouldScanForModelID(ext) {
			f, openErr := os.Open(path)
			if openErr != nil {
				return nil // skip file read errors silently
			}
			defer f.Close()
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := scanner.Text()
				matches := hfModelPattern.FindAllStringSubmatch(line, -1)
				for _, m := range matches {
					if len(m) > 1 {
						modelID := m[1]
						results = append(results, Component{
							ID:         modelID,
							Name:       modelID,
							Type:       "model",
							Path:       path,
							Evidence:   "from_pretrained() pattern",
							Confidence: 0.95,
						})
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dedupe(results), nil
}

func shouldScanForModelID(ext string) bool {
	switch ext {
	case ".py", ".ipynb", ".txt":
		return true
	default:
		return false
	}
}

// dedupe merges components with identical ID+Type.
func dedupe(components []Component) []Component {
	index := make(map[string]Component)
	for _, c := range components {
		key := c.Type + "::" + c.ID
		if existing, ok := index[key]; ok {
			// Prefer highest confidence & concatenate evidence.
			if c.Confidence > existing.Confidence {
				existing.Confidence = c.Confidence
			}
			if !strings.Contains(existing.Evidence, c.Evidence) {
				existing.Evidence += ";" + c.Evidence
			}
			index[key] = existing
		} else {
			index[key] = c
		}
	}
	out := make([]Component, 0, len(index))
	for _, v := range index {
		out = append(out, v)
	}
	return out
}
