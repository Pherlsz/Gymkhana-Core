package normalize

import "testing"

func TestValidBrazilAreaCode(t *testing.T) {
	t.Parallel()

	for _, areaCode := range []string{"11", "24", "38", "51", "69", "79", "89", "99"} {
		if !validBrazilAreaCode(areaCode) {
			t.Fatalf("expected valid area code %s", areaCode)
		}
	}

	for _, areaCode := range []string{"00", "10", "20", "23", "36", "52", "70", "72", "76", "78", "80", "90"} {
		if validBrazilAreaCode(areaCode) {
			t.Fatalf("expected invalid area code %s", areaCode)
		}
	}
}
