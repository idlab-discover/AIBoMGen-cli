package fetcher

import (
	"net/http"
	"time"
)

// NewHFClient creates an *http.Client configured for Hugging Face API calls.
// The timeout is the per-request deadline; pass 0 for no timeout.
func NewHFClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}
