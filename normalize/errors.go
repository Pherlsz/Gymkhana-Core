package normalize

import (
	"errors"
	"fmt"
)

// ValueKind identifies the value being normalized.
type ValueKind string

const (
	KindCPF         ValueKind = "cpf"
	KindCNPJ        ValueKind = "cnpj"
	KindEmail       ValueKind = "email"
	KindBrazilPhone ValueKind = "brazil_phone"
	KindDocument    ValueKind = "document"
	KindPostalCode  ValueKind = "postal_code"
	KindPlate       ValueKind = "vehicle_plate"
)

// ErrorCode identifies a stable validation failure without exposing the input.
type ErrorCode string

const (
	CodeEmpty           ErrorCode = "empty"
	CodeInvalidFormat   ErrorCode = "invalid_format"
	CodeInvalidLength   ErrorCode = "invalid_length"
	CodeInvalidChecksum ErrorCode = "invalid_checksum"
	CodeInvalidAreaCode ErrorCode = "invalid_area_code"
	CodeInvalidNumber   ErrorCode = "invalid_number"
	CodeUnknownKind     ErrorCode = "unknown_kind"
	CodeMissingState    ErrorCode = "missing_state"
	CodeInvalidState    ErrorCode = "invalid_state"
	CodeAmbiguous       ErrorCode = "ambiguous"
)

// ValidationError is returned when a value cannot be canonicalized safely.
type ValidationError struct {
	Kind ValueKind
	Code ErrorCode
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Kind, e.Code)
}

// IsCode reports whether err contains the requested stable validation code.
func IsCode(err error, code ErrorCode) bool {
	var validationError *ValidationError
	return errors.As(err, &validationError) && validationError.Code == code
}

func validationError(kind ValueKind, code ErrorCode) error {
	return &ValidationError{Kind: kind, Code: code}
}
