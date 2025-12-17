package builder

import (
	"aibomgen-cra/internal/metadata"
	"aibomgen-cra/internal/scanner"
)

type BuildContext struct {
	ModelID string
	Scan    scanner.Discovery
	Meta    metadata.View
}

type Options struct {
	IncludeEvidenceProperties bool
	HuggingFaceBaseURL        string
}

func DefaultOptions() Options {
	return Options{IncludeEvidenceProperties: true, HuggingFaceBaseURL: "https://huggingface.co/"}
}
