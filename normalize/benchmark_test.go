package normalize_test

import (
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/normalize"
)

func BenchmarkSearchText(b *testing.B) {
	const input = "  São-João da Barra / Nº 42 - Centro  "
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = normalize.SearchText(input)
	}
}

func BenchmarkCanonicalCPF(b *testing.B) {
	const input = "529.982.247-25"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = normalize.CanonicalCPF(input)
	}
}

func BenchmarkCanonicalCNPJAlphanumeric(b *testing.B) {
	const input = "12.abc.345/01de-35"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = normalize.CanonicalCNPJ(input)
	}
}

func BenchmarkFormatCityExact(b *testing.B) {
	const input = "porto alegre rs"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = normalize.FormatCity(input)
	}
}

func BenchmarkFormatCityFuzzy(b *testing.B) {
	const input = "charqueadass"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = normalize.FormatCity(input)
	}
}

func BenchmarkFormatBlock(b *testing.B) {
	const input = "bloco A"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = normalize.FormatBlock(input)
	}
}

func BenchmarkFormatApartment(b *testing.B) {
	const input = "apto 202"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = normalize.FormatApartment(input)
	}
}

func BenchmarkIdentifyDocument(b *testing.B) {
	const input = "OAB/RS 114.458"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = normalize.IdentifyDocument(input)
	}
}
