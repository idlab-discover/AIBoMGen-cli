package quantitative

import cdx "github.com/CycloneDX/cyclonedx-go"

// DummyQuantFetcher provides placeholder quantitative metrics for testing/demo.
type DummyQuantFetcher struct{}

func NewDummyQuantFetcher() *DummyQuantFetcher { return &DummyQuantFetcher{} }

func (s *DummyQuantFetcher) Get(id string) (*cdx.MLQuantitativeAnalysis, error) {
	perf := []cdx.MLPerformanceMetric{{
		Type:  "accuracy",
		Value: "0.84",
		Slice: "dev",
		ConfidenceInterval: &cdx.MLPerformanceMetricConfidenceInterval{
			LowerBound: "0.82",
			UpperBound: "0.86",
		},
	}}
	return &cdx.MLQuantitativeAnalysis{PerformanceMetrics: &perf}, nil
}
