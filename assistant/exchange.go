package assistant

// ValidateGenerationToolArguments validates every tool call in a normalized
// provider response against the exact definition exposed in the request. Hosts
// must perform this check before authorizing/executing any tool call.
func ValidateGenerationToolArguments(request GenerationRequest, response GenerationResponse) error {
	if err := ValidateGenerationExchange(request, response); err != nil {
		return err
	}
	definitions := make(map[string]ToolDefinition, len(request.Tools))
	for _, definition := range request.Tools {
		definitions[definition.Name] = definition
	}
	for _, part := range response.Message.Content {
		if part.Type != PartToolCall {
			continue
		}
		definition, ok := definitions[part.ToolCall.Name]
		if !ok {
			return validationError(CodeInvalidToolCall, "generation_response.tool_call.name")
		}
		if err := ValidateToolArguments(definition, *part.ToolCall); err != nil {
			return err
		}
	}
	return nil
}
