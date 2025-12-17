package builder

import (
	"fmt"
	"strings"

	"aibomgen-cra/internal/metadata"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

type MetadataComponentBuilder struct {
	Card ModelCardBuilder
	Opts Options
}

func NewMetadataComponentBuilder(opts Options) MetadataComponentBuilder {
	return MetadataComponentBuilder{Card: ModelCardBuilder{}, Opts: opts}
}

func (b MetadataComponentBuilder) Build(ctx BuildContext) (*cdx.Component, error) {
	logf(ctx.ModelID, "metadata component start")
	name := strings.TrimSpace(ctx.ModelID)
	if s, ok := ctx.Meta.String(metadata.ComponentName); ok {
		name = s
	} else if strings.TrimSpace(ctx.Scan.Name) != "" {
		name = strings.TrimSpace(ctx.Scan.Name)
	}

	card, err := b.Card.Build(ctx)
	if err != nil {
		logf(ctx.ModelID, "model card build failed (%v)", err)
		return nil, err
	}
	logf(ctx.ModelID, "model card built")

	comp := &cdx.Component{
		Type:      cdx.ComponentTypeMachineLearningModel,
		Name:      name,
		ModelCard: card,
	}

	// external references from metadata (expects []cdx.ExternalReference)
	if raw, ok := ctx.Meta.Raw(metadata.ComponentExternalReferences); ok {
		if refs, ok := raw.([]cdx.ExternalReference); ok && len(refs) > 0 {
			cp := append([]cdx.ExternalReference(nil), refs...)
			comp.ExternalReferences = &cp
		}
	}
	// fallback website
	if comp.ExternalReferences == nil && strings.TrimSpace(ctx.ModelID) != "" {
		base := strings.TrimSpace(b.Opts.HuggingFaceBaseURL)
		if base == "" {
			base = "https://huggingface.co/"
		}
		if !strings.HasSuffix(base, "/") {
			base += "/"
		}
		comp.ExternalReferences = &[]cdx.ExternalReference{{
			Type: cdx.ExternalReferenceType("website"),
			URL:  base + strings.TrimPrefix(ctx.ModelID, "/"),
		}}
	}

	// tags
	if tags, ok := ctx.Meta.Strings(metadata.ComponentTags); ok && len(tags) > 0 {
		cp := append([]string(nil), tags...)
		comp.Tags = &cp
	}

	// licenses (store HF license as license.name; do not assume SPDX)
	if ids, ok := ctx.Meta.Strings(metadata.ComponentLicenses); ok && len(ids) > 0 {
		ls := make(cdx.Licenses, 0, len(ids))
		for _, raw := range ids {
			raw = strings.TrimSpace(raw)
			if raw == "" {
				continue
			}
			ls = append(ls, cdx.LicenseChoice{License: &cdx.License{Name: raw}})
		}
		if len(ls) > 0 {
			comp.Licenses = &ls
		}
	}

	// hashes (expects []cdx.Hash)
	if raw, ok := ctx.Meta.Raw(metadata.ComponentHashes); ok {
		if hs, ok := raw.([]cdx.Hash); ok && len(hs) > 0 {
			cp := append([]cdx.Hash(nil), hs...)
			comp.Hashes = &cp
		}
	}

	// manufacturer/group
	if s, ok := ctx.Meta.String(metadata.ComponentManufacturer); ok {
		comp.Manufacturer = &cdx.OrganizationalEntity{Name: s}
	}
	if s, ok := ctx.Meta.String(metadata.ComponentGroup); ok {
		comp.Group = s
	}

	// properties (include scan evidence + HF properties)
	props := []cdx.Property{}
	if b.Opts.IncludeEvidenceProperties {
		props = append(props,
			cdx.Property{Name: "aibomgen.type", Value: ctx.Scan.Type},
			cdx.Property{Name: "aibomgen.evidence", Value: ctx.Scan.Evidence},
			cdx.Property{Name: "aibomgen.path", Value: ctx.Scan.Path},
		)
	}

	addProp := func(k metadata.Key) {
		if raw, ok := ctx.Meta.Raw(k); ok && raw != nil {
			props = append(props, cdx.Property{Name: k.String(), Value: strings.TrimSpace(fmt.Sprint(raw))})
		}
	}
	addProp(metadata.ComponentPropertiesHuggingFaceLastModified)
	addProp(metadata.ComponentPropertiesHuggingFaceCreatedAt)
	addProp(metadata.ComponentPropertiesHuggingFaceLanguage)
	addProp(metadata.ComponentPropertiesHuggingFaceUsedStorage)
	addProp(metadata.ComponentPropertiesHuggingFacePrivate)
	addProp(metadata.ComponentPropertiesHuggingFaceLibraryName)
	addProp(metadata.ComponentPropertiesHuggingFaceDownloads)
	addProp(metadata.ComponentPropertiesHuggingFaceLikes)

	if len(props) > 0 {
		comp.Properties = &props
	}
	logf(ctx.ModelID, "metadata component ok (name=%q)", strings.TrimSpace(comp.Name))
	return comp, nil
}
