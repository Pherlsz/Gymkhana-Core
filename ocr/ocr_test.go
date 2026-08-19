package ocr_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/ocr"
)

func schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"}},"required":["name"],"additionalProperties":false}`)
}

func confidence(value ocr.Confidence) *ocr.Confidence { return &value }

func validRequest() ocr.ExtractionRequest {
	return ocr.ExtractionRequest{
		Mode:          ocr.ModeSchemaGuided,
		Sources:       []ocr.SourceRef{{ID: "document-1", Modality: ocr.SourceDocument, MediaType: "application/pdf"}},
		TargetSchema:  schema(),
		MaxCandidates: 16,
	}
}

func validResult() ocr.ExtractionResult {
	return ocr.ExtractionResult{
		Mode: ocr.ModeSchemaGuided,
		Observations: []ocr.Observation{{
			ID: "obs-1",
			Evidence: []ocr.EvidenceRef{{
				SourceID: "document-1",
				Page:     1,
				Region:   &ocr.NormalizedRect{X: 100000, Y: 100000, Width: 400000, Height: 100000},
			}},
			RawText:    "Alice",
			Confidence: confidence(9800),
		}},
		Candidates: []ocr.FieldCandidate{{
			Path:           "/name",
			State:          ocr.ValuePresent,
			Value:          json.RawMessage(`"Alice"`),
			ObservationIDs: []string{"obs-1"},
			Basis:          ocr.BasisObserved,
			Confidence:     confidence(9800),
			Validation:     ocr.ValidationValid,
			Review:         ocr.ReviewUnreviewed,
		}},
		StructuredData: json.RawMessage(`{"name":"Alice"}`),
		Validation:     ocr.ValidationValid,
		Review:         ocr.ReviewUnreviewed,
	}
}

func TestValidateExtractionExchange(t *testing.T) {
	t.Parallel()
	if err := ocr.ValidateExtractionExchange(validRequest(), validResult()); err != nil {
		t.Fatalf("ValidateExtractionExchange(valid) = %v", err)
	}
}

func TestDiscoveryRejectsSchemaAndStructuredData(t *testing.T) {
	t.Parallel()

	request := ocr.ExtractionRequest{
		Mode:         ocr.ModeDiscovery,
		Sources:      []ocr.SourceRef{{ID: "image-1", Modality: ocr.SourceImage}},
		TargetSchema: schema(),
	}
	assertCode(t, ocr.ValidateExtractionRequest(request), ocr.CodeInvalidSchema)

	result := ocr.ExtractionResult{
		Mode:           ocr.ModeDiscovery,
		StructuredData: json.RawMessage(`{"name":"Alice"}`),
		Validation:     ocr.ValidationNotValidated,
		Review:         ocr.ReviewUnreviewed,
	}
	assertCode(t, ocr.ValidateExtractionResult(result), ocr.CodeInvalidResult)
}

func TestMissingCandidateDoesNotCollapseIntoNull(t *testing.T) {
	t.Parallel()

	result := ocr.ExtractionResult{
		Mode: ocr.ModeDiscovery,
		Observations: []ocr.Observation{{
			ID:       "obs-field-label",
			Evidence: []ocr.EvidenceRef{{SourceID: "image-1"}},
			RawText:  "Passport number",
		}},
		Candidates: []ocr.FieldCandidate{{
			Path:                   "/passport_number",
			State:                  ocr.ValueMissing,
			ObservationIDs:         []string{"obs-field-label"},
			Basis:                  ocr.BasisObserved,
			SemanticTypeCandidates: []string{"identity.passport"},
			Validation:             ocr.ValidationNotValidated,
			Review:                 ocr.ReviewNeedsReview,
		}},
		Validation: ocr.ValidationNotValidated,
		Review:     ocr.ReviewNeedsReview,
	}
	if err := ocr.ValidateExtractionResult(result); err != nil {
		t.Fatalf("missing candidate = %v", err)
	}

	result.Candidates[0].Value = json.RawMessage("null")
	assertCode(t, ocr.ValidateExtractionResult(result), ocr.CodeInvalidCandidate)
}

func TestNormalizationRequiresExplicitCanonicalizer(t *testing.T) {
	t.Parallel()

	result := validResult()
	result.Candidates[0].NormalizedValue = json.RawMessage(`"alice"`)
	assertCode(t, ocr.ValidateExtractionResult(result), ocr.CodeInvalidCandidate)

	result.Candidates[0].Normalizer = "text.display/v1"
	if err := ocr.ValidateExtractionResult(result); err != nil {
		t.Fatalf("normalized candidate = %v", err)
	}
}

func TestEvidenceBoundsAndModality(t *testing.T) {
	t.Parallel()

	result := validResult()
	result.Observations[0].Evidence[0].Region = &ocr.NormalizedRect{X: 900000, Y: 0, Width: 200000, Height: 1000}
	assertCode(t, ocr.ValidateExtractionResult(result), ocr.CodeInvalidEvidence)

	result = validResult()
	request := validRequest()
	request.Sources[0].Modality = ocr.SourceImage
	assertCode(t, ocr.ValidateExtractionExchange(request, result), ocr.CodeInvalidEvidence)
}

func TestSchemaValidationStatusMustMatchData(t *testing.T) {
	t.Parallel()

	result := validResult()
	result.StructuredData = json.RawMessage(`{}`)
	result.Candidates[0].State = ocr.ValueMissing
	result.Candidates[0].Value = nil
	result.Validation = ocr.ValidationValid
	assertCode(t, ocr.ValidateExtractionExchange(validRequest(), result), ocr.CodeInvalidValidation)

	result.Validation = ocr.ValidationInvalid
	if err := ocr.ValidateExtractionExchange(validRequest(), result); err != nil {
		t.Fatalf("explicit invalid extraction = %v", err)
	}
}

func TestSchemaGuidedCandidateMustMatchStructuredData(t *testing.T) {
	t.Parallel()

	result := validResult()
	result.Candidates[0].Value = json.RawMessage(`"Bob"`)
	assertCode(t, ocr.ValidateExtractionExchange(validRequest(), result), ocr.CodeInvalidCandidate)

	result = validResult()
	result.Candidates[0].Path = "/missing"
	assertCode(t, ocr.ValidateExtractionExchange(validRequest(), result), ocr.CodeInvalidCandidate)
}

func TestSchemaGuidedCandidateUsesMathematicalNumberEquality(t *testing.T) {
	t.Parallel()

	request := validRequest()
	request.TargetSchema = json.RawMessage(`{"type":"object","properties":{"score":{"type":"number"}},"required":["score"],"additionalProperties":false}`)
	result := validResult()
	result.Candidates[0].Path = "/score"
	result.Candidates[0].Value = json.RawMessage(`1.0`)
	result.StructuredData = json.RawMessage(`{"score":1}`)
	if err := ocr.ValidateExtractionExchange(request, result); err != nil {
		t.Fatalf("equivalent portable numbers = %v", err)
	}
}

func TestSchemaGuidedCandidateRejectsUnboundedNumber(t *testing.T) {
	t.Parallel()

	request := validRequest()
	request.TargetSchema = json.RawMessage(`{"type":"object","properties":{"score":{"type":"number"}},"required":["score"],"additionalProperties":false}`)
	result := validResult()
	result.Candidates[0].Path = "/score"
	result.Candidates[0].Value = json.RawMessage(`1e1000000`)
	result.StructuredData = json.RawMessage(`{"score":1}`)
	assertCode(t, ocr.ValidateExtractionExchange(request, result), ocr.CodeInvalidCandidate)
}

func TestResultAggregateDataIsBounded(t *testing.T) {
	t.Parallel()

	value := json.RawMessage(`"` + strings.Repeat("x", 9000) + `"`)
	result := ocr.ExtractionResult{
		Mode: ocr.ModeDiscovery,
		Observations: []ocr.Observation{{
			ID: "obs-1", Evidence: []ocr.EvidenceRef{{SourceID: "image-1"}}, RawText: "field",
		}},
		Validation: ocr.ValidationNotValidated,
		Review:     ocr.ReviewUnreviewed,
	}
	for i := 0; i < ocr.MaxCandidates; i++ {
		result.Candidates = append(result.Candidates, ocr.FieldCandidate{
			Path: "/field", State: ocr.ValuePresent, Value: value,
			ObservationIDs: []string{"obs-1"}, Basis: ocr.BasisObserved,
			Validation: ocr.ValidationNotValidated, Review: ocr.ReviewUnreviewed,
		})
	}
	assertCode(t, ocr.ValidateExtractionResult(result), ocr.CodeInvalidLimit)
}

func TestUnknownSourceAndObservation(t *testing.T) {
	t.Parallel()

	result := validResult()
	result.Observations[0].Evidence[0].SourceID = "different-source"
	assertCode(t, ocr.ValidateExtractionExchange(validRequest(), result), ocr.CodeUnknownSource)

	result = validResult()
	result.Candidates[0].ObservationIDs = []string{"missing-observation"}
	assertCode(t, ocr.ValidateExtractionResult(result), ocr.CodeUnknownObservation)
}

func TestValidationErrorsDoNotEchoSensitiveInput(t *testing.T) {
	t.Parallel()

	secret := "CPF 123.456.789-09"
	result := validResult()
	result.Observations[0].RawText = strings.Repeat("x", ocr.MaxObservationTextBytes+1) + secret
	err := ocr.ValidateExtractionResult(result)
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("error leaked source content: %v", err)
	}
}

func assertCode(t *testing.T, err error, code ocr.ErrorCode) {
	t.Helper()
	if !ocr.IsCode(err, code) {
		t.Fatalf("error = %v, want code %s", err, code)
	}
}
