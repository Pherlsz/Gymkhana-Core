package normalize_test

import (
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/normalize"
)

func TestCanonicalLooseDocumentPreservesCanonicalLabelLikePrefix(t *testing.T) {
	t.Parallel()

	const canonical = "OAB0"

	got, err := normalize.CanonicalDocument(normalize.DocumentStudentID, canonical)
	if err != nil {
		t.Fatalf("CanonicalDocument() error = %v", err)
	}
	if got != canonical {
		t.Fatalf("CanonicalDocument() = %q, want %q", got, canonical)
	}

	again, err := normalize.CanonicalDocument(normalize.DocumentStudentID, got)
	if err != nil {
		t.Fatalf("CanonicalDocument(canonical) error = %v", err)
	}
	if again != canonical {
		t.Fatalf("CanonicalDocument(canonical) = %q, want %q", again, canonical)
	}
}
