package conformance_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/civiltime"
	"github.com/Pherlsz/Gymkhana-Core/normalize"
)

type suite struct {
	SpecVersion string   `json:"spec_version"`
	Suite       string   `json:"suite"`
	Cases       []vector `json:"cases"`
}

type vector struct {
	Name      string  `json:"name"`
	Operation string  `json:"operation"`
	Input     string  `json:"input"`
	Output    *string `json:"output,omitempty"`
	Error     *string `json:"error,omitempty"`
}

func TestNormalizeConformance(t *testing.T) {
	runSuite(t, filepath.Join("v0.1", "normalize.json"), executeNormalize)
}

func TestTemporalConformance(t *testing.T) {
	runSuite(t, filepath.Join("v0.1", "temporal.json"), executeTemporal)
}

func runSuite(t *testing.T, path string, execute func(vector) (string, string)) {
	t.Helper()

	expectedVersionData, err := os.ReadFile(filepath.Join("..", "spec", "VERSION"))
	if err != nil {
		t.Fatalf("read specification version: %v", err)
	}
	expectedVersion := strings.TrimSpace(string(expectedVersionData))

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read conformance suite %s: %v", path, err)
	}

	var vectors suite
	if err := json.Unmarshal(data, &vectors); err != nil {
		t.Fatalf("decode conformance suite %s: %v", path, err)
	}
	if vectors.SpecVersion != expectedVersion {
		t.Fatalf("conformance suite %s targets spec %q, current spec is %q", path, vectors.SpecVersion, expectedVersion)
	}
	if vectors.Suite == "" || len(vectors.Cases) == 0 {
		t.Fatalf("conformance suite %s is incomplete", path)
	}

	for _, testCase := range vectors.Cases {
		if testCase.Name == "" || testCase.Operation == "" || (testCase.Output == nil) == (testCase.Error == nil) {
			t.Fatalf("conformance case in %s has an invalid result shape: %#v", path, testCase)
		}

		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			output, errorCode := execute(testCase)
			if testCase.Error != nil {
				if errorCode != *testCase.Error {
					t.Fatalf("%s error = %q, want %q", testCase.Operation, errorCode, *testCase.Error)
				}
				if output != "" {
					t.Fatalf("%s returned output %q with error %q", testCase.Operation, output, errorCode)
				}
				return
			}

			if errorCode != "" {
				t.Fatalf("%s unexpected error = %q", testCase.Operation, errorCode)
			}
			if output != *testCase.Output {
				t.Fatalf("%s output = %q, want %q", testCase.Operation, output, *testCase.Output)
			}
		})
	}
}

func executeNormalize(testCase vector) (string, string) {
	switch testCase.Operation {
	case "text.display":
		return normalize.DisplayText(testCase.Input), ""
	case "text.search":
		return normalize.SearchText(testCase.Input), ""
	case "contact.email.canonicalize":
		return normalizeResult(normalize.CanonicalEmail(testCase.Input))
	case "contact.phone.br.canonicalize":
		return normalizeResult(normalize.CanonicalBrazilPhone(testCase.Input))
	case "identity.br.cpf.canonicalize":
		return normalizeResult(normalize.CanonicalCPF(testCase.Input))
	case "identity.br.cnpj.canonicalize":
		return normalizeResult(normalize.CanonicalCNPJ(testCase.Input))
	case "postal.br.cep.canonicalize":
		return normalizeResult(normalize.CanonicalCEP(testCase.Input))
	default:
		return "", "unsupported_operation"
	}
}

func executeTemporal(testCase vector) (string, string) {
	switch testCase.Operation {
	case "temporal.civil_date.parse":
		value, err := civiltime.ParseCivilDate(testCase.Input)
		if err != nil {
			return "", civiltimeErrorCode(err)
		}
		return value.String(), ""
	case "temporal.year_month.parse":
		value, err := civiltime.ParseYearMonth(testCase.Input)
		if err != nil {
			return "", civiltimeErrorCode(err)
		}
		return value.String(), ""
	default:
		return "", "unsupported_operation"
	}
}

func normalizeResult(value string, err error) (string, string) {
	if err == nil {
		return value, ""
	}
	var validation *normalize.ValidationError
	if errors.As(err, &validation) {
		return "", string(validation.Code)
	}
	return "", "unknown_error"
}

func civiltimeErrorCode(err error) string {
	var validation *civiltime.ValidationError
	if errors.As(err, &validation) {
		return string(validation.Code)
	}
	return "unknown_error"
}
