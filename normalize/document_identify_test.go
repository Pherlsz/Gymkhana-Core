package normalize_test

import (
	"fmt"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/normalize"
)

func TestIdentifyDocumentFromBareNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input     string
		kind      normalize.DocumentKind
		canonical string
	}{
		{"529.982.247-25", normalize.DocumentCPF, "52998224725"},
		{"11.222.333/0001-81", normalize.DocumentCNPJ, "11222333000181"},
		{"12.ABC.345/01DE-35", normalize.DocumentCNPJ, "12ABC34501DE35"},
		{"02172931125", normalize.DocumentCNH, "02172931125"},
		{"1234 5678 0493", normalize.DocumentVoterID, "123456780493"},
		{"700000000000005", normalize.DocumentSUSCard, "700000000000005"},
		{"ab123456", normalize.DocumentPassport, "AB123456"},
		{"12345678-X", normalize.DocumentRG, "12345678X"},
		{"12345678X/RS", normalize.DocumentRG, "12345678X/RS"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.input, func(t *testing.T) {
			t.Parallel()
			got, err := normalize.IdentifyDocument(test.input)
			if err != nil {
				t.Fatalf("IdentifyDocument() error = %v", err)
			}
			if got.Kind != test.kind || got.Canonical != test.canonical {
				t.Fatalf("IdentifyDocument() = %+v, want kind %s canonical %q", got, test.kind, test.canonical)
			}
		})
	}
}

func TestIdentifyDocumentUsesLabelAndDoesNotReassign(t *testing.T) {
	t.Parallel()

	got, err := normalize.IdentifyDocument("OAB/RS 114.458")
	if err != nil {
		t.Fatalf("IdentifyDocument() error = %v", err)
	}
	if got.Kind != normalize.DocumentOAB || got.Canonical != "114458/RS" {
		t.Fatalf("IdentifyDocument() = %+v", got)
	}

	_, err = normalize.IdentifyDocument("CPF 02172931125")
	if !normalize.IsCode(err, normalize.CodeInvalidChecksum) {
		t.Fatalf("labeled invalid CPF should not become CNH, got %v", err)
	}

	got, err = normalize.IdentifyDocument("PIS 120.5824.883-1")
	if err != nil {
		t.Fatalf("labeled PIS error = %v", err)
	}
	if got.Kind != normalize.DocumentPIS || got.Canonical != "12058248831" {
		t.Fatalf("labeled PIS = %+v", got)
	}
}

func TestIdentifyDocumentRejectsConflictingExplicitLabels(t *testing.T) {
	t.Parallel()

	const input = "CPF CNH 529.982.247-25"
	_, err := normalize.IdentifyDocument(input)
	if !normalize.IsCode(err, normalize.CodeInvalidFormat) {
		t.Fatalf("IdentifyDocument(%q) error = %v, want invalid_format", input, err)
	}
	if matches := normalize.IdentifyDocumentMatches(input); len(matches) != 0 {
		t.Fatalf("IdentifyDocumentMatches(%q) = %#v, want no materialized matches", input, matches)
	}
}

func TestIdentifyDocumentRejectsWeakBareNumbers(t *testing.T) {
	t.Parallel()

	for _, input := range []string{
		"12345678",
		"114458/RS",
		"0012345",
		"182265",
		"1122334455",
		"12058248831",
		"nao",
		"sim",
	} {
		input := input
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			_, err := normalize.IdentifyDocument(input)
			if !normalize.IsCode(err, normalize.CodeUnknownKind) {
				t.Fatalf("IdentifyDocument(%q) error = %v, want unknown_kind", input, err)
			}
			if matches := normalize.IdentifyDocumentMatches(input); len(matches) != 0 {
				t.Fatalf("IdentifyDocumentMatches(%q) = %#v", input, matches)
			}
		})
	}
}

func TestIdentifyDocumentAmbiguousWhenCPFAndCNHBothMatch(t *testing.T) {
	t.Parallel()

	collision := findCPFCNHCollision()
	if collision == "" {
		t.Fatal("did not find a CPF/CNH checksum collision in the search window")
	}

	_, err := normalize.IdentifyDocument(collision)
	if !normalize.IsCode(err, normalize.CodeAmbiguous) {
		t.Fatalf("IdentifyDocument(%s) error = %v, want ambiguous", collision, err)
	}
	matches := normalize.IdentifyDocumentMatches(collision)
	if len(matches) != 2 || matches[0].Kind != normalize.DocumentCPF || matches[1].Kind != normalize.DocumentCNH {
		t.Fatalf("IdentifyDocumentMatches() = %#v", matches)
	}
}

func findCPFCNHCollision() string {
	for n := 0; n < 200000; n++ {
		base := fmt.Sprintf("%09d", n)
		for dv := 0; dv < 100; dv++ {
			candidate := base + fmt.Sprintf("%02d", dv)
			_, cpfErr := normalize.CanonicalDocument(normalize.DocumentCPF, candidate)
			if cpfErr != nil {
				continue
			}
			_, cnhErr := normalize.CanonicalDocument(normalize.DocumentCNH, candidate)
			if cnhErr == nil {
				return candidate
			}
			break
		}
	}
	return ""
}
