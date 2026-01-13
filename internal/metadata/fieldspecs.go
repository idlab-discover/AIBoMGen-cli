package metadata

import (
	"fmt"
	"strings"

	"github.com/idlab-discover/AIBoMGen-cli/internal/fetcher"
	"github.com/idlab-discover/AIBoMGen-cli/internal/scanner"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

// Key identifies a CycloneDX field (or pseudo-field) we want to populate/check.
type Key string

func (k Key) String() string { return string(k) }

const (
	// BOM.metadata.component.* (MODEL)
	ComponentName               Key = "BOM.metadata.component.name"
	ComponentExternalReferences Key = "BOM.metadata.component.externalReferences"
	ComponentTags               Key = "BOM.metadata.component.tags"
	ComponentLicenses           Key = "BOM.metadata.component.licenses"
	ComponentHashes             Key = "BOM.metadata.component.hashes"
	ComponentManufacturer       Key = "BOM.metadata.component.manufacturer"
	ComponentGroup              Key = "BOM.metadata.component.group"

	// Component-level extra properties (stored later as CycloneDX Component.Properties)
	ComponentPropertiesHuggingFaceLastModified Key = "BOM.metadata.component.properties.huggingface:lastModified"
	ComponentPropertiesHuggingFaceCreatedAt    Key = "BOM.metadata.component.properties.huggingface:createdAt"
	ComponentPropertiesHuggingFaceLanguage     Key = "BOM.metadata.component.properties.huggingface:language"
	ComponentPropertiesHuggingFaceUsedStorage  Key = "BOM.metadata.component.properties.huggingface:usedStorage"
	ComponentPropertiesHuggingFacePrivate      Key = "BOM.metadata.component.properties.huggingface:private"
	ComponentPropertiesHuggingFaceLibraryName  Key = "BOM.metadata.component.properties.huggingface:libraryName"
	ComponentPropertiesHuggingFaceDownloads    Key = "BOM.metadata.component.properties.huggingface:downloads"
	ComponentPropertiesHuggingFaceLikes        Key = "BOM.metadata.component.properties.huggingface:likes"
	ComponentPropertiesModelCardBaseModel      Key = "BOM.metadata.component.properties.huggingface:baseModel"
	ComponentPropertiesModelCardContact        Key = "BOM.metadata.component.properties.huggingface:modelCardContact"

	// BOM.metadata.component.modelCard.* (MODEL CARD)
	ModelCardModelParametersTask                                 Key = "BOM.metadata.component.modelCard.modelParameters.task"
	ModelCardModelParametersArchitectureFamily                   Key = "BOM.metadata.component.modelCard.modelParameters.architectureFamily"
	ModelCardModelParametersModelArchitecture                    Key = "BOM.metadata.component.modelCard.modelParameters.modelArchitecture"
	ModelCardModelParametersDatasets                             Key = "BOM.metadata.component.modelCard.modelParameters.datasets"
	ModelCardConsiderationsUseCases                              Key = "BOM.metadata.component.modelCard.considerations.useCases"
	ModelCardConsiderationsTechnicalLimitations                  Key = "BOM.metadata.component.modelCard.considerations.technicalLimitations"
	ModelCardConsiderationsEthicalConsiderations                 Key = "BOM.metadata.component.modelCard.considerations.ethicalConsiderations"
	ModelCardQuantitativeAnalysisPerformanceMetrics              Key = "BOM.metadata.component.modelCard.quantitativeAnalysis.performanceMetrics"
	ModelCardConsiderationsEnvironmentalConsiderationsProperties Key = "BOM.metadata.component.modelCard.considerations.environmentalConsiderations.properties"
)

// Source is everything FieldSpecs can read from.
type Source struct {
	ModelID string
	Scan    scanner.Discovery
	HF      *fetcher.ModelAPIResponse
	Readme  *fetcher.ModelReadmeCard
}

// Target is everything FieldSpecs are allowed to mutate.
type Target struct {
	BOM       *cdx.BOM
	Component *cdx.Component
	ModelCard *cdx.MLModelCard

	// Options (builder can set these when calling Apply)
	IncludeEvidenceProperties bool
	HuggingFaceBaseURL        string
}

// FieldSpec is a first-class definition of a field:
// - how it contributes to completeness
// - how it is populated into the BOM
// - how its presence is detected
type FieldSpec struct {
	Key      Key
	Weight   float64
	Required bool

	Apply   func(src Source, tgt Target)
	Present func(b *cdx.BOM) bool
}

// This is the central registry of all known FieldSpecs.
// Each spec defines how to apply itself and how to check presence.
// The registry is used by the BOM builder and completeness checker.
// It is the single source of truth for what fields we care about.
func Registry() []FieldSpec {
	return []FieldSpec{
		{
			Key:      ComponentName,
			Weight:   1.0,
			Required: true,
			Apply: func(src Source, tgt Target) {
				if tgt.Component == nil {
					return
				}
				name := strings.TrimSpace(src.ModelID)
				// prefer HF ID if available
				// consider also scan name
				// finally fall back to model ID
				if src.HF != nil {
					if s := strings.TrimSpace(src.HF.ID); s != "" {
						name = s
					} else if s := strings.TrimSpace(src.HF.ModelID); s != "" {
						name = s
					}
				}
				if strings.TrimSpace(src.Scan.Name) != "" {
					name = strings.TrimSpace(src.Scan.Name)
				}
				if name != "" {
					tgt.Component.Name = name
					logf(src.ModelID, "apply %s set=%s", ComponentName, summarizeValue(name))
				}
			},
			Present: func(b *cdx.BOM) bool {
				ok := bomHasComponentName(b)
				mid := ""
				if c := bomComponent(b); c != nil {
					mid = c.Name
				}
				logf(mid, "present %s ok=%t", ComponentName, ok)
				return ok
			},
		},
		{
			Key:      ComponentExternalReferences,
			Weight:   0.5,
			Required: false,
			Apply: func(src Source, tgt Target) {
				if tgt.Component == nil {
					return
				}
				// Always provide HF website fallback if we know modelID
				modelID := strings.TrimSpace(src.ModelID)
				if modelID == "" {
					return
				}
				base := strings.TrimSpace(tgt.HuggingFaceBaseURL)
				if base == "" {
					base = "https://huggingface.co/"
				}
				if !strings.HasSuffix(base, "/") {
					base += "/"
				}

				url := base + strings.TrimPrefix(modelID, "/")
				refs := []cdx.ExternalReference{{
					Type: cdx.ExternalReferenceType("website"),
					URL:  url,
				}}

				// Add Paper URL if available
				if src.Readme != nil && strings.TrimSpace(src.Readme.PaperURL) != "" {
					refs = append(refs, cdx.ExternalReference{
						Type: cdx.ExternalReferenceType("documentation"),
						URL:  strings.TrimSpace(src.Readme.PaperURL),
					})
				}

				// Add Demo URL if available
				if src.Readme != nil && strings.TrimSpace(src.Readme.DemoURL) != "" {
					refs = append(refs, cdx.ExternalReference{
						Type: cdx.ExternalReferenceType("other"),
						URL:  strings.TrimSpace(src.Readme.DemoURL),
					})
				}

				tgt.Component.ExternalReferences = &refs
				logf(src.ModelID, "apply %s set=%s", ComponentExternalReferences, summarizeValue(refs))
			},
			Present: func(b *cdx.BOM) bool {
				c := bomComponent(b)
				ok := c != nil && c.ExternalReferences != nil && len(*c.ExternalReferences) > 0
				mid := ""
				if c != nil {
					mid = c.Name
				}
				logf(mid, "present %s ok=%t", ComponentExternalReferences, ok)
				return ok
			},
		},
		{
			Key:      ComponentTags,
			Weight:   0.5,
			Required: false,
			Apply: func(src Source, tgt Target) {
				if tgt.Component == nil {
					return
				}
				if tgt.Component.Tags != nil && len(*tgt.Component.Tags) > 0 {
					return
				}
				var tags []string
				if src.HF != nil && len(src.HF.Tags) > 0 {
					tags = normalizeStrings(src.HF.Tags)
				} else if src.Readme != nil && len(src.Readme.Tags) > 0 {
					tags = normalizeStrings(src.Readme.Tags)
				}
				if len(tags) == 0 {
					return
				}
				tgt.Component.Tags = &tags
				logf(src.ModelID, "apply %s set=%s", ComponentTags, summarizeValue(tags))
			},
			Present: func(b *cdx.BOM) bool {
				c := bomComponent(b)
				ok := c != nil && c.Tags != nil && len(*c.Tags) > 0
				mid := ""
				if c != nil {
					mid = c.Name
				}
				logf(mid, "present %s ok=%t", ComponentTags, ok)
				return ok
			},
		},
		{
			Key:      ComponentLicenses,
			Weight:   1.0,
			Required: false,
			Apply: func(src Source, tgt Target) {
				if tgt.Component == nil {
					return
				}
				if tgt.Component.Licenses != nil && len(*tgt.Component.Licenses) > 0 {
					return
				}
				lic := ""
				if src.HF != nil {
					lic = extractLicense(src.HF.CardData, src.HF.Tags)
				}
				if lic == "" && src.Readme != nil {
					lic = strings.TrimSpace(src.Readme.License)
				}
				if lic == "" {
					return
				}
				ls := cdx.Licenses{
					{License: &cdx.License{Name: lic}},
				}
				tgt.Component.Licenses = &ls
				logf(src.ModelID, "apply %s set=%s", ComponentLicenses, summarizeValue(lic))
			},
			Present: func(b *cdx.BOM) bool {
				c := bomComponent(b)
				ok := c != nil && c.Licenses != nil && len(*c.Licenses) > 0
				mid := ""
				if c != nil {
					mid = c.Name
				}
				logf(mid, "present %s ok=%t", ComponentLicenses, ok)
				return ok
			},
		},
		{
			Key:      ComponentHashes,
			Weight:   1.0,
			Required: false,
			Apply: func(src Source, tgt Target) {
				if tgt.Component == nil || src.HF == nil {
					return
				}
				sha := strings.TrimSpace(src.HF.SHA)
				if sha == "" {
					return
				}
				hs := []cdx.Hash{{Algorithm: cdx.HashAlgoSHA1, Value: sha}}
				tgt.Component.Hashes = &hs
				logf(src.ModelID, "apply %s set=%s", ComponentHashes, summarizeValue(sha))
			},
			Present: func(b *cdx.BOM) bool {
				c := bomComponent(b)
				ok := c != nil && c.Hashes != nil && len(*c.Hashes) > 0
				mid := ""
				if c != nil {
					mid = c.Name
				}
				logf(mid, "present %s ok=%t", ComponentHashes, ok)
				return ok
			},
		},
		{
			Key:      ComponentManufacturer,
			Weight:   0.5,
			Required: false,
			Apply: func(src Source, tgt Target) {
				if tgt.Component == nil {
					return
				}
				if tgt.Component.Manufacturer != nil && strings.TrimSpace(tgt.Component.Manufacturer.Name) != "" {
					return
				}
				s := ""
				if src.HF != nil {
					s = strings.TrimSpace(src.HF.Author)
				}
				if s == "" && src.Readme != nil {
					s = strings.TrimSpace(src.Readme.DevelopedBy)
				}
				if s == "" {
					return
				}
				tgt.Component.Manufacturer = &cdx.OrganizationalEntity{Name: s}
				logf(src.ModelID, "apply %s set=%s", ComponentManufacturer, summarizeValue(s))
			},
			Present: func(b *cdx.BOM) bool {
				c := bomComponent(b)
				ok := c != nil && c.Manufacturer != nil && strings.TrimSpace(c.Manufacturer.Name) != ""
				mid := ""
				if c != nil {
					mid = c.Name
				}
				logf(mid, "present %s ok=%t", ComponentManufacturer, ok)
				return ok
			},
		},
		{
			Key:      ComponentGroup,
			Weight:   0.25,
			Required: false,
			Apply: func(src Source, tgt Target) {
				if tgt.Component == nil {
					return
				}
				if strings.TrimSpace(tgt.Component.Group) != "" {
					return
				}
				s := ""
				if src.HF != nil {
					s = strings.TrimSpace(src.HF.Author)
				}
				if s == "" && src.Readme != nil {
					s = strings.TrimSpace(src.Readme.DevelopedBy)
				}
				if s == "" {
					return
				}
				tgt.Component.Group = s
				logf(src.ModelID, "apply %s set=%s", ComponentGroup, summarizeValue(s))
			},
			Present: func(b *cdx.BOM) bool {
				c := bomComponent(b)
				ok := c != nil && strings.TrimSpace(c.Group) != ""
				mid := ""
				if c != nil {
					mid = c.Name
				}
				logf(mid, "present %s ok=%t", ComponentGroup, ok)
				return ok
			},
		},
		// Evidence properties
		{
			Key:      Key("aibomgen.evidence"),
			Weight:   0,
			Required: false,
			Apply: func(src Source, tgt Target) {
				if tgt.Component == nil || !tgt.IncludeEvidenceProperties {
					return
				}
				setProperty(tgt.Component, "aibomgen.type", src.Scan.Type)
				setProperty(tgt.Component, "aibomgen.evidence", src.Scan.Evidence)
				setProperty(tgt.Component, "aibomgen.path", src.Scan.Path)
				logf(src.ModelID, "apply aibomgen.evidence type=%s path=%s evidence=%s", summarizeValue(src.Scan.Type), summarizeValue(src.Scan.Path), summarizeValue(src.Scan.Evidence))
			},
			Present: func(b *cdx.BOM) bool {
				mid := ""
				if c := bomComponent(b); c != nil {
					mid = c.Name
				}
				logf(mid, "present aibomgen.evidence ok=true")
				return true
			},
		},
		// Component.Properties (Property.Name = key without the BOM.* prefix)
		hfProp(ComponentPropertiesHuggingFaceLastModified, 0.2, func(src Source) (any, bool) {
			r := src.HF
			if r == nil {
				return nil, false
			}
			s := strings.TrimSpace(r.LastMod)
			return s, s != ""
		}),
		hfProp(ComponentPropertiesHuggingFaceCreatedAt, 0.2, func(src Source) (any, bool) {
			r := src.HF
			if r == nil {
				return nil, false
			}
			s := strings.TrimSpace(r.CreatedAt)
			return s, s != ""
		}),
		hfProp(ComponentPropertiesHuggingFaceLanguage, 0.2, func(src Source) (any, bool) {
			r := src.HF
			if r == nil {
				return nil, false
			}
			s := extractLanguage(r.CardData)
			return s, s != ""
		}),
		hfProp(ComponentPropertiesHuggingFaceUsedStorage, 0.2, func(src Source) (any, bool) {
			r := src.HF
			if r == nil || r.UsedStorage <= 0 {
				return nil, false
			}
			return r.UsedStorage, true
		}),
		hfProp(ComponentPropertiesHuggingFacePrivate, 0.2, func(src Source) (any, bool) {
			r := src.HF
			if r == nil {
				return nil, false
			}
			// keep boolean present always (even false) if HF response exists
			return r.Private, true
		}),
		hfProp(ComponentPropertiesHuggingFaceLibraryName, 0.2, func(src Source) (any, bool) {
			r := src.HF
			if r == nil {
				return nil, false
			}
			s := strings.TrimSpace(r.LibraryName)
			return s, s != ""
		}),
		hfProp(ComponentPropertiesHuggingFaceDownloads, 0.2, func(src Source) (any, bool) {
			r := src.HF
			if r == nil || r.Downloads <= 0 {
				return nil, false
			}
			return r.Downloads, true
		}),
		hfProp(ComponentPropertiesHuggingFaceLikes, 0.2, func(src Source) (any, bool) {
			r := src.HF
			if r == nil || r.Likes <= 0 {
				return nil, false
			}
			return r.Likes, true
		}),
		hfProp(ComponentPropertiesModelCardBaseModel, 0.2, func(src Source) (any, bool) {
			r := src.Readme
			if r == nil {
				return nil, false
			}
			s := strings.TrimSpace(r.BaseModel)
			return s, s != ""
		}),
		hfProp(ComponentPropertiesModelCardContact, 0.2, func(src Source) (any, bool) {
			r := src.Readme
			if r == nil {
				return nil, false
			}
			s := strings.TrimSpace(r.ModelCardContact)
			return s, s != ""
		}),
		// Model card fields
		{
			Key:      ModelCardModelParametersTask,
			Weight:   1.0,
			Required: false,
			Apply: func(src Source, tgt Target) {
				if tgt.ModelCard == nil {
					return
				}
				if mp := bomModelParameters(tgt.BOM); mp != nil && strings.TrimSpace(mp.Task) != "" {
					return
				}
				s := ""
				if src.HF != nil {
					s = strings.TrimSpace(src.HF.PipelineTag)
				}
				if s == "" && src.Readme != nil {
					s = strings.TrimSpace(src.Readme.TaskType)
				}
				if s == "" {
					return
				}
				mp := ensureModelParameters(tgt.ModelCard)
				mp.Task = s
				logf(src.ModelID, "apply %s set=%s", ModelCardModelParametersTask, summarizeValue(s))
			},
			Present: func(b *cdx.BOM) bool {
				mp := bomModelParameters(b)
				ok := mp != nil && strings.TrimSpace(mp.Task) != ""
				mid := ""
				if c := bomComponent(b); c != nil {
					mid = c.Name
				}
				logf(mid, "present %s ok=%t", ModelCardModelParametersTask, ok)
				return ok
			},
		},
		{
			Key:      ModelCardModelParametersArchitectureFamily,
			Weight:   0.5,
			Required: false,
			Apply: func(src Source, tgt Target) {
				if tgt.ModelCard == nil || src.HF == nil {
					return
				}
				s := strings.TrimSpace(src.HF.Config.ModelType)
				if s == "" {
					return
				}
				mp := ensureModelParameters(tgt.ModelCard)
				mp.ArchitectureFamily = s
				logf(src.ModelID, "apply %s set=%s", ModelCardModelParametersArchitectureFamily, summarizeValue(s))
			},
			Present: func(b *cdx.BOM) bool {
				mp := bomModelParameters(b)
				ok := mp != nil && strings.TrimSpace(mp.ArchitectureFamily) != ""
				mid := ""
				if c := bomComponent(b); c != nil {
					mid = c.Name
				}
				logf(mid, "present %s ok=%t", ModelCardModelParametersArchitectureFamily, ok)
				return ok
			},
		},
		{
			Key:      ModelCardModelParametersModelArchitecture,
			Weight:   0.5,
			Required: false,
			Apply: func(src Source, tgt Target) {
				if tgt.ModelCard == nil || src.HF == nil {
					return
				}
				if len(src.HF.Config.Architectures) == 0 {
					return
				}
				s := strings.TrimSpace(src.HF.Config.Architectures[0])
				if s == "" {
					return
				}
				mp := ensureModelParameters(tgt.ModelCard)
				mp.ModelArchitecture = s
				logf(src.ModelID, "apply %s set=%s", ModelCardModelParametersModelArchitecture, summarizeValue(s))
			},
			Present: func(b *cdx.BOM) bool {
				mp := bomModelParameters(b)
				ok := mp != nil && strings.TrimSpace(mp.ModelArchitecture) != ""
				mid := ""
				if c := bomComponent(b); c != nil {
					mid = c.Name
				}
				logf(mid, "present %s ok=%t", ModelCardModelParametersModelArchitecture, ok)
				return ok
			},
		},
		{
			Key:      ModelCardModelParametersDatasets,
			Weight:   0.5,
			Required: false,
			Apply: func(src Source, tgt Target) {
				if tgt.ModelCard == nil {
					return
				}
				// Do not overwrite existing datasets
				if tgt.ModelCard.ModelParameters != nil && tgt.ModelCard.ModelParameters.Datasets != nil && len(*tgt.ModelCard.ModelParameters.Datasets) > 0 {
					return
				}
				var ds []string
				if src.HF != nil {
					ds = extractDatasets(src.HF.CardData, src.HF.Tags)
				}
				if len(ds) == 0 && src.Readme != nil {
					ds = normalizeStrings(src.Readme.Datasets)
					for i := range ds {
						ds[i] = normalizeDatasetRef(ds[i])
					}
				}
				if len(ds) == 0 {
					return
				}
				choices := make([]cdx.MLDatasetChoice, 0, len(ds))
				for _, ref := range ds {
					ref = strings.TrimSpace(ref)
					choices = append(choices, cdx.MLDatasetChoice{Ref: ref})
				}
				mp := ensureModelParameters(tgt.ModelCard)
				mp.Datasets = &choices
				logf(src.ModelID, "apply %s set=%s", ModelCardModelParametersDatasets, summarizeValue(ds))
			},
			Present: func(b *cdx.BOM) bool {
				mp := bomModelParameters(b)
				if mp == nil || mp.Datasets == nil || len(*mp.Datasets) == 0 {
					mid := ""
					if c := bomComponent(b); c != nil {
						mid = c.Name
					}
					logf(mid, "present %s ok=false", ModelCardModelParametersDatasets)
					return false
				}
				for _, d := range *mp.Datasets {
					if strings.TrimSpace(d.Ref) != "" {
						mid := ""
						if c := bomComponent(b); c != nil {
							mid = c.Name
						}
						logf(mid, "present %s ok=true", ModelCardModelParametersDatasets)
						return true
					}
				}
				mid := ""
				if c := bomComponent(b); c != nil {
					mid = c.Name
				}
				logf(mid, "present %s ok=false", ModelCardModelParametersDatasets)
				return false
			},
		},
		{
			Key:      ModelCardConsiderationsUseCases,
			Weight:   0.5,
			Required: false,
			Apply: func(src Source, tgt Target) {
				if tgt.ModelCard == nil || src.Readme == nil {
					return
				}
				cons := ensureConsiderations(tgt.ModelCard)
				if cons.UseCases != nil && len(*cons.UseCases) > 0 {
					return
				}
				useCases := []string{}
				if s := strings.TrimSpace(src.Readme.DirectUse); s != "" {
					useCases = append(useCases, s)
				}
				if s := strings.TrimSpace(src.Readme.OutOfScopeUse); s != "" {
					useCases = append(useCases, "out-of-scope: "+s)
				}
				useCases = normalizeStrings(useCases)
				if len(useCases) == 0 {
					return
				}
				cons.UseCases = &useCases
				logf(src.ModelID, "apply %s set=%s", ModelCardConsiderationsUseCases, summarizeValue(useCases))
			},
			Present: func(b *cdx.BOM) bool {
				c := bomComponent(b)
				ok := c != nil && c.ModelCard != nil && c.ModelCard.Considerations != nil && c.ModelCard.Considerations.UseCases != nil && len(*c.ModelCard.Considerations.UseCases) > 0
				mid := ""
				if c != nil {
					mid = c.Name
				}
				logf(mid, "present %s ok=%t", ModelCardConsiderationsUseCases, ok)
				return ok
			},
		},
		{
			Key:      ModelCardConsiderationsTechnicalLimitations,
			Weight:   0.5,
			Required: false,
			Apply: func(src Source, tgt Target) {
				if tgt.ModelCard == nil || src.Readme == nil {
					return
				}
				cons := ensureConsiderations(tgt.ModelCard)
				if cons.TechnicalLimitations != nil && len(*cons.TechnicalLimitations) > 0 {
					return
				}
				s := strings.TrimSpace(src.Readme.BiasRisksLimitations)
				if s == "" {
					return
				}
				vals := []string{s}
				cons.TechnicalLimitations = &vals
				logf(src.ModelID, "apply %s set=%s", ModelCardConsiderationsTechnicalLimitations, summarizeValue(s))
			},
			Present: func(b *cdx.BOM) bool {
				c := bomComponent(b)
				ok := c != nil && c.ModelCard != nil && c.ModelCard.Considerations != nil && c.ModelCard.Considerations.TechnicalLimitations != nil && len(*c.ModelCard.Considerations.TechnicalLimitations) > 0
				mid := ""
				if c != nil {
					mid = c.Name
				}
				logf(mid, "present %s ok=%t", ModelCardConsiderationsTechnicalLimitations, ok)
				return ok
			},
		},
		{
			Key:      ModelCardConsiderationsEthicalConsiderations,
			Weight:   0.25,
			Required: false,
			Apply: func(src Source, tgt Target) {
				if tgt.ModelCard == nil || src.Readme == nil {
					return
				}
				cons := ensureConsiderations(tgt.ModelCard)
				if cons.EthicalConsiderations != nil && len(*cons.EthicalConsiderations) > 0 {
					return
				}
				name := strings.TrimSpace(src.Readme.BiasRisksLimitations)
				mit := strings.TrimSpace(src.Readme.BiasRecommendations)
				if name == "" && mit == "" {
					return
				}
				if name == "" {
					name = "bias_risks_limitations"
				}
				ethics := []cdx.MLModelCardEthicalConsideration{{Name: name, MitigationStrategy: mit}}
				cons.EthicalConsiderations = &ethics
				logf(src.ModelID, "apply %s set=true", ModelCardConsiderationsEthicalConsiderations)
			},
			Present: func(b *cdx.BOM) bool {
				c := bomComponent(b)
				ok := c != nil && c.ModelCard != nil && c.ModelCard.Considerations != nil && c.ModelCard.Considerations.EthicalConsiderations != nil && len(*c.ModelCard.Considerations.EthicalConsiderations) > 0
				mid := ""
				if c != nil {
					mid = c.Name
				}
				logf(mid, "present %s ok=%t", ModelCardConsiderationsEthicalConsiderations, ok)
				return ok
			},
		},
		{
			Key:      ModelCardQuantitativeAnalysisPerformanceMetrics,
			Weight:   0.5,
			Required: false,
			Apply: func(src Source, tgt Target) {
				if tgt.ModelCard == nil || src.Readme == nil {
					return
				}
				qa := ensureQuantitativeAnalysis(tgt.ModelCard)
				if qa.PerformanceMetrics != nil && len(*qa.PerformanceMetrics) > 0 {
					return
				}
				metrics := make([]cdx.MLPerformanceMetric, 0)

				// 1) From model-index in README YAML (detailed metrics with values)
				for _, m := range src.Readme.ModelIndexMetrics {
					mt := strings.TrimSpace(m.Type)
					mv := strings.TrimSpace(m.Value)
					if mt == "" && mv == "" {
						continue
					}
					metrics = append(metrics, cdx.MLPerformanceMetric{Type: mt, Value: mv})
				}

				// 2) From simple metrics list in YAML frontmatter
				for _, mt := range src.Readme.Metrics {
					mt = strings.TrimSpace(mt)
					if mt == "" {
						continue
					}
					// Check if already added from model-index
					alreadyExists := false
					for _, existing := range metrics {
						if existing.Type == mt {
							alreadyExists = true
							break
						}
					}
					if !alreadyExists {
						metrics = append(metrics, cdx.MLPerformanceMetric{Type: mt, Value: ""})
					}
				}

				// 3) From markdown sections (may contain placeholders for templates)
				// If we have TestingMetrics or Results sections, add them as a generic entry
				if len(metrics) == 0 {
					testingMetrics := strings.TrimSpace(src.Readme.TestingMetrics)
					results := strings.TrimSpace(src.Readme.Results)

					if testingMetrics != "" || results != "" {
						// Create a placeholder metric entry to indicate presence of these sections
						metricType := "testing_metrics"
						metricValue := ""

						if testingMetrics != "" {
							metricType = testingMetrics
						}
						if results != "" {
							metricValue = results
						}

						metrics = append(metrics, cdx.MLPerformanceMetric{
							Type:  metricType,
							Value: metricValue,
						})
					}
				}

				if len(metrics) == 0 {
					return
				}
				qa.PerformanceMetrics = &metrics
				logf(src.ModelID, "apply %s set=%s", ModelCardQuantitativeAnalysisPerformanceMetrics, summarizeValue(metrics))
			},
			Present: func(b *cdx.BOM) bool {
				c := bomComponent(b)
				ok := c != nil && c.ModelCard != nil && c.ModelCard.QuantitativeAnalysis != nil && c.ModelCard.QuantitativeAnalysis.PerformanceMetrics != nil && len(*c.ModelCard.QuantitativeAnalysis.PerformanceMetrics) > 0
				mid := ""
				if c != nil {
					mid = c.Name
				}
				logf(mid, "present %s ok=%t", ModelCardQuantitativeAnalysisPerformanceMetrics, ok)
				return ok
			},
		},
		{
			Key:      ModelCardConsiderationsEnvironmentalConsiderationsProperties,
			Weight:   0.25,
			Required: false,
			Apply: func(src Source, tgt Target) {
				if tgt.ModelCard == nil || src.Readme == nil {
					return
				}
				// If already populated, don't overwrite.
				if tgt.ModelCard.Considerations != nil && tgt.ModelCard.Considerations.EnvironmentalConsiderations != nil {
					env := tgt.ModelCard.Considerations.EnvironmentalConsiderations
					if env.Properties != nil && len(*env.Properties) > 0 {
						return
					}
				}

				props := []cdx.Property{}
				add := func(name, value string) {
					name = strings.TrimSpace(name)
					value = strings.TrimSpace(value)
					if name == "" || value == "" {
						return
					}
					props = append(props, cdx.Property{Name: name, Value: value})
				}

				add("hardwareType", src.Readme.EnvironmentalHardwareType)
				add("hoursUsed", src.Readme.EnvironmentalHoursUsed)
				add("cloudProvider", src.Readme.EnvironmentalCloudProvider)
				add("computeRegion", src.Readme.EnvironmentalComputeRegion)
				add("carbonEmitted", src.Readme.EnvironmentalCarbonEmitted)

				if len(props) == 0 {
					return
				}

				cons := ensureConsiderations(tgt.ModelCard)
				if cons.EnvironmentalConsiderations == nil {
					cons.EnvironmentalConsiderations = &cdx.MLModelCardEnvironmentalConsiderations{}
				}
				cons.EnvironmentalConsiderations.Properties = &props
				logf(src.ModelID, "apply %s set=%s", ModelCardConsiderationsEnvironmentalConsiderationsProperties, summarizeValue(props))
			},
			Present: func(b *cdx.BOM) bool {
				c := bomComponent(b)
				ok := c != nil && c.ModelCard != nil && c.ModelCard.Considerations != nil && c.ModelCard.Considerations.EnvironmentalConsiderations != nil && c.ModelCard.Considerations.EnvironmentalConsiderations.Properties != nil && len(*c.ModelCard.Considerations.EnvironmentalConsiderations.Properties) > 0
				mid := ""
				if c != nil {
					mid = c.Name
				}
				logf(mid, "present %s ok=%t", ModelCardConsiderationsEnvironmentalConsiderationsProperties, ok)
				return ok
			},
		},
	}
}

func hfProp(key Key, weight float64, get func(src Source) (any, bool)) FieldSpec {
	return FieldSpec{
		Key:      key,
		Weight:   weight,
		Required: false,
		Apply: func(src Source, tgt Target) {
			if tgt.Component == nil || get == nil {
				return
			}
			v, ok := get(src)
			if !ok || v == nil {
				return
			}
			propName := strings.TrimPrefix(key.String(), "BOM.metadata.component.properties.")
			setProperty(tgt.Component, propName, strings.TrimSpace(fmt.Sprint(v)))
			logf(src.ModelID, "apply %s set=%s", key, summarizeValue(v))
		},
		Present: func(b *cdx.BOM) bool {
			c := bomComponent(b)
			propName := strings.TrimPrefix(key.String(), "BOM.metadata.component.properties.")
			ok := c != nil && hasProperty(c, propName)
			mid := ""
			if c != nil {
				mid = c.Name
			}
			logf(mid, "present %s ok=%t", key, ok)
			return ok
		},
	}
}

func ensureModelParameters(card *cdx.MLModelCard) *cdx.MLModelParameters {
	if card.ModelParameters == nil {
		card.ModelParameters = &cdx.MLModelParameters{}
	}
	return card.ModelParameters
}

func ensureConsiderations(card *cdx.MLModelCard) *cdx.MLModelCardConsiderations {
	if card.Considerations == nil {
		card.Considerations = &cdx.MLModelCardConsiderations{}
	}
	return card.Considerations
}

func ensureQuantitativeAnalysis(card *cdx.MLModelCard) *cdx.MLQuantitativeAnalysis {
	if card.QuantitativeAnalysis == nil {
		card.QuantitativeAnalysis = &cdx.MLQuantitativeAnalysis{}
	}
	return card.QuantitativeAnalysis
}

func normalizeDatasetRef(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "dataset:") {
		return s
	}
	// If it already looks like a namespaced identifier (e.g., "org/ds"), still prefix with dataset:
	return "dataset:" + s
}

func setProperty(c *cdx.Component, name, value string) {
	if c == nil {
		return
	}
	name = strings.TrimSpace(name)
	value = strings.TrimSpace(value)
	if name == "" || value == "" {
		return
	}
	if c.Properties == nil {
		c.Properties = &[]cdx.Property{}
	}
	*c.Properties = append(*c.Properties, cdx.Property{Name: name, Value: value})
}

func hasProperty(c *cdx.Component, name string) bool {
	if c == nil || c.Properties == nil {
		return false
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}
	for _, p := range *c.Properties {
		if strings.TrimSpace(p.Name) == name && strings.TrimSpace(p.Value) != "" {
			return true
		}
	}
	return false
}

func bomComponent(b *cdx.BOM) *cdx.Component {
	if b == nil || b.Metadata == nil {
		return nil
	}
	return b.Metadata.Component
}

func bomHasComponentName(b *cdx.BOM) bool {
	c := bomComponent(b)
	return c != nil && strings.TrimSpace(c.Name) != ""
}

func bomModelParameters(b *cdx.BOM) *cdx.MLModelParameters {
	c := bomComponent(b)
	if c == nil || c.ModelCard == nil {
		return nil
	}
	return c.ModelCard.ModelParameters
}

// ---- helpers  ----

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
