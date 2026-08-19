package assistant

import (
	"bytes"
	"encoding/json"
)

// ValidatePortableToolSchema validates the portable schema profile and requires
// an object-root schema because ToolCall.Arguments is always a JSON object. `{}`
// remains the explicit unconstrained-object compatibility form.
func ValidatePortableToolSchema(raw json.RawMessage) error {
	if err := ValidatePortableJSONSchema(raw); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var root map[string]any
	if err := decoder.Decode(&root); err != nil {
		return validationError(CodeInvalidSchema, "tool_schema")
	}
	if len(root) == 0 {
		return nil
	}
	types, err := portableSchemaTypes(root["type"])
	if err != nil || len(types) == 0 || types[0] != "object" || len(types) != 1 {
		return validationError(CodeInvalidSchema, "tool_schema.type")
	}
	return nil
}
