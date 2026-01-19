package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// represents an AI-related artifact detected in a project.
type Discovery struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Path     string `json:"path"`
	Evidence string `json:"evidence"`
}

// Weight file detection intentionally disabled for now; focus on HF usage.
// var weightExtensions = map[string]struct{}{
//     ".pt":          {},
//     ".pth":         {},
//     ".bin":         {},
//     ".safetensors": {},
//     ".ckpt":        {},
//     ".onnx":        {},
//     ".tflite":      {},
// }

// hfModelPattern matches Hugging Face model IDs inside from_pretrained("...") calls.
// Supports both single segment IDs (e.g., bert-base-uncased) and org/model forms (e.g., facebook/opt-1.3b).
var hfModelPattern = regexp.MustCompile(`from_pretrained\("([A-Za-z0-9_.-]+(?:/[A-Za-z0-9_.-]+)?)"\)`)

// Scan walks the target path and returns detected AI components.
func Scan(root string) ([]Discovery, error) {
	var results []Discovery

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(info.Name()))

		// Weight file detection disabled.

		// Lightweight content scan for HF model IDs.
		if shouldScanForModelID(ext) {
			f, openErr := os.Open(path)
			if openErr != nil {
				return nil // skip file read errors silently
			}
			defer f.Close()
			scanner := bufio.NewScanner(f)
			lineNum := 0
			for scanner.Scan() {
				lineNum++
				line := scanner.Text()

				matches := hfModelPattern.FindAllStringSubmatch(line, -1)
				for _, m := range matches {
					if len(m) > 1 {
						modelID := m[1]
						evidence := "from_pretrained() pattern at line " +
							strconv.Itoa(lineNum) + ": " + line
						results = append(results, Discovery{
							ID:       modelID,
							Name:     modelID,
							Type:     "model",
							Path:     path,
							Evidence: evidence,
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
	deduped := dedupe(results)
	return deduped, nil
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
func dedupe(components []Discovery) []Discovery {
	index := make(map[string]Discovery)
	for _, c := range components {
		key := c.Type + "::" + c.ID
		if existing, ok := index[key]; ok {
			if !strings.Contains(existing.Evidence, c.Evidence) {
				existing.Evidence += ". " + c.Evidence
			}
			index[key] = existing
		} else {
			index[key] = c
		}
	}
	out := make([]Discovery, 0, len(index))
	for _, v := range index {
		out = append(out, v)
	}
	return out
}
