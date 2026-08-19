package ocr_test

import (
	"encoding/json"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/ocr"
)

func FuzzExtractionRequestJSON(f *testing.F) {
	f.Add([]byte(`{"mode":"discovery","sources":[{"id":"image-1","modality":"image"}]}`))
	f.Add([]byte(`{"mode":"schema_guided","sources":[{"id":"doc-1","modality":"document"}],"target_schema":{"type":"object","properties":{},"required":[],"additionalProperties":false}}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			t.Skip()
		}
		var request ocr.ExtractionRequest
		if json.Unmarshal(data, &request) != nil {
			return
		}
		if ocr.ValidateExtractionRequest(request) != nil {
			return
		}
		roundTrip, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("marshal valid request: %v", err)
		}
		var decoded ocr.ExtractionRequest
		if err := json.Unmarshal(roundTrip, &decoded); err != nil {
			t.Fatalf("unmarshal round trip: %v", err)
		}
		if err := ocr.ValidateExtractionRequest(decoded); err != nil {
			t.Fatalf("round-trip request invalid: %v", err)
		}
	})
}

func FuzzExtractionResultJSON(f *testing.F) {
	f.Add([]byte(`{"mode":"discovery","observations":[],"candidates":[],"validation":"not_validated","review":"unreviewed"}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			t.Skip()
		}
		var result ocr.ExtractionResult
		if json.Unmarshal(data, &result) != nil {
			return
		}
		if ocr.ValidateExtractionResult(result) != nil {
			return
		}
		roundTrip, err := json.Marshal(result)
		if err != nil {
			t.Fatalf("marshal valid result: %v", err)
		}
		var decoded ocr.ExtractionResult
		if err := json.Unmarshal(roundTrip, &decoded); err != nil {
			t.Fatalf("unmarshal round trip: %v", err)
		}
		if err := ocr.ValidateExtractionResult(decoded); err != nil {
			t.Fatalf("round-trip result invalid: %v", err)
		}
	})
}
