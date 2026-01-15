package metadata

import (
	"fmt"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/idlab-discover/AIBoMGen-cli/internal/fetcher"
	"github.com/idlab-discover/AIBoMGen-cli/internal/scanner"
)

// DatasetKey identifies dataset-specific CycloneDX fields
type DatasetKey string

func (k DatasetKey) String() string { return string(k) }

const (
	// BOM.components[DATA].* (DATASET)
	DatasetName               DatasetKey = "BOM.components[DATA].name"
	DatasetExternalReferences DatasetKey = "BOM.components[DATA].externalReferences"
	DatasetTags               DatasetKey = "BOM.components[DATA].tags"
	DatasetLicenses           DatasetKey = "BOM.components[DATA].licenses"
	DatasetDescription        DatasetKey = "BOM.components[DATA].description"
	DatasetManufacturer       DatasetKey = "BOM.components[DATA].manufacturer"
	DatasetGroup              DatasetKey = "BOM.components[DATA].group"
	DatasetContents           DatasetKey = "BOM.components[DATA].data.contents.attachments"
	DatasetSensitiveData      DatasetKey = "BOM.components[DATA].data.sensitiveData"
	DatasetContact            DatasetKey = "BOM.components[DATA].properties.contact"
)

// DatasetSource mirrors Source but for datasets
type DatasetSource struct {
	DatasetID string
	Scan      scanner.Discovery
	HF        *fetcher.DatasetAPIResponse
	Readme    *fetcher.DatasetReadmeCard
}

// DatasetTarget is the dataset component being built
type DatasetTarget struct {
	Component *cdx.Component

	// Options
	IncludeEvidenceProperties bool
	HuggingFaceBaseURL        string
}

// DatasetFieldSpec is the dataset analog of FieldSpec
type DatasetFieldSpec struct {
	Key      DatasetKey
	Weight   float64
	Required bool

	Apply        func(src DatasetSource, tgt DatasetTarget)
	Present      func(comp *cdx.Component) bool
	SetUserValue func(value string, tgt DatasetTarget) error
}

// DatasetRegistry returns all dataset field specifications
func DatasetRegistry() []DatasetFieldSpec {
	return []DatasetFieldSpec{
		{
			Key:      DatasetName,
			Weight:   1.0,
			Required: true,
			Apply: func(src DatasetSource, tgt DatasetTarget) {
				if tgt.Component == nil {
					return
				}
				name := strings.TrimSpace(src.DatasetID)
				if src.HF != nil {
					if s := strings.TrimSpace(src.HF.ID); s != "" {
						name = s
					}
				}
				if strings.TrimSpace(src.Scan.Name) != "" {
					name = strings.TrimSpace(src.Scan.Name)
				}
				if name != "" {
					tgt.Component.Name = name
					logf(src.DatasetID, "apply %s set=%s", DatasetName, summarizeValue(name))
				}
			},
			Present: func(comp *cdx.Component) bool {
				ok := comp != nil && strings.TrimSpace(comp.Name) != ""
				logf("", "present %s ok=%t", DatasetName, ok)
				return ok
			},
			SetUserValue: func(value string, tgt DatasetTarget) error {
				if tgt.Component == nil {
					return fmt.Errorf("component is nil")
				}
				name := strings.TrimSpace(value)
				if name == "" {
					return fmt.Errorf("name value is empty")
				}
				tgt.Component.Name = name
				return nil
			},
		},
		{
			Key:      DatasetExternalReferences,
			Weight:   0.5,
			Required: false,
			Apply: func(src DatasetSource, tgt DatasetTarget) {
				if tgt.Component == nil {
					return
				}
				datasetID := strings.TrimSpace(src.DatasetID)
				if datasetID == "" {
					return
				}

				base := strings.TrimSpace(tgt.HuggingFaceBaseURL)
				if base == "" {
					base = "https://huggingface.co/"
				}
				if !strings.HasSuffix(base, "/") {
					base += "/"
				}

				// Use /datasets/ path for HF datasets
				url := base + "datasets/" + strings.TrimPrefix(datasetID, "/")
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
				logf(src.DatasetID, "apply %s set=%d refs", DatasetExternalReferences, len(refs))
			},
			Present: func(comp *cdx.Component) bool {
				ok := comp != nil && comp.ExternalReferences != nil && len(*comp.ExternalReferences) > 0
				return ok
			},
			SetUserValue: func(value string, tgt DatasetTarget) error {
				if tgt.Component == nil {
					return fmt.Errorf("component is nil")
				}
				url := strings.TrimSpace(value)
				if url == "" {
					return fmt.Errorf("externalReferences value is empty")
				}
				refs := []cdx.ExternalReference{{
					Type: cdx.ExternalReferenceType("website"),
					URL:  url,
				}}
				tgt.Component.ExternalReferences = &refs
				return nil
			},
		},
		{
			Key:      DatasetTags,
			Weight:   0.5,
			Required: false,
			Apply: func(src DatasetSource, tgt DatasetTarget) {
				if tgt.Component == nil || (tgt.Component.Tags != nil && len(*tgt.Component.Tags) > 0) {
					return
				}
				var tags []string
				if src.HF != nil && len(src.HF.Tags) > 0 {
					tags = normalizeStrings(src.HF.Tags)
				} else if src.Readme != nil && len(src.Readme.Tags) > 0 {
					tags = normalizeStrings(src.Readme.Tags)
				}
				if len(tags) > 0 {
					tgt.Component.Tags = &tags
					logf(src.DatasetID, "apply %s set=%d tags", DatasetTags, len(tags))
				}
			},
			Present: func(comp *cdx.Component) bool {
				return comp != nil && comp.Tags != nil && len(*comp.Tags) > 0
			},
			SetUserValue: func(value string, tgt DatasetTarget) error {
				if tgt.Component == nil {
					return fmt.Errorf("component is nil")
				}
				tags := normalizeStrings(strings.Split(value, ","))
				tgt.Component.Tags = &tags
				return nil
			},
		},
		{
			Key:      DatasetLicenses,
			Weight:   0.8,
			Required: false,
			Apply: func(src DatasetSource, tgt DatasetTarget) {
				if tgt.Component == nil {
					return
				}
				licenseStr := ""
				if src.Readme != nil && strings.TrimSpace(src.Readme.License) != "" {
					licenseStr = strings.TrimSpace(src.Readme.License)
				} else if src.HF != nil && src.HF.CardData != nil {
					if licData, ok := src.HF.CardData["license"]; ok {
						licenseStr = strings.TrimSpace(fmt.Sprintf("%v", licData))
					}
				}
				if licenseStr == "" {
					return
				}

				ls := cdx.Licenses{
					{License: &cdx.License{Name: licenseStr}},
				}
				tgt.Component.Licenses = &ls
				logf(src.DatasetID, "apply %s set=%s", DatasetLicenses, licenseStr)
			},
			Present: func(comp *cdx.Component) bool {
				return comp != nil && comp.Licenses != nil && len(*comp.Licenses) > 0
			},
			SetUserValue: func(value string, tgt DatasetTarget) error {
				if tgt.Component == nil {
					return fmt.Errorf("component is nil")
				}
				license := strings.TrimSpace(value)
				if license == "" {
					return fmt.Errorf("license value is empty")
				}
				ls := cdx.Licenses{
					{License: &cdx.License{Name: license}},
				}
				tgt.Component.Licenses = &ls
				return nil
			},
		},
		{
			Key:      DatasetDescription,
			Weight:   0.7,
			Required: false,
			Apply: func(src DatasetSource, tgt DatasetTarget) {
				if tgt.Component == nil || src.Readme == nil {
					return
				}
				desc := strings.TrimSpace(src.Readme.DatasetDescription)
				if desc == "" {
					return
				}
				tgt.Component.Description = desc
				logf(src.DatasetID, "apply %s set=%s", DatasetDescription, summarizeValue(desc))
			},
			Present: func(comp *cdx.Component) bool {
				return comp != nil && strings.TrimSpace(comp.Description) != ""
			},
			SetUserValue: func(value string, tgt DatasetTarget) error {
				if tgt.Component == nil {
					return fmt.Errorf("component is nil")
				}
				desc := strings.TrimSpace(value)
				if desc == "" {
					return fmt.Errorf("description value is empty")
				}
				tgt.Component.Description = desc
				return nil
			},
		},
		{
			Key:      DatasetManufacturer,
			Weight:   0.6,
			Required: false,
			Apply: func(src DatasetSource, tgt DatasetTarget) {
				if tgt.Component == nil || src.Readme == nil {
					return
				}
				if len(src.Readme.AnnotationCreators) == 0 {
					return
				}
				tgt.Component.Manufacturer = &cdx.OrganizationalEntity{
					Name: src.Readme.AnnotationCreators[0],
				}
				logf(src.DatasetID, "apply %s set=%s", DatasetManufacturer, src.Readme.AnnotationCreators[0])
			},
			Present: func(comp *cdx.Component) bool {
				return comp != nil && comp.Manufacturer != nil && strings.TrimSpace(comp.Manufacturer.Name) != ""
			},
			SetUserValue: func(value string, tgt DatasetTarget) error {
				if tgt.Component == nil {
					return fmt.Errorf("component is nil")
				}
				name := strings.TrimSpace(value)
				if name == "" {
					return fmt.Errorf("manufacturer value is empty")
				}
				tgt.Component.Manufacturer = &cdx.OrganizationalEntity{Name: name}
				return nil
			},
		},
		{
			Key:      DatasetGroup,
			Weight:   0.4,
			Required: false,
			Apply: func(src DatasetSource, tgt DatasetTarget) {
				if tgt.Component == nil || src.Readme == nil {
					return
				}
				if len(src.Readme.AnnotationCreators) < 2 {
					return
				}
				tgt.Component.Group = src.Readme.AnnotationCreators[1]
				logf(src.DatasetID, "apply %s set=%s", DatasetGroup, src.Readme.AnnotationCreators[1])
			},
			Present: func(comp *cdx.Component) bool {
				return comp != nil && strings.TrimSpace(comp.Group) != ""
			},
			SetUserValue: func(value string, tgt DatasetTarget) error {
				if tgt.Component == nil {
					return fmt.Errorf("component is nil")
				}
				group := strings.TrimSpace(value)
				if group == "" {
					return fmt.Errorf("group value is empty")
				}
				tgt.Component.Group = group
				return nil
			},
		},
		{
			Key:      DatasetContents,
			Weight:   0.5,
			Required: false,
			Apply: func(src DatasetSource, tgt DatasetTarget) {
				if tgt.Component == nil || src.Readme == nil {
					return
				}
				if len(src.Readme.Configs) == 0 {
					return
				}

				// Store config/datafile info in component description or properties
				// as Contents/Attachments may not be available in all CycloneDX versions
				var configInfo []string
				for _, config := range src.Readme.Configs {
					for _, df := range config.DataFiles {
						info := fmt.Sprintf("%s/%s: %s", config.Name, df.Split, df.Path)
						configInfo = append(configInfo, info)
					}
				}

				if len(configInfo) > 0 {
					// Store as properties for now
					for i, info := range configInfo {
						setProperty(tgt.Component, fmt.Sprintf("config_%d", i), info)
					}
					logf(src.DatasetID, "apply %s set=%d config items", DatasetContents, len(configInfo))
				}
			},
			Present: func(comp *cdx.Component) bool {
				return comp != nil && hasProperty(comp, "config_0")
			},
			SetUserValue: func(value string, tgt DatasetTarget) error {
				// Not typically set by user for configs
				return nil
			},
		},
		{
			Key:      DatasetSensitiveData,
			Weight:   0.6,
			Required: false,
			Apply: func(src DatasetSource, tgt DatasetTarget) {
				if tgt.Component == nil || src.Readme == nil {
					return
				}

				var sensitiveItems []string
				if out := strings.TrimSpace(src.Readme.OutOfScopeUse); out != "" {
					sensitiveItems = append(sensitiveItems, out)
				}
				if psi := strings.TrimSpace(src.Readme.PersonalSensitiveInfo); psi != "" {
					sensitiveItems = append(sensitiveItems, psi)
				}
				if brl := strings.TrimSpace(src.Readme.BiasRisksLimitations); brl != "" {
					sensitiveItems = append(sensitiveItems, brl)
				}

				if len(sensitiveItems) > 0 {
					// Store sensitive data info in properties
					for i, item := range sensitiveItems {
						setProperty(tgt.Component, fmt.Sprintf("sensitive_%d", i), item)
					}
					logf(src.DatasetID, "apply %s marked sensitive", DatasetSensitiveData)
				}
			},
			Present: func(comp *cdx.Component) bool {
				return comp != nil && hasProperty(comp, "sensitive_0")
			},
			SetUserValue: func(value string, tgt DatasetTarget) error {
				if tgt.Component == nil {
					return fmt.Errorf("component is nil")
				}
				sensitive := strings.TrimSpace(value)
				if sensitive == "" {
					return fmt.Errorf("sensitive data value is empty")
				}
				setProperty(tgt.Component, "sensitive_info", sensitive)
				return nil
			},
		},
		{
			Key:      DatasetContact,
			Weight:   0.5,
			Required: false,
			Apply: func(src DatasetSource, tgt DatasetTarget) {
				if tgt.Component == nil || src.Readme == nil {
					return
				}
				contact := strings.TrimSpace(src.Readme.DatasetCardContact)
				if contact == "" {
					return
				}
				setProperty(tgt.Component, "contact", contact)
				logf(src.DatasetID, "apply %s set=%s", DatasetContact, contact)
			},
			Present: func(comp *cdx.Component) bool {
				return hasProperty(comp, "contact")
			},
			SetUserValue: func(value string, tgt DatasetTarget) error {
				if tgt.Component == nil {
					return fmt.Errorf("component is nil")
				}
				contact := strings.TrimSpace(value)
				if contact == "" {
					return fmt.Errorf("contact value is empty")
				}
				setProperty(tgt.Component, "contact", contact)
				return nil
			},
		},
	}
}
