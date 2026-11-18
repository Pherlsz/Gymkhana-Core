package ocr_test

import (
	"encoding/json"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/ocr"
)

func BenchmarkValidateExtractionExchangeSchemaGuided(b *testing.B) {
	request := validRequest()
	result := validResult()

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if err := ocr.ValidateExtractionExchange(request, result); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateExtractionResultDiscoveryAmbiguous(b *testing.B) {
	result := ocr.ExtractionResult{
		Mode: ocr.ModeDiscovery,
		Observations: []ocr.Observation{{
			ID:       "obs-1",
			Evidence: []ocr.EvidenceRef{{SourceID: "image-1", Region: &ocr.NormalizedRect{X: 100, Y: 200, Width: 500, Height: 300}}},
			RawText:  "42",
		}},
		Candidates: []ocr.FieldCandidate{
			{
				Path:           "/invoice/items/0/amount",
				State:          ocr.ValuePresent,
				Value:          json.RawMessage(`"42"`),
				ObservationIDs: []string{"obs-1"},
				Basis:          ocr.BasisObserved,
				Ambiguous:      true,
				Validation:     ocr.ValidationNotValidated,
				Review:         ocr.ReviewUnreviewed,
			},
			{
				Path:           "/invoice/items/0/amount",
				State:          ocr.ValuePresent,
				Value:          json.RawMessage(`42`),
				ObservationIDs: []string{"obs-1"},
				Basis:          ocr.BasisInferred,
				Ambiguous:      true,
				Validation:     ocr.ValidationNotValidated,
				Review:         ocr.ReviewNeedsReview,
			},
		},
		Validation: ocr.ValidationNotValidated,
		Review:     ocr.ReviewNeedsReview,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if err := ocr.ValidateExtractionResult(result); err != nil {
			b.Fatal(err)
		}
	}
}
