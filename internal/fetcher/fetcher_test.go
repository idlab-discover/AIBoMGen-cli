package fetcher

import "testing"

func TestFetchModelCard_EmptyWhenNoDefault(t *testing.T) {
	// No Default set; expect an empty card
	SetDefault(nil)
	card := FetchModelCard("bert-base-uncased")
	if card == nil {
		t.Fatalf("expected non-nil MLModelCard")
	}
}
