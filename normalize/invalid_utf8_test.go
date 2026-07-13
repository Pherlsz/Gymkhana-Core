package normalize_test

import (
	"testing"
	"unicode/utf8"

	"github.com/Pherlsz/Gymkhana-Core/normalize"
)

func TestSearchTextMalformedUTF8IsStable(t *testing.T) {
	t.Parallel()

	input := "0000000000000000000000000000000000000000000000000000000\xf1º"
	canonical := normalize.SearchText(input)

	if !utf8.ValidString(canonical) {
		t.Fatal("SearchText returned invalid UTF-8")
	}
	if repeated := normalize.SearchText(canonical); repeated != canonical {
		t.Fatalf("SearchText is not idempotent: first %q, second %q", canonical, repeated)
	}
}
