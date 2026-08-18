package normalize_test

import (
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/normalize"
)

func TestDisplayText(t *testing.T) {
	t.Parallel()

	got := normalize.DisplayText("  João\t da\nSilva  ")
	if got != "João da Silva" {
		t.Fatalf("DisplayText() = %q", got)
	}
}

func TestSearchText(t *testing.T) {
	t.Parallel()

	got := normalize.SearchText("  São-João / Nº 42  ")
	if got != "sao joao no 42" {
		t.Fatalf("SearchText() = %q", got)
	}
}

func TestAlphanumericPreservesLeadingZeros(t *testing.T) {
	t.Parallel()

	got := normalize.Alphanumeric(" 00-ab.ç9 ")
	if got != "00ABÇ9" {
		t.Fatalf("Alphanumeric() = %q", got)
	}
}

func TestCanonicalCPF(t *testing.T) {
	t.Parallel()

	got, err := normalize.CanonicalCPF("529.982.247-25")
	if err != nil {
		t.Fatalf("CanonicalCPF() error = %v", err)
	}
	if got != "52998224725" {
		t.Fatalf("CanonicalCPF() = %q", got)
	}

	if _, err := normalize.CanonicalCPF("111.111.111-11"); !normalize.IsCode(err, normalize.CodeInvalidChecksum) {
		t.Fatalf("expected checksum error, got %v", err)
	}
}

func TestMaskCPF(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"52998224725", "***.982.247-25"},
		{"", ""},
		{"1234567890", "1234567890"},         // too short
		{"123456789012", "123456789012"},     // too long
		{"5299822472a", "5299822472a"},       // non-digit
		{"529.982.247-25", "529.982.247-25"}, // formatted, not canonical
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got := normalize.MaskCPF(tc.input)
			if got != tc.expected {
				t.Fatalf("MaskCPF(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestCanonicalEmail(t *testing.T) {
	t.Parallel()

	got, err := normalize.CanonicalEmail("User.Name@EXAMPLE.COM")
	if err != nil {
		t.Fatalf("CanonicalEmail() error = %v", err)
	}
	if got != "User.Name@example.com" {
		t.Fatalf("CanonicalEmail() = %q", got)
	}
}

func TestCanonicalBrazilPhone(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"(51) 99999-1234":  "+5551999991234",
		"+55 51 3333-1234": "+555133331234",
	}

	for input, expected := range tests {
		input, expected := input, expected
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			got, err := normalize.CanonicalBrazilPhone(input)
			if err != nil {
				t.Fatalf("CanonicalBrazilPhone() error = %v", err)
			}
			if got != expected {
				t.Fatalf("CanonicalBrazilPhone() = %q", got)
			}
		})
	}
}
