package fetcher

import (
	"testing"
)

func TestFetchModelCard(t *testing.T) {
	card := FetchModelCard("bert-base-uncased")
	if card == nil {
		t.Fatalf("expected non-nil MLModelCard")
	}
	if card.ModelParameters == nil {
		t.Fatalf("expected ModelParameters to be set")
	}
	if card.ModelParameters.Task == "" {
		t.Errorf("expected Task to be non-empty")
	}
	if card.QuantitativeAnalysis == nil || card.QuantitativeAnalysis.PerformanceMetrics == nil {
		t.Fatalf("expected QuantitativeAnalysis with PerformanceMetrics")
	}
	if len(*card.QuantitativeAnalysis.PerformanceMetrics) == 0 {
		t.Errorf("expected at least one performance metric")
	}
	if card.Considerations == nil || card.Considerations.Users == nil {
		t.Fatalf("expected Considerations with Users")
	}
	if len(*card.Considerations.Users) == 0 {
		t.Errorf("expected at least one user in considerations")
	}
}
