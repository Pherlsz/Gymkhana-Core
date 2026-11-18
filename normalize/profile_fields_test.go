package normalize_test

import (
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/normalize"
)

func TestFormatPersonName(t *testing.T) {
	t.Parallel()

	tests := []struct{ in, want string }{
		{"  Ana   da Silva  ", "Ana da Silva"},
		{"---", ""},
		{"9", ""},
		{"A J U R I S", ""},
		{"mae", ""},
		{"ANA DE SOUZA", "Ana de Souza"},
		{"A.p.athanasio & Cia Ltda", ""},
		{"Industria de Calcados Norte", ""},
		{"Foo S.A.", ""},
		{"Joao Sa", "Joao Sa"},
		{"Maria da Silva", "Maria da Silva"},
		{"Comercial Silva", ""},
		{"Padaria do Centro", ""},
		{"Igreja Matriz", ""},
		{"Clube do Povo", ""},
	}
	for _, test := range tests {
		if got := normalize.FormatPersonName(test.in); got != test.want {
			t.Fatalf("FormatPersonName(%q) = %q, want %q", test.in, got, test.want)
		}
	}
}

func TestFormatCivilStatus(t *testing.T) {
	t.Parallel()

	if got := normalize.FormatGender("masculino"); got != "M" {
		t.Fatalf("FormatGender = %q", got)
	}
	if got := normalize.FormatMaritalStatus("uniao estavel"); got != "" {
		t.Fatalf("união estável must not be casado, got %q", got)
	}
	if got := normalize.FormatMaritalStatus("casada"); got != "casado" {
		t.Fatalf("FormatMaritalStatus = %q", got)
	}
	if got := normalize.FormatBloodType("o positivo"); got != "O+" {
		t.Fatalf("FormatBloodType = %q", got)
	}
	if got := normalize.FormatBloodType("O+"); got != "O+" {
		t.Fatalf("FormatBloodType(O+) = %q, want O+", got)
	}
	if got := normalize.FormatBloodType("AB-"); got != "AB-" {
		t.Fatalf("FormatBloodType(AB-) = %q, want AB-", got)
	}
	for in, want := range map[string]string{
		"0":          "O",
		"0+":         "O+",
		"0-":         "O-",
		"0 positivo": "O+",
		"0 negativo": "O-",
		"O−":         "O-",  // Unicode minus U+2212
		"AB–":        "AB-", // En-dash U+2013
		"nao sei":    "",
	} {
		if got := normalize.FormatBloodType(in); got != want {
			t.Fatalf("FormatBloodType(%q) = %q, want %q", in, got, want)
		}
	}
	var typedNil *normalize.ValidationError
	if normalize.IsCode(error(typedNil), "") {
		t.Fatal("IsCode matched typed-nil *ValidationError")
	}
}

func TestFormatGeo(t *testing.T) {
	t.Parallel()

	if got := normalize.FormatUF("rs"); got != "RS" {
		t.Fatalf("FormatUF = %q", got)
	}
	if got := normalize.FormatCEP("90000-000"); got != "90000-000" {
		t.Fatalf("FormatCEP = %q", got)
	}
	canonical, err := normalize.CanonicalCEP("90000-000")
	if err != nil || canonical != "90000000" {
		t.Fatalf("CanonicalCEP = %q, %v", canonical, err)
	}
	if got := normalize.FormatCity("chaqueadas"); got != "Charqueadas" {
		t.Fatalf("FormatCity = %q", got)
	}
	if got := normalize.FormatMunicipality("Arroio do Sal"); got != "Arroio do Sal" {
		t.Fatalf("FormatMunicipality = %q", got)
	}
	if got := normalize.FormatMunicipality("Centro"); got != "" {
		t.Fatalf("Centro must not be a municipality, got %q", got)
	}
	if got := normalize.FormatNeighborhood("bella vista"); got != "Bela Vista" {
		t.Fatalf("FormatNeighborhood = %q", got)
	}
	if got := normalize.FormatNationality("bra"); got != "Brasil" {
		t.Fatalf("FormatNationality = %q", got)
	}
	if got := normalize.FormatNationality("Canada e Mexico"); got != "Canadá; México" {
		t.Fatalf("FormatNationality multi = %q", got)
	}
	// Removed gincana-specific noise "bacanal" from cityHeaderJunk:
	// it is now titled as a regular unknown place rather than dropped as a header.
	if got := normalize.FormatCity("bacanal"); got != "Bacanal" {
		t.Fatalf("FormatCity(bacanal) = %q, want %q", got, "Bacanal")
	}
}

func TestCityCatalogIsolationAndNilReceiver(t *testing.T) {
	t.Parallel()

	var nilCatalog *normalize.CityCatalog
	if got := nilCatalog.FormatCity("porto alegre"); got != "Porto Alegre" {
		t.Fatalf("nilCatalog.FormatCity(porto alegre) = %q, want %q", got, "Porto Alegre")
	}
	if got := nilCatalog.FormatCity("poa"); got != "Poa" {
		t.Fatalf("nilCatalog.FormatCity(poa) = %q, want %q", got, "Poa")
	}
	if got := nilCatalog.FormatCity("cidade"); got != "" {
		t.Fatalf("nilCatalog.FormatCity(cidade) = %q, want empty", got)
	}
	if got := nilCatalog.FormatMunicipality("Porto Alegre"); got != "" {
		t.Fatalf("nilCatalog.FormatMunicipality(Porto Alegre) = %q, want empty", got)
	}

	emptyCatalog := normalize.NewCityCatalog(nil)
	if got := emptyCatalog.FormatMunicipality("Porto Alegre"); got != "" {
		t.Fatalf("emptyCatalog.FormatMunicipality() = %q, want empty", got)
	}

	custom := normalize.NewCityCatalog(map[string]string{
		"sp":        "São Paulo",
		"sao paulo": "São Paulo",
		"campinas":  "Campinas",
	})
	if got := custom.FormatCity("sp"); got != "São Paulo" {
		t.Fatalf("custom.FormatCity(sp) = %q, want São Paulo", got)
	}
	if got := custom.FormatMunicipality("campinas"); got != "Campinas" {
		t.Fatalf("custom.FormatMunicipality(campinas) = %q, want Campinas", got)
	}
	if got := custom.FormatMunicipality("Porto Alegre"); got != "" {
		t.Fatalf("custom.FormatMunicipality(Porto Alegre) = %q, want empty", got)
	}
	// Default catalog remains isolated from custom catalog.
	if got := normalize.FormatMunicipality("Porto Alegre"); got != "Porto Alegre" {
		t.Fatalf("default FormatMunicipality(Porto Alegre) = %q, want Porto Alegre", got)
	}
}

func TestFormatCityTieAmbiguityAndDeterminism(t *testing.T) {
	t.Parallel()

	// Two distinct canonical cities at equal edit distance (d == 1 for "barato"):
	// "barata" -> "Barata" (sub o->a) and "barate" -> "Barate" (sub o->e).
	tiedCatalog := normalize.NewCityCatalog(map[string]string{
		"barata": "Barata",
		"barate": "Barate",
	})
	for i := 0; i < 50; i++ {
		if got := tiedCatalog.FormatCity("barato"); got != "Barato" {
			t.Fatalf("iteration %d: ambiguous tie FormatCity(barato) = %q, want fallback %q", i, got, "Barato")
		}
		if got := tiedCatalog.FormatMunicipality("barato"); got != "" {
			t.Fatalf("iteration %d: ambiguous tie FormatMunicipality(barato) = %q, want empty", i, got)
		}
	}

	// Tie between two aliases of the SAME canonical city must remain unambiguous.
	sameCanonicalCatalog := normalize.NewCityCatalog(map[string]string{
		"charqueadas": "Charqueadas",
		"charquedas":  "Charqueadas",
	})
	for i := 0; i < 50; i++ {
		if got := sameCanonicalCatalog.FormatCity("charqueda"); got != "Charqueadas" {
			t.Fatalf("iteration %d: same-canonical tie FormatCity(charqueda) = %q, want Charqueadas", i, got)
		}
	}
}

func TestFormatCityMultiByteRuneLevenshtein(t *testing.T) {
	t.Parallel()

	// "münchen" in SearchText is "munchen" (ASCII), whereas non-decomposable letters
	// like "ø" or "ł" stay multi-byte UTF-8 runes in SearchText ("trømsø" -> "trømsø").
	// Verify that edit distance is measured in runes, not UTF-8 byte length.
	nordicCatalog := normalize.NewCityCatalog(map[string]string{
		"trømsø": "Trømsø",
	})
	// "trømso" differs from "trømsø" by 1 rune ('o' vs 'ø'), though 'ø' is 2 bytes in UTF-8.
	if got := nordicCatalog.FormatCity("trømso"); got != "Trømsø" {
		t.Fatalf("FormatCity(trømso) = %q, want Trømsø", got)
	}
}

func TestFormattersIdempotency(t *testing.T) {
	t.Parallel()

	inputs := []string{
		"",
		"nao",
		"Porto Alegre",
		"poa",
		"chaqueadas",
		"barato",
		"bella vista",
		"Centro",
		"bra",
		"Canada e Mexico",
		"Ana da Silva",
		"masculino",
		"casada",
		"o positivo",
		"90010-000",
		"rs",
	}

	formatters := []struct {
		name string
		fn   func(string) string
	}{
		{"DisplayText", normalize.DisplayText},
		{"SearchText", normalize.SearchText},
		{"Digits", normalize.Digits},
		{"Alphanumeric", normalize.Alphanumeric},
		{"FormatCity", normalize.FormatCity},
		{"FormatMunicipality", normalize.FormatMunicipality},
		{"FormatNeighborhood", normalize.FormatNeighborhood},
		{"FormatCountry", normalize.FormatCountry},
		{"FormatNationality", normalize.FormatNationality},
		{"FormatPersonName", normalize.FormatPersonName},
		{"FormatGender", normalize.FormatGender},
		{"FormatMaritalStatus", normalize.FormatMaritalStatus},
		{"FormatBloodType", normalize.FormatBloodType},
		{"FormatCEP", normalize.FormatCEP},
		{"FormatUF", normalize.FormatUF},
		{"FormatStreet", normalize.FormatStreet},
		{"FormatHouseNumber", normalize.FormatHouseNumber},
		{"FormatBlock", normalize.FormatBlock},
		{"FormatApartment", normalize.FormatApartment},
	}

	for _, f := range formatters {
		for _, in := range inputs {
			first := f.fn(in)
			second := f.fn(first)
			if second != first {
				t.Fatalf("%s not idempotent on %q: first=%q, second=%q", f.name, in, first, second)
			}
		}
	}
}
