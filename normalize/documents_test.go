package normalize_test

import (
	"strings"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/normalize"
)

func TestCanonicalDocumentCPFMatchesCanonicalCPF(t *testing.T) {
	t.Parallel()

	got, err := normalize.CanonicalDocument(normalize.DocumentCPF, "529.982.247-25")
	if err != nil {
		t.Fatalf("CanonicalDocument() error = %v", err)
	}
	direct, err := normalize.CanonicalCPF("529.982.247-25")
	if err != nil {
		t.Fatalf("CanonicalCPF() error = %v", err)
	}
	if got != direct || got != "52998224725" {
		t.Fatalf("CanonicalDocument() = %q, CanonicalCPF() = %q", got, direct)
	}
}

func TestFormatDocumentNationalNumbers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		kind      normalize.DocumentKind
		input     string
		canonical string
		display   string
	}{
		{normalize.DocumentCPF, "529.982.247-25", "52998224725", "529.982.247-25"},
		{normalize.DocumentCNPJ, "11.222.333/0001-81", "11222333000181", "11.222.333/0001-81"},
		{normalize.DocumentCNPJ, "12.abc.345/01de-35", "12ABC34501DE35", "12.ABC.345/01DE-35"},
		{normalize.DocumentCNH, "02172931125", "02172931125", "021 729 311 25"},
		{normalize.DocumentCNH, "12345678900", "12345678900", "123 456 789 00"},
		{normalize.DocumentCNH, "62472927637", "62472927637", "624 729 276 37"},
		{normalize.DocumentCNH, "69044271146", "69044271146", "690 442 711 46"},
		{normalize.DocumentCNH, "00000001801", "00000001801", "000 000 018 01"},
		{normalize.DocumentPIS, "120.5824.883-1", "12058248831", "120.58248.83-1"},
		{normalize.DocumentPIS, "12345678900", "12345678900", "123.45678.90-0"},
		{normalize.DocumentVoterID, "1234 5678 0493", "123456780493", "1234 5678 0493"},
		{normalize.DocumentVoterID, "123456780191", "123456780191", "1234 5678 0191"},
		{normalize.DocumentVoterID, "000000010299", "000000010299", "0000 0001 0299"},
		{normalize.DocumentSUSCard, "123 4567 8901 0000", "123456789010000", "123 4567 8901 0000"},
		{normalize.DocumentSUSCard, "200000000000003", "200000000000003", "200 0000 0000 0003"},
		{normalize.DocumentSUSCard, "700000000000005", "700000000000005", "700 0000 0000 0005"},
		{normalize.DocumentPassport, "ab 123456", "AB123456", "AB123456"},
	}

	for _, test := range tests {
		test := test
		t.Run(string(test.kind)+"/"+test.input, func(t *testing.T) {
			t.Parallel()
			got, err := normalize.CanonicalDocument(test.kind, test.input)
			if err != nil {
				t.Fatalf("CanonicalDocument() error = %v", err)
			}
			if got != test.canonical {
				t.Fatalf("CanonicalDocument() = %q, want %q", got, test.canonical)
			}
			display, err := normalize.FormatDocument(test.kind, test.input)
			if err != nil {
				t.Fatalf("FormatDocument() error = %v", err)
			}
			if display != test.display {
				t.Fatalf("FormatDocument() = %q, want %q", display, test.display)
			}
			again, err := normalize.CanonicalDocument(test.kind, display)
			if err != nil || again != test.canonical {
				t.Fatalf("round-trip CanonicalDocument() = %q, err = %v", again, err)
			}
		})
	}
}

func TestCanonicalDocumentStateIdentifiersIncludeUF(t *testing.T) {
	t.Parallel()

	tests := []struct {
		kind      normalize.DocumentKind
		input     string
		canonical string
		display   string
	}{
		{normalize.DocumentOAB, "OAB/RS 114.458", "114458/RS", "114.458/RS"},
		{normalize.DocumentOAB, "114458-rs", "114458/RS", "114.458/RS"},
		{normalize.DocumentOAB, "RS114458", "114458/RS", "114.458/RS"},
		{normalize.DocumentCREA, "CREA 12345/SP", "12345/SP", "12.345/SP"},
		{normalize.DocumentCRM, "crm-mg 67890", "67890/MG", "67.890/MG"},
		{normalize.DocumentCRO, "1234 / PR", "1234/PR", "1.234/PR"},
		{normalize.DocumentCOREN, "COREN/SC 998877", "998877/SC", "998.877/SC"},
		{normalize.DocumentCREAOAB, "114458/RS", "114458/RS", "114.458/RS"},
		{normalize.DocumentRG, "12.345.678-9/SP", "123456789/SP", "12.345.678-9/SP"},
		{normalize.DocumentRG, "12345678-X / RS", "12345678X/RS", "12.345.678-X/RS"},
		{normalize.DocumentRG, "12345678", "12345678", "1.234.567-8"},
		{normalize.DocumentIdentidade, "4.567.890-1", "45678901", "4.567.890-1"},
		{normalize.DocumentCTPS, "1234567-0012/RS", "1234567-0012/RS", "1.234.567-0012/RS"},
		{normalize.DocumentCTPS, "0012345", "0012345", "0.012.345"},
	}

	for _, test := range tests {
		test := test
		t.Run(string(test.kind)+"/"+test.input, func(t *testing.T) {
			t.Parallel()
			got, err := normalize.CanonicalDocument(test.kind, test.input)
			if err != nil {
				t.Fatalf("CanonicalDocument() error = %v", err)
			}
			if got != test.canonical {
				t.Fatalf("CanonicalDocument() = %q, want %q", got, test.canonical)
			}
			display, err := normalize.FormatDocument(test.kind, got)
			if err != nil {
				t.Fatalf("FormatDocument() error = %v", err)
			}
			if display != test.display {
				t.Fatalf("FormatDocument() = %q, want %q", display, test.display)
			}
		})
	}
}

func TestCanonicalDocumentLooseIdentifiers(t *testing.T) {
	t.Parallel()

	got, err := normalize.CanonicalDocument(normalize.DocumentStudentID, "carteirinha 182265")
	if err != nil {
		t.Fatalf("student_id error = %v", err)
	}
	if got != "182265" {
		t.Fatalf("student_id = %q", got)
	}

	got, err = normalize.CanonicalDocument(normalize.DocumentCitizenCard, "123456789012")
	if err != nil {
		t.Fatalf("citizen_card error = %v", err)
	}
	if got != "123456789012" {
		t.Fatalf("citizen_card = %q", got)
	}

	got, err = normalize.CanonicalDocument(normalize.DocumentBirthCertificate, "12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("birth_certificate error = %v", err)
	}
	if got != "12345678901234567890123456789012" {
		t.Fatalf("birth_certificate = %q", got)
	}
}

func TestCanonicalDocumentRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		kind normalize.DocumentKind
		in   string
		code normalize.ErrorCode
	}{
		{normalize.DocumentKind("unknown"), "123", normalize.CodeUnknownKind},
		{normalize.DocumentCPF, "111.111.111-11", normalize.CodeInvalidChecksum},
		{normalize.DocumentCNPJ, "00.000.000/0000-00", normalize.CodeInvalidChecksum},
		{normalize.DocumentCNPJ, "12.ABC.345/01DE-00", normalize.CodeInvalidChecksum},
		{normalize.DocumentCNPJ, "123", normalize.CodeInvalidLength},
		{normalize.DocumentCNH, "0217293112", normalize.CodeInvalidLength},
		{normalize.DocumentCNH, "11111111111", normalize.CodeInvalidChecksum},
		{normalize.DocumentCNH, "46190476839", normalize.CodeInvalidChecksum},
		{normalize.DocumentPIS, "12058248838", normalize.CodeInvalidChecksum},
		{normalize.DocumentPIS, "11111111111", normalize.CodeInvalidChecksum},
		{normalize.DocumentPIS, "1205824883", normalize.CodeInvalidLength},
		{normalize.DocumentVoterID, "123456780000", normalize.CodeInvalidChecksum},
		{normalize.DocumentSUSCard, "000000000000000", normalize.CodeInvalidChecksum},
		{normalize.DocumentSUSCard, "123456", normalize.CodeInvalidLength},
		{normalize.DocumentOAB, "114458", normalize.CodeMissingState},
		{normalize.DocumentOAB, "114458/XX", normalize.CodeInvalidState},
		{normalize.DocumentOAB, "nao", normalize.CodeInvalidFormat},
		{normalize.DocumentRG, "sim", normalize.CodeInvalidFormat},
		{normalize.DocumentPassport, "nao", normalize.CodeInvalidLength},
		{normalize.DocumentPassport, "sim 15 05 2017", normalize.CodeInvalidLength},
		{normalize.DocumentStudentID, "nao tenho", normalize.CodeInvalidFormat},
		{normalize.DocumentCPF, "", normalize.CodeEmpty},
	}

	for _, test := range tests {
		test := test
		t.Run(string(test.kind)+"/"+test.in+"/"+string(test.code), func(t *testing.T) {
			t.Parallel()
			_, err := normalize.CanonicalDocument(test.kind, test.in)
			if !normalize.IsCode(err, test.code) {
				t.Fatalf("error = %v, want %s", err, test.code)
			}
		})
	}
}

func TestDocumentKindsCoverCatalog(t *testing.T) {
	t.Parallel()

	kinds := normalize.DocumentKinds()
	if len(kinds) == 0 {
		t.Fatal("DocumentKinds() is empty")
	}
	seen := map[normalize.DocumentKind]bool{}
	for _, kind := range kinds {
		if seen[kind] {
			t.Fatalf("duplicate kind %s", kind)
		}
		seen[kind] = true
		if !normalize.KnownDocumentKind(kind) {
			t.Fatalf("KnownDocumentKind(%s) = false", kind)
		}
	}
	if !normalize.KnownDocumentKind(normalize.DocumentCPF) || normalize.KnownDocumentKind("other") {
		t.Fatal("KnownDocumentKind mismatch")
	}
}

func TestCanonicalDocumentPreservesLeadingZeros(t *testing.T) {
	t.Parallel()

	got, err := normalize.CanonicalDocument(normalize.DocumentCNH, "02172931125")
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if !strings.HasPrefix(got, "0") {
		t.Fatalf("lost leading zero: %q", got)
	}
}

func TestFormatDocumentIsIdempotentForCanonicalInput(t *testing.T) {
	t.Parallel()

	display, err := normalize.FormatDocument(normalize.DocumentOAB, "114458/RS")
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	again, err := normalize.FormatDocument(normalize.DocumentOAB, display)
	if err != nil {
		t.Fatalf("second FormatDocument() error = %v", err)
	}
	if display != again {
		t.Fatalf("FormatDocument not idempotent: %q vs %q", display, again)
	}
}

func TestNationalChecksumsMatchPublishedAlgorithms(t *testing.T) {
	t.Parallel()

	if got := receitaCPFCheck("529982247"); got != "25" {
		t.Fatalf("Receita CPF DV for 529982247 = %q, want 25", got)
	}
	if _, err := normalize.CanonicalDocument(normalize.DocumentCPF, "52998224725"); err != nil {
		t.Fatalf("CanonicalDocument(cpf) rejected the Receita example: %v", err)
	}

	contran := []string{"62472927637", "69044271146", "00000001801"}
	for _, number := range contran {
		if _, err := normalize.CanonicalDocument(normalize.DocumentCNH, number); err != nil {
			t.Fatalf("CanonicalDocument(cnh, %s) error = %v", number, err)
		}
	}
	if _, err := normalize.CanonicalDocument(normalize.DocumentCNH, "46190476839"); !normalize.IsCode(err, normalize.CodeInvalidChecksum) {
		t.Fatalf("invalid CONTRAN CNH was accepted: %v", err)
	}

	if got := caixaPISCheck("1205824883"); got != 1 {
		t.Fatalf("Caixa PIS DV for 1205824883 = %d, want 1", got)
	}
	if _, err := normalize.CanonicalDocument(normalize.DocumentPIS, "12058248831"); err != nil {
		t.Fatalf("CanonicalDocument(pis) rejected the Caixa-weight example: %v", err)
	}
	if _, err := normalize.CanonicalDocument(normalize.DocumentPIS, "52998224725"); !normalize.IsCode(err, normalize.CodeInvalidChecksum) {
		t.Fatalf("valid CPF was accepted as PIS: %v", err)
	}
	if _, err := normalize.CanonicalDocument(normalize.DocumentCPF, "12058248831"); !normalize.IsCode(err, normalize.CodeInvalidChecksum) {
		t.Fatalf("valid PIS was accepted as CPF: %v", err)
	}

	if _, err := normalize.CanonicalDocument(normalize.DocumentSUSCard, "700000000000005"); err != nil {
		t.Fatalf("provisional CNS rejected: %v", err)
	}
	if _, err := normalize.CanonicalDocument(normalize.DocumentSUSCard, "200000000000003"); err != nil {
		t.Fatalf("definitive CNS rejected: %v", err)
	}
	if _, err := normalize.CanonicalDocument(normalize.DocumentSUSCard, "30754.0"); !normalize.IsCode(err, normalize.CodeInvalidLength) {
		t.Fatalf("six-digit Excel SUS artifact should fail length, got %v", err)
	}
}

func TestTenDigitRGIsNotIdentifiedAsCPF(t *testing.T) {
	t.Parallel()

	got, err := normalize.CanonicalDocument(normalize.DocumentRG, "1122334455")
	if err != nil {
		t.Fatalf("CanonicalDocument(rg) error = %v", err)
	}
	if got != "1122334455" {
		t.Fatalf("CanonicalDocument(rg) = %q", got)
	}
	if _, err := normalize.IdentifyDocument("1122334455"); !normalize.IsCode(err, normalize.CodeUnknownKind) {
		t.Fatalf("IdentifyDocument(10-digit RG) = %v, want unknown_kind", err)
	}
	if _, err := normalize.CanonicalDocument(normalize.DocumentCPF, "1122334455"); !normalize.IsCode(err, normalize.CodeInvalidLength) {
		t.Fatalf("10-digit value should not pass CPF length, got %v", err)
	}
}

func receitaCPFCheck(base9 string) string {
	digit := func(prefix string, startWeight int) byte {
		sum := 0
		for i := 0; i < len(prefix); i++ {
			sum += int(prefix[i]-'0') * (startWeight - i)
		}
		remainder := sum % 11
		if remainder < 2 {
			return '0'
		}
		return byte('0' + (11 - remainder))
	}
	first := digit(base9, 10)
	second := digit(base9+string(first), 11)
	return string([]byte{first, second})
}

func caixaPISCheck(base10 string) int {
	weights := []int{3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i, weight := range weights {
		sum += int(base10[i]-'0') * weight
	}
	remainder := sum % 11
	if remainder < 2 {
		return 0
	}
	return 11 - remainder
}
