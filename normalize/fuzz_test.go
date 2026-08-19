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

func FuzzCanonicalCNPJ(f *testing.F) {
	for _, seed := range []string{"11.222.333/0001-81", "12.ABC.345/01DE-35", "00000000000000", "", "abc"} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, value string) {
		if len(value) > 256 {
			return
		}

		canonical, err := normalize.CanonicalCNPJ(value)
		if err != nil {
			return
		}
		if len(canonical) != 14 {
			t.Fatalf("successful CNPJ has length %d", len(canonical))
		}
		if canonical[12] < '0' || canonical[12] > '9' || canonical[13] < '0' || canonical[13] > '9' {
			t.Fatalf("successful CNPJ check digits are not numeric: %q", canonical)
		}
		again, againErr := normalize.CanonicalCNPJ(canonical)
		if againErr != nil || again != canonical {
			t.Fatalf("CanonicalCNPJ is not idempotent: %q -> %q, %v", canonical, again, againErr)
		}
	})
}

func FuzzCanonicalDocument(f *testing.F) {
	for _, seed := range []string{
		"529.982.247-25",
		"OAB/RS 114.458",
		"02172931125",
		"AB123456",
		"nao",
		"12345678-X / RS",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, value string) {
		if len(value) > 256 {
			return
		}

		for _, kind := range normalize.DocumentKinds() {
			canonical, err := normalize.CanonicalDocument(kind, value)
			if err != nil {
				continue
			}
			again, againErr := normalize.CanonicalDocument(kind, canonical)
			if againErr != nil {
				t.Fatalf("%s: canonical %q rejected: %v", kind, canonical, againErr)
			}
			if again != canonical {
				t.Fatalf("%s: CanonicalDocument is not idempotent: %q -> %q", kind, canonical, again)
			}
			display, formatErr := normalize.FormatDocument(kind, canonical)
			if formatErr != nil {
				t.Fatalf("%s: FormatDocument(%q) error = %v", kind, canonical, formatErr)
			}
			fromDisplay, displayErr := normalize.CanonicalDocument(kind, display)
			if displayErr != nil || fromDisplay != canonical {
				t.Fatalf("%s: display %q did not round-trip to %q: %q, %v", kind, display, canonical, fromDisplay, displayErr)
			}
		}
	})
}

func FuzzIdentifyDocument(f *testing.F) {
	for _, seed := range []string{
		"529.982.247-25",
		"02172931125",
		"OAB/RS 114.458",
		"12345678",
		"12345678-X",
		"AB123456",
		"nao",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, value string) {
		if len(value) > 256 {
			return
		}

		match, err := normalize.IdentifyDocument(value)
		matches := normalize.IdentifyDocumentMatches(value)
		if err == nil {
			if len(matches) != 1 || matches[0] != match {
				t.Fatalf("unique identify %v does not match %#v", match, matches)
			}
			canonical, canonicalErr := normalize.CanonicalDocument(match.Kind, match.Canonical)
			if canonicalErr != nil || canonical != match.Canonical {
				t.Fatalf("identified %v is not canonical: %q, %v", match, canonical, canonicalErr)
			}
			return
		}
		if normalize.IsCode(err, normalize.CodeAmbiguous) && len(matches) < 2 {
			t.Fatalf("ambiguous without multiple matches: %#v", matches)
		}
		if normalize.IsCode(err, normalize.CodeUnknownKind) && len(matches) != 0 {
			t.Fatalf("unknown kind with matches %#v", matches)
		}
	})
}

func FuzzFormatAddressSlots(f *testing.F) {
	for _, seed := range []string{
		"rua primavera",
		"bl. A",
		"bl. BL",
		"apt. 202",
		"apt. AP",
		"123-A",
		"",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, value string) {
		if len(value) > 256 {
			return
		}
		street := normalize.FormatStreet(value)
		if normalize.FormatStreet(street) != street {
			t.Fatalf("FormatStreet not idempotent: %q", street)
		}
		block := normalize.FormatBlock(value)
		if normalize.FormatBlock(block) != block {
			t.Fatalf("FormatBlock not idempotent: %q", block)
		}
		apt := normalize.FormatApartment(value)
		if normalize.FormatApartment(apt) != apt {
			t.Fatalf("FormatApartment not idempotent: %q", apt)
		}
		number := normalize.FormatHouseNumber(value)
		if normalize.FormatHouseNumber(number) != number {
			t.Fatalf("FormatHouseNumber not idempotent: %q", number)
		}
	})
}
