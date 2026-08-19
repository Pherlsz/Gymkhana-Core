package conformance

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/ocr"
)

func TestOCRConformance(t *testing.T) { runSuite(t, "ocr.json", executeOCR) }

type extractionExchangeInput struct { Request ocr.ExtractionRequest `json:"request"`; Result ocr.ExtractionResult `json:"result"` }

func executeOCR(testCase vector) (any, string) {
	switch testCase.Operation {
	case "ocr.request.validate":
		var value ocr.ExtractionRequest; if err := json.Unmarshal(testCase.Input, &value); err != nil { return nil, "invalid_conformance_input" }; return ocrValidationResult(ocr.ValidateExtractionRequest(value))
	case "ocr.result.validate":
		var value ocr.ExtractionResult; if err := json.Unmarshal(testCase.Input, &value); err != nil { return nil, "invalid_conformance_input" }; return ocrValidationResult(ocr.ValidateExtractionResult(value))
	case "ocr.exchange.validate":
		var value extractionExchangeInput; if err := json.Unmarshal(testCase.Input, &value); err != nil { return nil, "invalid_conformance_input" }; return ocrValidationResult(ocr.ValidateExtractionExchange(value.Request, value.Result))
	default:
		return nil, "unsupported_operation"
	}
}

func ocrValidationResult(err error) (any, string) {
	if err == nil { return true, "" }
	var validation *ocr.ValidationError; if errors.As(err, &validation) { return nil, string(validation.Code) }; return nil, "unknown_error"
}
