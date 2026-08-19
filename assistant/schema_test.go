package assistant_test

import (
	"encoding/json"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
)

func TestPortableJSONCompatibilityWrappers(t *testing.T) {
	t.Parallel()

	validObject := json.RawMessage("{\"city\":\"Porto Alegre\"}")
	if err := assistant.ValidatePortableJSONObject(validObject); err != nil {
		t.Fatalf("ValidatePortableJSONObject(valid) = %v", err)
	}
	if err := assistant.ValidatePortableJSONSchema(json.RawMessage("{\"type\":\"string\"}")); err != nil {
		t.Fatalf("ValidatePortableJSONSchema(valid) = %v", err)
	}
	assertAssistantCode(t, assistant.ValidatePortableJSONObject(json.RawMessage("[1,2,3]")), assistant.CodeInvalidJSON)
	assertAssistantCode(t, assistant.ValidatePortableJSONSchema(json.RawMessage("{\"type\":\"string\",\"pattern\":\".*\"}")), assistant.CodeInvalidSchema)
}

func TestToolDefinitionUsesPortableObjectSchema(t *testing.T) {
	t.Parallel()

	definition := assistant.ToolDefinition{
		Name:        "lookup",
		InputSchema: json.RawMessage("{\"type\":\"object\",\"properties\":{\"query\":{\"type\":\"string\"}},\"required\":[\"query\"],\"additionalProperties\":false}"),
	}
	if err := assistant.ValidateToolDefinition(definition); err != nil {
		t.Fatalf("ValidateToolDefinition(valid) = %v", err)
	}

	definition.InputSchema = json.RawMessage("{\"type\":\"string\"}")
	assertAssistantCode(t, assistant.ValidateToolDefinition(definition), assistant.CodeInvalidSchema)
}
