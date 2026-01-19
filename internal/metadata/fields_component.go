package metadata

import (
	"fmt"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

type componentExternalRefsSource struct {
	ModelID  string
	PaperURL string
	DemoURL  string
}

func componentFields() []FieldSpec {
	return []FieldSpec{
		{
			Key:      ComponentName,
			Weight:   1.0,
			Required: true,
			Sources: []func(Source) (any, bool){
				func(src Source) (any, bool) {
					if s := strings.TrimSpace(src.Scan.Name); s != "" {
						return s, true
					}
					return nil, false
				},
				func(src Source) (any, bool) {
					if src.HF == nil {
						return nil, false
					}
					if s := strings.TrimSpace(src.HF.ID); s != "" {
						return s, true
					}
					return nil, false
				},
				func(src Source) (any, bool) {
					if src.HF == nil {
						return nil, false
					}
					if s := strings.TrimSpace(src.HF.ModelID); s != "" {
						return s, true
					}
					return nil, false
				},
				func(src Source) (any, bool) {
					if s := strings.TrimSpace(src.ModelID); s != "" {
						return s, true
					}
					return nil, false
				},
			},
			Parse: func(value string) (any, error) {
				return parseNonEmptyString(value, "name")
			},
			Apply: func(tgt Target, value any) error {
				input, ok := value.(applyInput)
				if !ok {
					return fmt.Errorf("invalid input for %s", ComponentName)
				}
				name, _ := input.Value.(string)
				name = strings.TrimSpace(name)
				if name == "" {
					return fmt.Errorf("name value is empty")
				}
				if tgt.Component == nil {
					return fmt.Errorf("component is nil")
				}
				tgt.Component.Name = name
				logf("", "apply %s set=%s", ComponentName, summarizeValue(name))
				return nil
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
			Sources: []func(Source) (any, bool){
				func(src Source) (any, bool) {
					modelID := strings.TrimSpace(src.ModelID)
					if modelID == "" {
						return nil, false
					}
					input := componentExternalRefsSource{ModelID: modelID}
					if src.Readme != nil {
						input.PaperURL = strings.TrimSpace(src.Readme.PaperURL)
						input.DemoURL = strings.TrimSpace(src.Readme.DemoURL)
					}
					return input, true
				},
			},
			Parse: func(value string) (any, error) {
				return parseNonEmptyString(value, "externalReferences")
			},
			Apply: func(tgt Target, value any) error {
				input, ok := value.(applyInput)
				if !ok {
					return fmt.Errorf("invalid input for %s", ComponentExternalReferences)
				}
				if tgt.Component == nil {
					return fmt.Errorf("component is nil")
				}

				var refs []cdx.ExternalReference

				switch v := input.Value.(type) {
				case string:
					url := strings.TrimSpace(v)
					if url == "" {
						return fmt.Errorf("externalReferences value is empty")
					}
					refs = []cdx.ExternalReference{{
						Type: cdx.ExternalReferenceType("website"),
						URL:  url,
					}}
				case componentExternalRefsSource:
					base := strings.TrimSpace(tgt.HuggingFaceBaseURL)
					if base == "" {
						base = "https://huggingface.co/"
					}
					if !strings.HasSuffix(base, "/") {
						base += "/"
					}
					url := base + strings.TrimPrefix(v.ModelID, "/")
					refs = []cdx.ExternalReference{{
						Type: cdx.ExternalReferenceType("website"),
						URL:  url,
					}}
					if v.PaperURL != "" {
						refs = append(refs, cdx.ExternalReference{
							Type: cdx.ExternalReferenceType("documentation"),
							URL:  v.PaperURL,
						})
					}
					if v.DemoURL != "" {
						refs = append(refs, cdx.ExternalReference{
							Type: cdx.ExternalReferenceType("other"),
							URL:  v.DemoURL,
						})
					}
				default:
					return fmt.Errorf("invalid externalReferences value")
				}

				tgt.Component.ExternalReferences = &refs
				logf("", "apply %s set=%s", ComponentExternalReferences, summarizeValue(refs))
				return nil
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
			Sources: []func(Source) (any, bool){
				func(src Source) (any, bool) {
					if src.HF != nil && len(src.HF.Tags) > 0 {
						tags := normalizeStrings(src.HF.Tags)
						if len(tags) > 0 {
							return tags, true
						}
					}
					return nil, false
				},
				func(src Source) (any, bool) {
					if src.Readme != nil && len(src.Readme.Tags) > 0 {
						tags := normalizeStrings(src.Readme.Tags)
						if len(tags) > 0 {
							return tags, true
						}
					}
					return nil, false
				},
			},
			Parse: func(value string) (any, error) {
				return parseTagsPreserveEmpty(value, "tags")
			},
			Apply: func(tgt Target, value any) error {
				input, ok := value.(applyInput)
				if !ok {
					return fmt.Errorf("invalid input for %s", ComponentTags)
				}
				tags, _ := input.Value.([]string)
				if len(tags) == 0 {
					return fmt.Errorf("tags value is empty")
				}
				if tgt.Component == nil {
					return fmt.Errorf("component is nil")
				}
				if !input.Force && tgt.Component.Tags != nil && len(*tgt.Component.Tags) > 0 {
					return nil
				}
				tgt.Component.Tags = &tags
				logf("", "apply %s set=%s", ComponentTags, summarizeValue(tags))
				return nil
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
			Sources: []func(Source) (any, bool){
				func(src Source) (any, bool) {
					if src.HF == nil {
						return nil, false
					}
					lic := extractLicense(src.HF.CardData, src.HF.Tags)
					if lic == "" {
						return nil, false
					}
					return lic, true
				},
				func(src Source) (any, bool) {
					if src.Readme == nil {
						return nil, false
					}
					lic := strings.TrimSpace(src.Readme.License)
					if lic == "" {
						return nil, false
					}
					return lic, true
				},
			},
			Parse: func(value string) (any, error) {
				return parseNonEmptyString(value, "license")
			},
			Apply: func(tgt Target, value any) error {
				input, ok := value.(applyInput)
				if !ok {
					return fmt.Errorf("invalid input for %s", ComponentLicenses)
				}
				lic, _ := input.Value.(string)
				lic = strings.TrimSpace(lic)
				if lic == "" {
					return fmt.Errorf("license value is empty")
				}
				if tgt.Component == nil {
					return fmt.Errorf("component is nil")
				}
				if !input.Force && tgt.Component.Licenses != nil && len(*tgt.Component.Licenses) > 0 {
					return nil
				}
				ls := cdx.Licenses{
					{License: &cdx.License{Name: lic}},
				}
				tgt.Component.Licenses = &ls
				logf("", "apply %s set=%s", ComponentLicenses, summarizeValue(lic))
				return nil
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
			Sources: []func(Source) (any, bool){
				func(src Source) (any, bool) {
					if src.HF == nil {
						return nil, false
					}
					sha := strings.TrimSpace(src.HF.SHA)
					if sha == "" {
						return nil, false
					}
					return sha, true
				},
			},
			Parse: func(value string) (any, error) {
				return parseNonEmptyString(value, "hash")
			},
			Apply: func(tgt Target, value any) error {
				input, ok := value.(applyInput)
				if !ok {
					return fmt.Errorf("invalid input for %s", ComponentHashes)
				}
				sha, _ := input.Value.(string)
				sha = strings.TrimSpace(sha)
				if sha == "" {
					return fmt.Errorf("hash value is empty")
				}
				if tgt.Component == nil {
					return fmt.Errorf("component is nil")
				}
				hs := []cdx.Hash{{Algorithm: cdx.HashAlgoSHA1, Value: sha}}
				tgt.Component.Hashes = &hs
				logf("", "apply %s set=%s", ComponentHashes, summarizeValue(sha))
				return nil
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
			Sources: []func(Source) (any, bool){
				func(src Source) (any, bool) {
					if src.HF != nil {
						if s := strings.TrimSpace(src.HF.Author); s != "" {
							return s, true
						}
					}
					return nil, false
				},
				func(src Source) (any, bool) {
					if src.Readme != nil {
						if s := strings.TrimSpace(src.Readme.DevelopedBy); s != "" {
							return s, true
						}
					}
					return nil, false
				},
			},
			Parse: func(value string) (any, error) {
				return parseNonEmptyString(value, "manufacturer")
			},
			Apply: func(tgt Target, value any) error {
				input, ok := value.(applyInput)
				if !ok {
					return fmt.Errorf("invalid input for %s", ComponentManufacturer)
				}
				s, _ := input.Value.(string)
				s = strings.TrimSpace(s)
				if s == "" {
					return fmt.Errorf("manufacturer value is empty")
				}
				if tgt.Component == nil {
					return fmt.Errorf("component is nil")
				}
				if !input.Force && tgt.Component.Manufacturer != nil && strings.TrimSpace(tgt.Component.Manufacturer.Name) != "" {
					return nil
				}
				tgt.Component.Manufacturer = &cdx.OrganizationalEntity{Name: s}
				logf("", "apply %s set=%s", ComponentManufacturer, summarizeValue(s))
				return nil
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
			Sources: []func(Source) (any, bool){
				func(src Source) (any, bool) {
					if src.HF != nil {
						if s := strings.TrimSpace(src.HF.Author); s != "" {
							return s, true
						}
					}
					return nil, false
				},
				func(src Source) (any, bool) {
					if src.Readme != nil {
						if s := strings.TrimSpace(src.Readme.DevelopedBy); s != "" {
							return s, true
						}
					}
					return nil, false
				},
			},
			Parse: func(value string) (any, error) {
				return parseNonEmptyString(value, "group")
			},
			Apply: func(tgt Target, value any) error {
				input, ok := value.(applyInput)
				if !ok {
					return fmt.Errorf("invalid input for %s", ComponentGroup)
				}
				s, _ := input.Value.(string)
				s = strings.TrimSpace(s)
				if s == "" {
					return fmt.Errorf("group value is empty")
				}
				if tgt.Component == nil {
					return fmt.Errorf("component is nil")
				}
				if !input.Force && strings.TrimSpace(tgt.Component.Group) != "" {
					return nil
				}
				tgt.Component.Group = s
				logf("", "apply %s set=%s", ComponentGroup, summarizeValue(s))
				return nil
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
	}
}

func evidenceFields() []FieldSpec {
	return []FieldSpec{
		{
			Key:      Key("aibomgen.evidence"),
			Weight:   0,
			Required: false,
			Sources: []func(Source) (any, bool){
				func(src Source) (any, bool) {
					return src, true
				},
			},
			Apply: func(tgt Target, value any) error {
				input, ok := value.(applyInput)
				if !ok {
					return fmt.Errorf("invalid input for aibomgen.evidence")
				}
				src, ok := input.Value.(Source)
				if !ok {
					return fmt.Errorf("invalid evidence value")
				}
				if tgt.Component == nil || !tgt.IncludeEvidenceProperties {
					return nil
				}
				setProperty(tgt.Component, "aibomgen.type", src.Scan.Type)
				setProperty(tgt.Component, "aibomgen.evidence", src.Scan.Evidence)
				setProperty(tgt.Component, "aibomgen.path", src.Scan.Path)
				logf(src.ModelID, "apply aibomgen.evidence type=%s path=%s evidence=%s", summarizeValue(src.Scan.Type), summarizeValue(src.Scan.Path), summarizeValue(src.Scan.Evidence))
				return nil
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
	}
}
