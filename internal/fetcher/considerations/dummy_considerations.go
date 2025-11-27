package considerations

import cdx "github.com/CycloneDX/cyclonedx-go"

// DummyConsiderationsFetcher provides placeholder considerations when needed for testing/demo.
type DummyConsiderationsFetcher struct{}

func NewDummyConsiderationsFetcher() *DummyConsiderationsFetcher {
	return &DummyConsiderationsFetcher{}
}

func (s *DummyConsiderationsFetcher) Get(id string) (*cdx.MLModelCardConsiderations, error) {
	users := []string{"NLP researchers", "Developers"}
	useCases := []string{"Sentiment analysis", "Intent classification"}
	techLim := []string{"Not suitable for non-English text"}
	tradeoffs := []string{"Large model size increases inference time"}
	ethics := []cdx.MLModelCardEthicalConsideration{{
		Name:               "Bias in training data",
		MitigationStrategy: "Careful dataset curation",
	}}
	return &cdx.MLModelCardConsiderations{
		Users:                 &users,
		UseCases:              &useCases,
		TechnicalLimitations:  &techLim,
		PerformanceTradeoffs:  &tradeoffs,
		EthicalConsiderations: &ethics,
	}, nil
}
