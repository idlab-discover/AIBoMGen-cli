package completeness

import (
	"testing"

	"github.com/idlab-discover/AIBoMGen-cli/internal/metadata"
)

func TestJoinKeys_Empty(t *testing.T) {
	if got := joinKeys(nil); got != "" {
		t.Fatalf("joinKeys(nil) = %q, want empty", got)
	}
	if got := joinKeys([]metadata.Key{}); got != "" {
		t.Fatalf("joinKeys(empty) = %q, want empty", got)
	}
}
