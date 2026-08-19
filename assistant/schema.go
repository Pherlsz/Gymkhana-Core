package assistant

import (
	"encoding/json"
	"errors"

	"github.com/Pherlsz/Gymkhana-Core/portablejson"
)

const (
	// PortableJSONSchemaV1 is retained as an Assistant compatibility alias. The
	// language-neutral contract is owned by the generic portablejson package.
	PortableJSONSchemaV1 = portablejson.SchemaV1
	maxPortableJSONBytes = portablejson.MaxBytes
)

// ValidatePortableJSONObject is retained for compatibility. New generic Core
// consumers should import portablejson directly.
func ValidatePortableJSONObject(raw json.RawMessage) error {
	return mapPortableJSONError(portablejson.ValidateObject(raw), CodeInvalidJSON, "json")
}

// ValidatePortableJSONValue is retained for compatibility. New generic Core
// consumers should import portablejson directly.
func ValidatePortableJSONValue(raw json.RawMessage) error {
	return mapPortableJSONError(portablejson.ValidateValue(raw), CodeInvalidJSON, "json")
}

// ValidatePortableJSONSchema is retained for compatibility. The schema profile
// itself is no longer conceptually owned by Assistant.
func ValidatePortableJSONSchema(raw json.RawMessage) error {
	return mapPortableJSONError(portablejson.ValidateSchema(raw), CodeInvalidSchema, "schema")
}

// ValidatePortableToolSchema validates the stricter object-root schema required
// by tool-call arguments.
func ValidatePortableToolSchema(raw json.RawMessage) error {
	return mapPortableJSONError(portablejson.ValidateObjectSchema(raw), CodeInvalidSchema, "schema")
}

// ValidatePortableJSONInstance is retained for compatibility with the original
// Assistant Foundation API. Generic callers should use portablejson.ValidateInstance.
func ValidatePortableJSONInstance(schemaRaw, instanceRaw json.RawMessage) error {
	return mapPortableJSONError(portablejson.ValidateInstance(schemaRaw, instanceRaw), CodeInvalidSchema, "instance")
}

func mapPortableJSONError(err error, fallback ErrorCode, field string) error {
	if err == nil {
		return nil
	}
	var validation *portablejson.ValidationError
	if errors.As(err, &validation) {
		switch validation.Code {
		case portablejson.CodeInvalidJSON:
			return validationError(CodeInvalidJSON, field)
		case portablejson.CodeInvalidSchema:
			return validationError(CodeInvalidSchema, field)
		}
	}
	return validationError(fallback, field)
}
