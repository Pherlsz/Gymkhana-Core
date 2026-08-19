package fingerprint

import (
	"errors"
	"fmt"
)

// ErrorCode identifies a stable fingerprint validation failure.
type ErrorCode string

const (
	CodeInvalidNamespace ErrorCode = "invalid_namespace"
	CodeInvalidLength    ErrorCode = "invalid_length"
	CodeInvalidFormat    ErrorCode = "invalid_format"
	CodeInvalidUTF8      ErrorCode = "invalid_utf8"
	CodeTooManyParts     ErrorCode = "too_many_parts"
)

// ValidationError reports a stable failure without embedding caller input.
type ValidationError struct {
	Code ErrorCode
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("fingerprint: %s", e.Code)
}

// IsCode reports whether err contains the requested stable fingerprint code.
func IsCode(err error, code ErrorCode) bool {
	var validationError *ValidationError
	return errors.As(err, &validationError) && validationError.Code == code
}

func validationError(code ErrorCode) error {
	return &ValidationError{Code: code}
}
