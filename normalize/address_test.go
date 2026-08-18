package normalize_test

import (
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/normalize"
)

func TestFormatStreet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"rua primavera", "Rua Primavera"},
		{"R. Primavera", "Rua Primavera"},
		{"Rua Primavera", "Rua Primavera"},
		{"Rua Rua Primavera", "Rua Primavera"},
		{"av. bento goncalves", "Avenida Bento Goncalves"},
		{"Travessa Independência", "Travessa Independência"},
		{"rua 25 de julho", "Rua 25 de Julho"},
		{"walter sander", "Walter Sander"},
		{"RS 244", "RS 244"},
	}
	for _, test := range tests {
		got := normalize.FormatStreet(test.in)
		if got != test.want {
			t.Fatalf("FormatStreet(%q) = %q, want %q", test.in, got, test.want)
		}
		if again := normalize.FormatStreet(got); again != got {
			t.Fatalf("FormatStreet not idempotent: %q -> %q", got, again)
		}
	}
}

func TestFormatHouseNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"123", "123"},
		{"nº 123", "123"},
		{"123-A", "123A"},
		{"8e", "8E"},
		{"s/n", ""},
		{"sem numero", ""},
	}
	for _, test := range tests {
		got := normalize.FormatHouseNumber(test.in)
		if got != test.want {
			t.Fatalf("FormatHouseNumber(%q) = %q, want %q", test.in, got, test.want)
		}
		if got != "" {
			if again := normalize.FormatHouseNumber(got); again != got {
				t.Fatalf("FormatHouseNumber not idempotent: %q -> %q", got, again)
			}
		}
	}
}

func TestFormatBlock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"A", "bl. A"},
		{"bloco A", "bl. A"},
		{"bl. A", "bl. A"},
		{"blc 2", "bl. 2"},
		{"bloco", ""},
	}
	for _, test := range tests {
		got := normalize.FormatBlock(test.in)
		if got != test.want {
			t.Fatalf("FormatBlock(%q) = %q, want %q", test.in, got, test.want)
		}
		if got != "" {
			if again := normalize.FormatBlock(got); again != got {
				t.Fatalf("FormatBlock not idempotent: %q -> %q", got, again)
			}
		}
	}
}

func TestFormatApartment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"202", "apt. 202"},
		{"apto 202", "apt. 202"},
		{"ap 7B", "apt. 7B"},
		{"apt. 202", "apt. 202"},
		{"aptº 101", "apt. 101"},
		{"apartamento", ""},
	}
	for _, test := range tests {
		got := normalize.FormatApartment(test.in)
		if got != test.want {
			t.Fatalf("FormatApartment(%q) = %q, want %q", test.in, got, test.want)
		}
		if got != "" {
			if again := normalize.FormatApartment(got); again != got {
				t.Fatalf("FormatApartment not idempotent: %q -> %q", got, again)
			}
		}
	}
}
