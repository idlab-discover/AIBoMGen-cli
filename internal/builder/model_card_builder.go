package builder

import cdx "github.com/CycloneDX/cyclonedx-go"

type ModelCardBuilder struct{}

func (b ModelCardBuilder) Build(ctx BuildContext) (*cdx.MLModelCard, error) {
	logf(ctx.ModelID, "model card start")
	card := &cdx.MLModelCard{}
	logf(ctx.ModelID, "model card ok")
	return card, nil
}
