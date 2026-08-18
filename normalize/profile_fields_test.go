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
}

func TestFormatCatalogs(t *testing.T) {
	t.Parallel()

	if got := normalize.FormatTeam("Tamo nessa por cerveja"); got != "TNC" {
		t.Fatalf("FormatTeam = %q", got)
	}
	if got := normalize.FormatSector("maquiagem"); got != "Artística" {
		t.Fatalf("FormatSector = %q", got)
	}
	club := normalize.FormatClub("internacioal")
	if club.Club != "Internacional" {
		t.Fatalf("FormatClub = %#v", club)
	}
	if got := normalize.FormatClub("tiradentes"); got != (normalize.Club{}) {
		t.Fatalf("tiradentes must be empty, got %#v", got)
	}
	if got := normalize.FormatHealthPlan("Sim Unimed 12345"); got != "Unimed" {
		t.Fatalf("FormatHealthPlan = %q", got)
	}
	if got := normalize.FormatCollection("nenhuma"); got != "" {
		t.Fatalf("FormatCollection nenhum = %q", got)
	}
	if got := normalize.FormatAnimal("sim 1 poodle"); got != "Cachorro" {
		t.Fatalf("FormatAnimal = %q", got)
	}
}

func TestFormatVehicle(t *testing.T) {
	t.Parallel()

	if got := normalize.FormatVehiclePlate("UNO2010"); got != "" {
		t.Fatalf("UNO2010 must not be a plate, got %q", got)
	}
	if got := normalize.FormatVehicleModel("UNO2010"); got != "Uno" {
		t.Fatalf("FormatVehicleModel = %q", got)
	}
	if got := normalize.FormatVehicleYear("ANO2002"); got != "2002" {
		t.Fatalf("FormatVehicleYear = %q", got)
	}
	if got := normalize.FormatVehiclePlate("HYP1983"); got != "HYP1983" {
		t.Fatalf("real plate = %q", got)
	}
}
