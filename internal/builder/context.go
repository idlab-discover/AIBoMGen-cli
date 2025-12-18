package builder

import (
	"aibomgen-cra/internal/fetcher"
	"aibomgen-cra/internal/scanner"
)

type BuildContext struct {
	ModelID string
	Scan    scanner.Discovery
	HF      *fetcher.ModelAPIResponse
}

type Options struct {
	IncludeEvidenceProperties bool
	HuggingFaceBaseURL        string
}

func DefaultOptions() Options {
	return Options{
		IncludeEvidenceProperties: true,
		HuggingFaceBaseURL:        "https://huggingface.co/",
	}
}
