package normalize_test

import (
	"testing"
	"unicode/utf8"

	"github.com/Pherlsz/Gymkhana-Core/normalize"
)

func FuzzSearchText(f *testing.F) {
	for _, seed := range []string{
		"São João",
		"A---B",
		"  001  ",
		"\xff",
		"0000000000000000000000000000000000000000000000000000000\xf1º",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, value string) {
		if len(value) > 4096 {
			return
		}

		got := normalize.SearchText(value)
		if !utf8.ValidString(got) {
			t.Fatalf("SearchText returned invalid UTF-8")
		}
		if normalize.SearchText(got) != got {
			t.Fatalf("SearchText is not idempotent for %q", got)
		}
	})
}

func FuzzCanonicalCPF(f *testing.F) {
	for _, seed := range []string{"529.982.247-25", "11111111111", "", "abc"} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, value string) {
		if len(value) > 256 {
			return
		}

		canonical, err := normalize.CanonicalCPF(value)
		if err == nil && len(canonical) != 11 {
			t.Fatalf("successful CPF has length %d", len(canonical))
		}
	})
}
