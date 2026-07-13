package civiltime

import (
	"errors"
	"fmt"
)

// ValueKind identifies the civil value being parsed or constructed.
type ValueKind string

const (
	KindCivilDate ValueKind = "civil_date"
	KindYearMonth ValueKind = "year_month"
)

// ErrorCode identifies a stable validation or arithmetic failure.
type ErrorCode string

const (
	CodeEmpty         ErrorCode = "empty"
	CodeInvalidFormat ErrorCode = "invalid_format"
	CodeInvalidYear   ErrorCode = "invalid_year"
	CodeInvalidMonth  ErrorCode = "invalid_month"
	CodeInvalidDay    ErrorCode = "invalid_day"
	CodeZeroValue     ErrorCode = "zero_value"
	CodeOutOfRange    ErrorCode = "out_of_range"
)

// ValidationError contains a stable error kind and code without localized text.
type ValidationError struct {
	Kind ValueKind
	Code ErrorCode
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Kind, e.Code)
}

// IsCode reports whether err contains the requested stable code.
func IsCode(err error, code ErrorCode) bool {
	var validationError *ValidationError
	return errors.As(err, &validationError) && validationError.Code == code
}

func validationError(kind ValueKind, code ErrorCode) error {
	return &ValidationError{Kind: kind, Code: code}
}
