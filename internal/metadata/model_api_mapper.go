package metadata

import (
	"strings"

	"aibomgen-cra/internal/fetcher"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

func MapModelAPIToMetadata(modelID string, r *fetcher.ModelAPIResponse, store *Store) {
	if store == nil || strings.TrimSpace(modelID) == "" || r == nil {
		return
	}

	modelID = strings.TrimSpace(modelID)
	logf(modelID, "map start")
	put := func(key Key, value any) {
		if key == "" {
			return
		}
		if value == nil {
			return
		}
		logf(modelID, "map put key=%s value=%s", key.String(), summarizeValue(value))
		store.Put(modelID, key, value)
	}

	// component.name
	// first try: response.ID
	// second try: response.ModelID
	// third try: use modelID param
	name := strings.TrimSpace(r.ID)
	if name == "" {
		name = strings.TrimSpace(r.ModelID)
	}
	if name == "" {
		name = strings.TrimSpace(modelID)
	}
	// set component.name
	put(ComponentName, name)

	// manufacturer/author/group (currently author maps all three)
	author := strings.TrimSpace(r.Author)
	if author != "" {
		put(ComponentManufacturer, author)
		put(ComponentAuthor, author)
		put(ComponentGroup, author)
	}

	// tags
	if len(r.Tags) > 0 {
		put(ComponentTags, normalizeStrings(r.Tags))
	}
	// licenses (prefer cardData.license; else tag license:*)
	if lic := extractLicense(r.CardData, r.Tags); lic != "" {
		put(ComponentLicenses, []string{lic})
	}

	// hashes (Hub sha is git SHA-1)
	if strings.TrimSpace(r.SHA) != "" {
		put(ComponentHashes, []cdx.Hash{
			{Algorithm: cdx.HashAlgoSHA1, Value: strings.TrimSpace(r.SHA)},
		})
	}

	// external ref (always provide HF website)
	put(ComponentExternalReferences, []cdx.ExternalReference{{
		Type: cdx.ExternalReferenceType("website"),
		URL:  "https://huggingface.co/" + strings.TrimPrefix(modelID, "/"),
	}})

	// modelcard parameters
	if strings.TrimSpace(r.PipelineTag) != "" {
		put(ModelCardModelParametersTask, strings.TrimSpace(r.PipelineTag))
	}
	if strings.TrimSpace(r.Config.ModelType) != "" {
		put(ModelCardModelParametersArchitectureFamily, strings.TrimSpace(r.Config.ModelType))
	}
	if len(r.Config.Architectures) > 0 && strings.TrimSpace(r.Config.Architectures[0]) != "" {
		put(ModelCardModelParametersModelArchitecture, strings.TrimSpace(r.Config.Architectures[0]))
	}

	// datasets (cardData.datasets + tag dataset:*)
	if ds := extractDatasets(r.CardData, r.Tags); len(ds) > 0 {
		put(ModelCardModelParametersDatasets, ds)
	}

	// HF "properties" (stored as primitives; builders can turn into CycloneDX Component.Properties)
	if strings.TrimSpace(r.LastMod) != "" {
		put(ComponentPropertiesHuggingFaceLastModified, strings.TrimSpace(r.LastMod))
	}
	if strings.TrimSpace(r.CreatedAt) != "" {
		put(ComponentPropertiesHuggingFaceCreatedAt, strings.TrimSpace(r.CreatedAt))
	}
	if strings.TrimSpace(r.LibraryName) != "" {
		put(ComponentPropertiesHuggingFaceLibraryName, strings.TrimSpace(r.LibraryName))
	}
	if r.Downloads > 0 {
		put(ComponentPropertiesHuggingFaceDownloads, r.Downloads)
	}
	if r.Likes > 0 {
		put(ComponentPropertiesHuggingFaceLikes, r.Likes)
	}
	if r.UsedStorage > 0 {
		put(ComponentPropertiesHuggingFaceUsedStorage, r.UsedStorage)
	}
	put(ComponentPropertiesHuggingFacePrivate, r.Private)

	if lang := extractLanguage(r.CardData); lang != "" {
		put(ComponentPropertiesHuggingFaceLanguage, lang)
	}
	logf(modelID, "map done")
}

// extractLicense extracts license from cardData.license or tag license:*.
func extractLicense(cardData map[string]any, tags []string) string {
	// cardData.license
	if cardData != nil {
		if v, ok := cardData["license"]; ok {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}
	// tag license:apache-2.0
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if strings.HasPrefix(t, "license:") {
			return strings.TrimSpace(strings.TrimPrefix(t, "license:"))
		}
	}
	return ""
}

// extractLanguage extracts language from cardData.language.
func extractLanguage(cardData map[string]any) string {
	if cardData == nil {
		return ""
	}
	v, ok := cardData["language"]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case []any:
		var out []string
		for _, it := range t {
			if s, ok := it.(string); ok {
				s = strings.TrimSpace(s)
				if s != "" {
					out = append(out, s)
				}
			}
		}
		return strings.Join(out, ",")
	default:
		return ""
	}
}

// extractDatasets extracts datasets from cardData.datasets and tag dataset:*.
func extractDatasets(cardData map[string]any, tags []string) []string {
	seen := map[string]struct{}{}
	var out []string

	add := func(raw string) {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return
		}
		if !strings.Contains(raw, ":") {
			raw = "dataset:" + raw
		}
		if _, ok := seen[raw]; ok {
			return
		}
		seen[raw] = struct{}{}
		out = append(out, raw)
	}

	// cardData.datasets: string or array
	if cardData != nil {
		if v, ok := cardData["datasets"]; ok && v != nil {
			switch t := v.(type) {
			case string:
				add(t)
			case []any:
				for _, it := range t {
					if s, ok := it.(string); ok {
						add(s)
					}
				}
			}
		}
	}

	// tags: dataset:NAME
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if strings.HasPrefix(t, "dataset:") {
			add(t)
		}
	}

	return out
}

func normalizeStrings(in []string) []string {
	out := make([]string, 0, len(in))
	seen := map[string]struct{}{}
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
