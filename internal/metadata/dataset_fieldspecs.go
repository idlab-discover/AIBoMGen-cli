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
	DatasetDescription        DatasetKey = "BOM.components[DATA].data.description"
	DatasetManufacturer       DatasetKey = "BOM.components[DATA].manufacturer"
	DatasetAuthor             DatasetKey = "BOM.components[DATA].author"
	DatasetGroup              DatasetKey = "BOM.components[DATA].group"
	DatasetContents           DatasetKey = "BOM.components[DATA].data.contents.attachments"
	DatasetSensitiveData      DatasetKey = "BOM.components[DATA].data.sensitiveData"
	DatasetClassification     DatasetKey = "BOM.components[DATA].data.classification"
	DatasetGovernance         DatasetKey = "BOM.components[DATA].data.governance"
	DatasetHashes             DatasetKey = "BOM.components[DATA].hashes"
	DatasetContact            DatasetKey = "BOM.components[DATA].properties.contact"
	DatasetCreatedAt          DatasetKey = "BOM.components[DATA].properties.createdAt"
	DatasetUsedStorage        DatasetKey = "BOM.components[DATA].properties.usedStorage"
	DatasetLastModified       DatasetKey = "BOM.components[DATA].tags.lastModified"
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

// Helper functions for working with Component.Data slice
func ensureComponentData(comp *cdx.Component) *cdx.ComponentData {
	if comp.Data == nil {
		comp.Data = &[]cdx.ComponentData{{
			Type: cdx.ComponentDataTypeDataset,
		}}
	} else if len(*comp.Data) == 0 {
		*comp.Data = []cdx.ComponentData{{
			Type: cdx.ComponentDataTypeDataset,
		}}
	}
	// Ensure Type is always set
	data := &(*comp.Data)[0]
	if data.Type == "" {
		data.Type = cdx.ComponentDataTypeDataset
	}
	return data
}

func getComponentData(comp *cdx.Component) *cdx.ComponentData {
	if comp.Data == nil || len(*comp.Data) == 0 {
		return nil
	}
	return &(*comp.Data)[0]
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
				if tgt.Component == nil {
					return
				}

				var desc string
				if src.Readme != nil {
					desc = strings.TrimSpace(src.Readme.DatasetDescription)
				}
				if desc == "" && src.HF != nil {
					desc = strings.TrimSpace(src.HF.Description)
				}
				if desc == "" {
					return
				}

				data := ensureComponentData(tgt.Component)
				data.Description = desc
				logf(src.DatasetID, "apply %s set=%s", DatasetDescription, summarizeValue(desc))
			},
			Present: func(comp *cdx.Component) bool {
				data := getComponentData(comp)
				return data != nil && strings.TrimSpace(data.Description) != ""
			},
			SetUserValue: func(value string, tgt DatasetTarget) error {
				if tgt.Component == nil {
					return fmt.Errorf("component is nil")
				}
				desc := strings.TrimSpace(value)
				if desc == "" {
					return fmt.Errorf("description value is empty")
				}
				data := ensureComponentData(tgt.Component)
				data.Description = desc
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

				data := ensureComponentData(tgt.Component)

				// Build content string from configs and data files
				var contentParts []string
				for _, config := range src.Readme.Configs {
					for _, df := range config.DataFiles {
						contentParts = append(contentParts, fmt.Sprintf("config:%s split:%s path:%s", config.Name, df.Split, df.Path))
					}
				}

				if len(contentParts) > 0 {
					if data.Contents == nil {
						data.Contents = &cdx.ComponentDataContents{}
					}
					data.Contents.Attachment = &cdx.AttachedText{
						Content:     strings.Join(contentParts, "\n"),
						ContentType: "text/plain",
					}
					logf(src.DatasetID, "apply %s set=%d config items", DatasetContents, len(contentParts))
				}
			},
			Present: func(comp *cdx.Component) bool {
				data := getComponentData(comp)
				return data != nil && data.Contents != nil && data.Contents.Attachment != nil
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
				if tgt.Component == nil {
					return
				}

				var sensitiveItems []string

				// From tags in cardData (HF API)
				if src.HF != nil && src.HF.CardData != nil {
					if tagsData, ok := src.HF.CardData["tags"]; ok {
						if tags, ok := tagsData.([]interface{}); ok {
							for _, tag := range tags {
								if tagStr, ok := tag.(string); ok {
									sensitiveItems = append(sensitiveItems, tagStr)
								}
							}
						}
					}
				}

				// From Readme fields
				if src.Readme != nil {
					if out := strings.TrimSpace(src.Readme.OutOfScopeUse); out != "" {
						sensitiveItems = append(sensitiveItems, "out-of-scope: "+out)
					}
					if psi := strings.TrimSpace(src.Readme.PersonalSensitiveInfo); psi != "" {
						sensitiveItems = append(sensitiveItems, "personal-info: "+psi)
					}
					if brl := strings.TrimSpace(src.Readme.BiasRisksLimitations); brl != "" {
						sensitiveItems = append(sensitiveItems, "bias-risks: "+brl)
					}
				}

				if len(sensitiveItems) > 0 {
					data := ensureComponentData(tgt.Component)
					data.SensitiveData = &sensitiveItems
					logf(src.DatasetID, "apply %s set=%d items", DatasetSensitiveData, len(sensitiveItems))
				}
			},
			Present: func(comp *cdx.Component) bool {
				data := getComponentData(comp)
				return data != nil && data.SensitiveData != nil && len(*data.SensitiveData) > 0
			},
			SetUserValue: func(value string, tgt DatasetTarget) error {
				if tgt.Component == nil {
					return fmt.Errorf("component is nil")
				}
				sensitive := strings.TrimSpace(value)
				if sensitive == "" {
					return fmt.Errorf("sensitive data value is empty")
				}
				data := ensureComponentData(tgt.Component)
				items := []string{sensitive}
				data.SensitiveData = &items
				return nil
			},
		},
		{
			Key:      DatasetClassification,
			Weight:   0.6,
			Required: false,
			Apply: func(src DatasetSource, tgt DatasetTarget) {
				if tgt.Component == nil {
					return
				}

				var classification string
				if src.HF != nil && src.HF.CardData != nil {
					if taskCats, ok := src.HF.CardData["task_categories"]; ok {
						if cats, ok := taskCats.([]interface{}); ok && len(cats) > 0 {
							if cat, ok := cats[0].(string); ok {
								classification = cat
							}
						}
					}
				}

				if classification != "" {
					data := ensureComponentData(tgt.Component)
					data.Classification = classification
					logf(src.DatasetID, "apply %s set=%s", DatasetClassification, classification)
				}
			},
			Present: func(comp *cdx.Component) bool {
				data := getComponentData(comp)
				return data != nil && strings.TrimSpace(data.Classification) != ""
			},
			SetUserValue: func(value string, tgt DatasetTarget) error {
				if tgt.Component == nil {
					return fmt.Errorf("component is nil")
				}
				classification := strings.TrimSpace(value)
				if classification == "" {
					return fmt.Errorf("classification value is empty")
				}
				data := ensureComponentData(tgt.Component)
				data.Classification = classification
				return nil
			},
		},
		{
			Key:      DatasetGovernance,
			Weight:   0.7,
			Required: false,
			Apply: func(src DatasetSource, tgt DatasetTarget) {
				if tgt.Component == nil {
					return
				}

				governance := &cdx.DataGovernance{}
				hasGovernance := false

				// Custodians from author (HF API) or SharedBy/CuratedBy (Readme)
				var custodianName string
				if src.HF != nil && strings.TrimSpace(src.HF.Author) != "" {
					custodianName = strings.TrimSpace(src.HF.Author)
				} else if src.Readme != nil {
					if strings.TrimSpace(src.Readme.SharedBy) != "" {
						custodianName = strings.TrimSpace(src.Readme.SharedBy)
					} else if strings.TrimSpace(src.Readme.CuratedBy) != "" {
						custodianName = strings.TrimSpace(src.Readme.CuratedBy)
					}
				}
				if custodianName != "" {
					governance.Custodians = &[]cdx.ComponentDataGovernanceResponsibleParty{{
						Organization: &cdx.OrganizationalEntity{Name: custodianName},
					}}
					hasGovernance = true
				}

				// Stewards from CuratedBy (Readme)
				if src.Readme != nil && strings.TrimSpace(src.Readme.CuratedBy) != "" {
					governance.Stewards = &[]cdx.ComponentDataGovernanceResponsibleParty{{
						Organization: &cdx.OrganizationalEntity{Name: strings.TrimSpace(src.Readme.CuratedBy)},
					}}
					hasGovernance = true
				}

				// Owners from FundedBy (Readme)
				if src.Readme != nil && strings.TrimSpace(src.Readme.FundedBy) != "" {
					governance.Owners = &[]cdx.ComponentDataGovernanceResponsibleParty{{
						Organization: &cdx.OrganizationalEntity{Name: strings.TrimSpace(src.Readme.FundedBy)},
					}}
					hasGovernance = true
				}

				if hasGovernance {
					data := ensureComponentData(tgt.Component)
					data.Governance = governance
					logf(src.DatasetID, "apply %s set governance", DatasetGovernance)
				}
			},
			Present: func(comp *cdx.Component) bool {
				data := getComponentData(comp)
				return data != nil && data.Governance != nil
			},
			SetUserValue: func(value string, tgt DatasetTarget) error {
				// Governance is complex, not typically set by simple string
				return nil
			},
		},
		{
			Key:      DatasetHashes,
			Weight:   0.5,
			Required: false,
			Apply: func(src DatasetSource, tgt DatasetTarget) {
				if tgt.Component == nil || src.HF == nil {
					return
				}

				sha := strings.TrimSpace(src.HF.SHA)
				if sha == "" {
					return
				}

				hashes := []cdx.Hash{{
					Algorithm: cdx.HashAlgoSHA1,
					Value:     sha,
				}}
				tgt.Component.Hashes = &hashes
				logf(src.DatasetID, "apply %s set=%s", DatasetHashes, sha)
			},
			Present: func(comp *cdx.Component) bool {
				return comp != nil && comp.Hashes != nil && len(*comp.Hashes) > 0
			},
			SetUserValue: func(value string, tgt DatasetTarget) error {
				if tgt.Component == nil {
					return fmt.Errorf("component is nil")
				}
				hash := strings.TrimSpace(value)
				if hash == "" {
					return fmt.Errorf("hash value is empty")
				}
				hashes := []cdx.Hash{{
					Algorithm: cdx.HashAlgoSHA1,
					Value:     hash,
				}}
				tgt.Component.Hashes = &hashes
				return nil
			},
		},
		{
			Key:      DatasetCreatedAt,
			Weight:   0.3,
			Required: false,
			Apply: func(src DatasetSource, tgt DatasetTarget) {
				if tgt.Component == nil || src.HF == nil {
					return
				}

				createdAt := strings.TrimSpace(src.HF.CreatedAt)
				if createdAt == "" {
					return
				}

				setProperty(tgt.Component, "createdAt", createdAt)
				logf(src.DatasetID, "apply %s set=%s", DatasetCreatedAt, createdAt)
			},
			Present: func(comp *cdx.Component) bool {
				return hasProperty(comp, "createdAt")
			},
			SetUserValue: func(value string, tgt DatasetTarget) error {
				if tgt.Component == nil {
					return fmt.Errorf("component is nil")
				}
				setProperty(tgt.Component, "createdAt", strings.TrimSpace(value))
				return nil
			},
		},
		{
			Key:      DatasetUsedStorage,
			Weight:   0.3,
			Required: false,
			Apply: func(src DatasetSource, tgt DatasetTarget) {
				if tgt.Component == nil || src.HF == nil {
					return
				}

				if src.HF.UsedStorage <= 0 {
					return
				}

				setProperty(tgt.Component, "usedStorage", fmt.Sprintf("%d", src.HF.UsedStorage))
				logf(src.DatasetID, "apply %s set=%d", DatasetUsedStorage, src.HF.UsedStorage)
			},
			Present: func(comp *cdx.Component) bool {
				return hasProperty(comp, "usedStorage")
			},
			SetUserValue: func(value string, tgt DatasetTarget) error {
				if tgt.Component == nil {
					return fmt.Errorf("component is nil")
				}
				setProperty(tgt.Component, "usedStorage", strings.TrimSpace(value))
				return nil
			},
		},
		{
			Key:      DatasetLastModified,
			Weight:   0.3,
			Required: false,
			Apply: func(src DatasetSource, tgt DatasetTarget) {
				if tgt.Component == nil || src.HF == nil {
					return
				}

				lastMod := strings.TrimSpace(src.HF.LastMod)
				if lastMod == "" {
					return
				}

				// Add to tags for tracking
				if tgt.Component.Tags != nil {
					tags := *tgt.Component.Tags
					tags = append(tags, "lastModified:"+lastMod)
					tgt.Component.Tags = &tags
				} else {
					tags := []string{"lastModified:" + lastMod}
					tgt.Component.Tags = &tags
				}
				logf(src.DatasetID, "apply %s set=%s", DatasetLastModified, lastMod)
			},
			Present: func(comp *cdx.Component) bool {
				if comp == nil || comp.Tags == nil {
					return false
				}
				for _, tag := range *comp.Tags {
					if strings.HasPrefix(tag, "lastModified:") {
						return true
					}
				}
				return false
			},
			SetUserValue: func(value string, tgt DatasetTarget) error {
				if tgt.Component == nil {
					return fmt.Errorf("component is nil")
				}
				lastMod := strings.TrimSpace(value)
				if lastMod == "" {
					return fmt.Errorf("lastModified value is empty")
				}
				if tgt.Component.Tags != nil {
					tags := *tgt.Component.Tags
					tags = append(tags, "lastModified:"+lastMod)
					tgt.Component.Tags = &tags
				} else {
					tags := []string{"lastModified:" + lastMod}
					tgt.Component.Tags = &tags
				}
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
