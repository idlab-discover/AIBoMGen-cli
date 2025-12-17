package builder

import cdx "github.com/CycloneDX/cyclonedx-go"

type BOMBuilder struct {
	MetaComp MetadataComponentBuilder
	Opts     Options
}

func NewBOMBuilder(opts Options) *BOMBuilder {
	return &BOMBuilder{MetaComp: NewMetadataComponentBuilder(opts), Opts: opts}
}

func (b BOMBuilder) Build(ctx BuildContext) (*cdx.BOM, error) {
	logf(ctx.ModelID, "build start")
	metaComp, err := b.MetaComp.Build(ctx)
	if err != nil {
		logf(ctx.ModelID, "metadata component build failed (%v)", err)
		return nil, err
	}

	bom := cdx.NewBOM()
	bom.Metadata = &cdx.Metadata{Component: metaComp}
	logf(ctx.ModelID, "build ok")
	return bom, nil
}
