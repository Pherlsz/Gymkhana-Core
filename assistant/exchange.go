package assistant

import "github.com/Pherlsz/Gymkhana-Core/portablejson"

// ValidateGenerationToolArguments validates every tool call in a normalized
// provider response against the exact definition exposed in the request. Hosts
// must perform this check before authorizing/executing any tool call.
func ValidateGenerationToolArguments(request GenerationRequest, response GenerationResponse) error {
	if err := ValidateGenerationExchange(request, response); err != nil {
		return err
	}
	compiledSchemas := make(map[string]*portablejson.CompiledSchema, len(request.Tools))
	for _, definition := range request.Tools {
		compiled, err := portablejson.CompileObjectSchema(definition.InputSchema)
		if err != nil {
			return validationError(CodeInvalidSchema, "input_schema")
		}
		compiledSchemas[definition.Name] = compiled
	}
	for _, part := range response.Message.Content {
		if part.Type != PartToolCall {
			continue
		}
		compiled, ok := compiledSchemas[part.ToolCall.Name]
		if !ok {
			return validationError(CodeInvalidToolCall, "generation_response.tool_call.name")
		}
		if err := compiled.ValidateInstance(part.ToolCall.Arguments); err != nil {
			return validationError(CodeInvalidToolCall, "tool_call.arguments")
		}
	}
	return nil
}
