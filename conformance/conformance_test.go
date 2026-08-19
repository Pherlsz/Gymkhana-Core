package conformance

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
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
	Name      string          `json:"name"`
	Operation string          `json:"operation"`
	Input     json.RawMessage `json:"input"`
	Output    json.RawMessage `json:"output,omitempty"`
	Error     *string         `json:"error,omitempty"`
}

type executor func(vector) (any, string)

func TestNormalizeConformance(t *testing.T) {
	runSuite(t, "normalize.json", executeNormalize)
}

func TestTemporalConformance(t *testing.T) {
	runSuite(t, "temporal.json", executeTemporal)
}

func runSuite(t *testing.T, filename string, execute executor) {
	t.Helper()

	expectedVersionData, err := os.ReadFile(filepath.Join("..", "spec", "VERSION"))
	if err != nil {
		t.Fatalf("read specification version: %v", err)
	}
	expectedVersion := strings.TrimSpace(string(expectedVersionData))
	path := filepath.Join("v"+expectedVersion, filename)

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
		if testCase.Name == "" || testCase.Operation == "" || len(testCase.Input) == 0 || (testCase.Output == nil) == (testCase.Error == nil) {
			t.Fatalf("conformance case in %s has an invalid shape: %#v", path, testCase)
		}

		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			output, errorCode := execute(testCase)
			if testCase.Error != nil {
				if errorCode != *testCase.Error {
					t.Fatalf("%s error = %q, want %q", testCase.Operation, errorCode, *testCase.Error)
				}
				if output != nil {
					t.Fatalf("%s returned output %#v with error %q", testCase.Operation, output, errorCode)
				}
				return
			}

			if errorCode != "" {
				t.Fatalf("%s unexpected error = %q", testCase.Operation, errorCode)
			}
			if equal, actualJSON, err := equivalentJSON(output, testCase.Output); err != nil {
				t.Fatalf("%s compare output: %v", testCase.Operation, err)
			} else if !equal {
				t.Fatalf("%s output = %s, want %s", testCase.Operation, actualJSON, testCase.Output)
			}
		})
	}
}

func equivalentJSON(actual any, expected json.RawMessage) (bool, []byte, error) {
	actualJSON, err := json.Marshal(actual)
	if err != nil {
		return false, nil, err
	}

	var actualValue any
	if err := json.Unmarshal(actualJSON, &actualValue); err != nil {
		return false, actualJSON, err
	}
	var expectedValue any
	if err := json.Unmarshal(expected, &expectedValue); err != nil {
		return false, actualJSON, err
	}
	return reflect.DeepEqual(actualValue, expectedValue), actualJSON, nil
}

func executeNormalize(testCase vector) (any, string) {
	switch testCase.Operation {
	case "text.display":
		input, ok := stringInput(testCase.Input)
		if !ok {
			return nil, "invalid_conformance_input"
		}
		return normalize.DisplayText(input), ""
	case "text.search":
		input, ok := stringInput(testCase.Input)
		if !ok {
			return nil, "invalid_conformance_input"
		}
		return normalize.SearchText(input), ""
	case "contact.email.canonicalize":
		return executeNormalizeString(testCase.Input, normalize.CanonicalEmail)
	case "contact.phone.br.canonicalize":
		return executeNormalizeString(testCase.Input, normalize.CanonicalBrazilPhone)
	case "identity.br.cpf.canonicalize":
		return executeNormalizeString(testCase.Input, normalize.CanonicalCPF)
	case "identity.br.cnpj.canonicalize":
		return executeNormalizeString(testCase.Input, normalize.CanonicalCNPJ)
	case "postal.br.cep.canonicalize":
		return executeNormalizeString(testCase.Input, normalize.CanonicalCEP)
	case "identity.br.document.canonicalize":
		input, ok := structuredDocumentInput(testCase.Input)
		if !ok {
			return nil, "invalid_conformance_input"
		}
		return normalizeResult(normalize.CanonicalDocument(normalize.DocumentKind(input.Kind), input.Value))
	case "identity.br.document.format":
		input, ok := structuredDocumentInput(testCase.Input)
		if !ok {
			return nil, "invalid_conformance_input"
		}
		return normalizeResult(normalize.FormatDocument(normalize.DocumentKind(input.Kind), input.Value))
	default:
		return nil, "unsupported_operation"
	}
}

func executeTemporal(testCase vector) (any, string) {
	input, ok := stringInput(testCase.Input)
	if !ok {
		return nil, "invalid_conformance_input"
	}

	switch testCase.Operation {
	case "temporal.civil_date.parse":
		value, err := civiltime.ParseCivilDate(input)
		if err != nil {
			return nil, civiltimeErrorCode(err)
		}
		return value.String(), ""
	case "temporal.year_month.parse":
		value, err := civiltime.ParseYearMonth(input)
		if err != nil {
			return nil, civiltimeErrorCode(err)
		}
		return value.String(), ""
	default:
		return nil, "unsupported_operation"
	}
}

func executeNormalizeString(input json.RawMessage, operation func(string) (string, error)) (any, string) {
	value, ok := stringInput(input)
	if !ok {
		return nil, "invalid_conformance_input"
	}
	return normalizeResult(operation(value))
}

func stringInput(raw json.RawMessage) (string, bool) {
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", false
	}
	return value, true
}

type documentInput struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

func structuredDocumentInput(raw json.RawMessage) (documentInput, bool) {
	var value documentInput
	if err := json.Unmarshal(raw, &value); err != nil || value.Kind == "" {
		return documentInput{}, false
	}
	return value, true
}

func normalizeResult(value string, err error) (any, string) {
	if err == nil {
		return value, ""
	}
	var validation *normalize.ValidationError
	if errors.As(err, &validation) {
		return nil, string(validation.Code)
	}
	return nil, "unknown_error"
}

func civiltimeErrorCode(err error) string {
	var validation *civiltime.ValidationError
	if errors.As(err, &validation) {
		return string(validation.Code)
	}
	return "unknown_error"
}
