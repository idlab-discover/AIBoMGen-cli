package builder

import (
	"strings"

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

	card, err := b.Card.Build(ctx)
	if err != nil {
		logf(ctx.ModelID, "model card build failed (%v)", err)
		return nil, err
	}

	// minimal skeleton; registry fills the rest
	name := strings.TrimSpace(ctx.ModelID)
	if name == "" && strings.TrimSpace(ctx.Scan.Name) != "" {
		name = strings.TrimSpace(ctx.Scan.Name)
	}
	if name == "" {
		name = "model"
	}

	comp := &cdx.Component{
		Type:      cdx.ComponentTypeMachineLearningModel,
		Name:      name,
		ModelCard: card,
	}

	logf(ctx.ModelID, "metadata component ok (name=%q)", strings.TrimSpace(comp.Name))
	return comp, nil
}
