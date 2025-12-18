package builder

import (
	"strings"

	"aibomgen-cra/internal/metadata"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

type BOMBuilder struct {
	MetaComp MetadataComponentBuilder
	Opts     Options
}

func NewBOMBuilder(opts Options) *BOMBuilder {
	return &BOMBuilder{MetaComp: NewMetadataComponentBuilder(opts), Opts: opts}
}

func (b BOMBuilder) Build(ctx BuildContext) (*cdx.BOM, error) {
	logf(ctx.ModelID, "build start")

	comp, err := b.MetaComp.Build(ctx)
	if err != nil {
		logf(ctx.ModelID, "metadata component build failed (%v)", err)
		return nil, err
	}

	bom := cdx.NewBOM()
	bom.Metadata = &cdx.Metadata{Component: comp}

	// Apply registry exactly once (no duplication)
	src := metadata.Source{
		ModelID: strings.TrimSpace(ctx.ModelID),
		Scan:    ctx.Scan,
		HF:      ctx.HF,
	}
	tgt := metadata.Target{
		BOM:                       bom,
		Component:                 comp,
		ModelCard:                 comp.ModelCard,
		IncludeEvidenceProperties: b.Opts.IncludeEvidenceProperties,
		HuggingFaceBaseURL:        b.Opts.HuggingFaceBaseURL,
	}

	for _, spec := range metadata.Registry() {
		if spec.Apply != nil {
			spec.Apply(src, tgt)
		}
	}

	// Optional cleanup: drop empty ModelParameters
	pruneEmptyModelParameters(comp)

	logf(ctx.ModelID, "build ok")
	return bom, nil
}

func pruneEmptyModelParameters(comp *cdx.Component) {
	if comp == nil || comp.ModelCard == nil || comp.ModelCard.ModelParameters == nil {
		return
	}
	mp := comp.ModelCard.ModelParameters
	emptyDatasets := mp.Datasets == nil || len(*mp.Datasets) == 0
	if strings.TrimSpace(mp.Task) == "" &&
		strings.TrimSpace(mp.ArchitectureFamily) == "" &&
		strings.TrimSpace(mp.ModelArchitecture) == "" &&
		emptyDatasets {
		comp.ModelCard.ModelParameters = nil
	}
}
