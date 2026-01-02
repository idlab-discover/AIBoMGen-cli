package builder

import (
	"strings"

	"github.com/idlab-discover/AIBoMGen-cli/internal/metadata"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

type BOMBuilder struct {
	Opts Options
}

func NewBOMBuilder(opts Options) *BOMBuilder {
	return &BOMBuilder{Opts: opts}
}

func (b BOMBuilder) Build(ctx BuildContext) (*cdx.BOM, error) {
	logf(ctx.ModelID, "build start")

	comp := buildMetadataComponent(ctx)

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

func buildMetadataComponent(ctx BuildContext) *cdx.Component {
	// Minimal skeleton; registry fills the rest
	name := strings.TrimSpace(ctx.ModelID)
	if name == "" && strings.TrimSpace(ctx.Scan.Name) != "" {
		name = strings.TrimSpace(ctx.Scan.Name)
	}
	if name == "" {
		name = "model"
	}

	return &cdx.Component{
		Type:      cdx.ComponentTypeMachineLearningModel,
		Name:      name,
		ModelCard: &cdx.MLModelCard{},
	}
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
