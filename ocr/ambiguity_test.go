package ocr_test

import (
	"encoding/json"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/ocr"
)

func TestCandidateAmbiguityMustBeExplicit(t *testing.T) {
	t.Parallel()

	result := ocr.ExtractionResult{
		Mode: ocr.ModeDiscovery,
		Observations: []ocr.Observation{{
			ID: "obs-1", Evidence: []ocr.EvidenceRef{{SourceID: "image-1"}}, RawText: "42",
		}},
		Candidates: []ocr.FieldCandidate{
			{Path: "/number", State: ocr.ValuePresent, Value: json.RawMessage(`"42"`), ObservationIDs: []string{"obs-1"}, Basis: ocr.BasisObserved, Validation: ocr.ValidationNotValidated, Review: ocr.ReviewUnreviewed},
			{Path: "/number", State: ocr.ValuePresent, Value: json.RawMessage(`42`), ObservationIDs: []string{"obs-1"}, Basis: ocr.BasisInferred, Validation: ocr.ValidationNotValidated, Review: ocr.ReviewNeedsReview},
		},
		Validation: ocr.ValidationNotValidated,
		Review:     ocr.ReviewNeedsReview,
	}
	assertCode(t, ocr.ValidateExtractionResult(result), ocr.CodeInvalidCandidate)

	result.Candidates[0].Ambiguous = true
	result.Candidates[1].Ambiguous = true
	if err := ocr.ValidateExtractionResult(result); err != nil {
		t.Fatalf("explicit path ambiguity = %v", err)
	}
}

func TestCandidateReviewResolvesPathAmbiguity(t *testing.T) {
	t.Parallel()

	result := ocr.ExtractionResult{
		Mode: ocr.ModeDiscovery,
		Observations: []ocr.Observation{{
			ID: "obs-1", Evidence: []ocr.EvidenceRef{{SourceID: "image-1"}}, RawText: "42",
		}},
		Candidates: []ocr.FieldCandidate{
			{Path: "/number", State: ocr.ValuePresent, Value: json.RawMessage(`"42"`), ObservationIDs: []string{"obs-1"}, Basis: ocr.BasisObserved, Validation: ocr.ValidationNotValidated, Review: ocr.ReviewAccepted},
			{Path: "/number", State: ocr.ValuePresent, Value: json.RawMessage(`42`), ObservationIDs: []string{"obs-1"}, Basis: ocr.BasisInferred, Ambiguous: true, Validation: ocr.ValidationNotValidated, Review: ocr.ReviewUnreviewed},
		},
		Validation: ocr.ValidationNotValidated,
		Review:     ocr.ReviewNeedsReview,
	}
	assertCode(t, ocr.ValidateExtractionResult(result), ocr.CodeInvalidReview)

	result.Candidates[1].Review = ocr.ReviewRejected
	if err := ocr.ValidateExtractionResult(result); err != nil {
		t.Fatalf("resolved path ambiguity = %v", err)
	}
}

func TestCandidateInternalAlternativesRequireAmbiguity(t *testing.T) {
	t.Parallel()

	result := ocr.ExtractionResult{
		Mode: ocr.ModeDiscovery,
		Observations: []ocr.Observation{{
			ID: "obs-1", Evidence: []ocr.EvidenceRef{{SourceID: "image-1"}}, RawText: "12345",
		}},
		Candidates: []ocr.FieldCandidate{{
			Path: "/identifier", State: ocr.ValuePresent, Value: json.RawMessage(`"12345"`), ObservationIDs: []string{"obs-1"}, Basis: ocr.BasisObserved,
			JurisdictionCandidates: []string{"br", "us"}, Validation: ocr.ValidationNotValidated, Review: ocr.ReviewNeedsReview,
		}},
		Validation: ocr.ValidationNotValidated,
		Review:     ocr.ReviewNeedsReview,
	}
	assertCode(t, ocr.ValidateExtractionResult(result), ocr.CodeInvalidCandidate)

	result.Candidates[0].Ambiguous = true
	if err := ocr.ValidateExtractionResult(result); err != nil {
		t.Fatalf("explicit internal ambiguity = %v", err)
	}

	result.Candidates[0].Review = ocr.ReviewAccepted
	assertCode(t, ocr.ValidateExtractionResult(result), ocr.CodeInvalidReview)
}
