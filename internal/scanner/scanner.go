package scanner

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"aibomgen-cra/internal/ui"
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

var logWriter io.Writer

// SetLogger sets an optional destination for scan logs.
func SetLogger(w io.Writer) { logWriter = w }

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
						if logWriter != nil {
							prefix := ui.Color("Scan:", ui.FgYellow)
							fmt.Fprintf(logWriter, "%s found model '%s' at %s:%d\n", prefix, modelID, path, lineNum)
						}
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
	if logWriter != nil {
		// Count models detected
		modelCount := 0
		for _, c := range deduped {
			if c.Type == "model" {
				modelCount++
			}
		}
		prefix := ui.Color("Scan:", ui.FgYellow)
		fmt.Fprintf(logWriter, "%s detected %d components (models: %d)\n", prefix, len(deduped), modelCount)
	}
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
