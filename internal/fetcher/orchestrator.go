package fetcher

import (
	cdx "github.com/CycloneDX/cyclonedx-go"
)

// ParametersFetcher specializes in fetching MLModelParameters
type ParametersFetcher interface {
	Get(id string) (*cdx.MLModelParameters, error)
}

// QuantitativeFetcher specializes in fetching MLQuantitativeAnalysis
type QuantitativeFetcher interface {
	Get(id string) (*cdx.MLQuantitativeAnalysis, error)
}

// ConsiderationsFetcher specializes in fetching MLModelCardConsiderations
type ConsiderationsFetcher interface {
	Get(id string) (*cdx.MLModelCardConsiderations, error)
}

// Orchestrator composes section fetchers to produce a full MLModelCard
type Orchestrator struct {
	P ParametersFetcher
	Q QuantitativeFetcher
	C ConsiderationsFetcher
}

func NewOrchestrator(p ParametersFetcher, q QuantitativeFetcher, c ConsiderationsFetcher) *Orchestrator {
	return &Orchestrator{P: p, Q: q, C: c}
}

func (o *Orchestrator) Get(id string) (*cdx.MLModelCard, error) {
	card := &cdx.MLModelCard{}
	if o.P != nil {
		if mp, err := o.P.Get(id); err == nil && mp != nil {
			card.ModelParameters = mp
		}
	}
	if o.Q != nil {
		if qa, err := o.Q.Get(id); err == nil && qa != nil {
			card.QuantitativeAnalysis = qa
		}
	}
	if o.C != nil {
		if cc, err := o.C.Get(id); err == nil && cc != nil {
			card.Considerations = cc
		}
	}
	return card, nil
}
