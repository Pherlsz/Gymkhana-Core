package ocr

import (
	"errors"
	"fmt"
)

// ErrorCode is a stable, non-localized OCR/extraction validation code.
type ErrorCode string

const (
	CodeInvalidMode           ErrorCode = "invalid_mode"
	CodeInvalidSource         ErrorCode = "invalid_source"
	CodeDuplicateSource       ErrorCode = "duplicate_source"
	CodeInvalidSchema         ErrorCode = "invalid_schema"
	CodeInvalidLimit          ErrorCode = "invalid_limit"
	CodeInvalidEvidence       ErrorCode = "invalid_evidence"
	CodeUnknownSource         ErrorCode = "unknown_source"
	CodeInvalidObservation    ErrorCode = "invalid_observation"
	CodeDuplicateObservation  ErrorCode = "duplicate_observation"
	CodeInvalidCandidate      ErrorCode = "invalid_candidate"
	CodeUnknownObservation    ErrorCode = "unknown_observation"
	CodeInvalidConfidence     ErrorCode = "invalid_confidence"
	CodeInvalidReview         ErrorCode = "invalid_review"
	CodeInvalidValidation     ErrorCode = "invalid_validation"
	CodeInvalidWarning        ErrorCode = "invalid_warning"
	CodeInvalidResult         ErrorCode = "invalid_result"
	CodeInvalidStructuredData ErrorCode = "invalid_structured_data"
)

// ValidationError deliberately contains only a stable code and semantic field.
// Rejected source content, extracted text, values, URIs, and provider payloads
// are never embedded in the error string.
type ValidationError struct {
	Code  ErrorCode
	Field string
}

func (e *ValidationError) Error() string {
	if e == nil {
		return "ocr: validation error"
	}
	if e.Field == "" {
		return fmt.Sprintf("ocr: %s", e.Code)
	}
	return fmt.Sprintf("ocr: %s: %s", e.Field, e.Code)
}

func validationError(code ErrorCode, field string) error {
	return &ValidationError{Code: code, Field: field}
}

// IsCode reports whether err contains the requested stable validation code.
func IsCode(err error, code ErrorCode) bool {
	var validation *ValidationError
	return errors.As(err, &validation) && validation.Code == code
}
