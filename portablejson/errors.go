// Package portablejson provides language-neutral JSON and structured-schema
// validation primitives shared by Core domains.
package portablejson

import "errors"

// ErrorCode is a stable, non-localized portable JSON validation code.
type ErrorCode string

const (
	CodeInvalidJSON   ErrorCode = "invalid_json"
	CodeInvalidSchema ErrorCode = "invalid_schema"
)

// ValidationError contains only a stable code and semantic field. It never
// embeds the rejected JSON payload.
type ValidationError struct {
	Code  ErrorCode
	Field string
}

func (err *ValidationError) Error() string {
	if err == nil {
		return ""
	}
	if err.Field == "" {
		return string(err.Code)
	}
	return err.Field + ": " + string(err.Code)
}

// IsCode reports whether err contains a portablejson ValidationError with code.
func IsCode(err error, code ErrorCode) bool {
	var validation *ValidationError
	return errors.As(err, &validation) && validation.Code == code
}

func validationError(code ErrorCode, field string) error {
	return &ValidationError{Code: code, Field: field}
}
