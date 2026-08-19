package assistant

// ValidateToolArguments validates provider-generated tool arguments against the
// exact portable schema exposed in the request before a host considers executing
// the call. Name equality is required so a schema cannot be borrowed from a
// different allowed tool.
func ValidateToolArguments(definition ToolDefinition, call ToolCall) error {
	if err := ValidateToolDefinition(definition); err != nil {
		return err
	}
	if err := ValidateToolCall(call); err != nil {
		return err
	}
	if definition.Name != call.Name {
		return validationError(CodeInvalidToolCall, "tool_call.name")
	}
	if err := ValidatePortableJSONInstance(definition.InputSchema, call.Arguments); err != nil {
		return validationError(CodeInvalidToolCall, "tool_call.arguments")
	}
	return nil
}
