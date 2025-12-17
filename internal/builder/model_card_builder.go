package builder

import (
	"aibomgen-cra/internal/metadata"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

type ModelCardBuilder struct{}

func (b ModelCardBuilder) Build(ctx BuildContext) (*cdx.MLModelCard, error) {
	logf(ctx.ModelID, "model card start")
	card := &cdx.MLModelCard{}
	mp := &cdx.MLModelParameters{}

	if s, ok := ctx.Meta.String(metadata.ModelCardModelParametersTask); ok {
		mp.Task = s
	}
	if s, ok := ctx.Meta.String(metadata.ModelCardModelParametersArchitectureFamily); ok {
		mp.ArchitectureFamily = s
	}
	if s, ok := ctx.Meta.String(metadata.ModelCardModelParametersModelArchitecture); ok {
		mp.ModelArchitecture = s
	}
	if ds, ok := ctx.Meta.Strings(metadata.ModelCardModelParametersDatasets); ok && len(ds) > 0 {
		choices := make([]cdx.MLDatasetChoice, 0, len(ds))
		for _, ref := range ds {
			choices = append(choices, cdx.MLDatasetChoice{Ref: ref})
		}
		mp.Datasets = &choices
	}

	// attach only if something is set
	if mp.Task != "" || mp.ArchitectureFamily != "" || mp.ModelArchitecture != "" || (mp.Datasets != nil && len(*mp.Datasets) > 0) {
		card.ModelParameters = mp
	}
	logf(ctx.ModelID, "model card ok")
	return card, nil
}
