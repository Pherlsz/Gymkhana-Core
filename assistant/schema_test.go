package assistant_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
)

func TestValidatePortableJSONObjectRejectsAmbiguousJSON(t *testing.T) {
	t.Parallel()

	valid := json.RawMessage(`{"city":"Porto Alegre","nested":{"value":1}}`)
	valid = json.RawMessage(strings.ReplaceAll(string(valid), `\"`, `"`))
	if err := assistant.ValidatePortableJSONObject(valid); err != nil {
		t.Fatalf("ValidatePortableJSONObject(valid) = %v", err)
	}

	for name, raw := range map[string]json.RawMessage{
		"duplicate_key":     json.RawMessage(`{"a":1,"a":2}`),
		"unpaired_surrogate": json.RawMessage(`{"value":"\ud800"}`),
		"array_root":         json.RawMessage(`[1,2,3]`),
		"trailing_value":     json.RawMessage(`{} {}`),
	} {
		name, raw := name, raw
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if err := assistant.ValidatePortableJSONObject(raw); err == nil {
				t.Fatalf("ValidatePortableJSONObject(%s) unexpectedly succeeded", raw)
			}
		})
	}
}

func TestValidatePortableJSONSchema(t *testing.T) {
	t.Parallel()

	valid := json.RawMessage(`{
		"type":"object",
		"properties":{
			"name":{"type":"string"},
			"age":{"type":["integer","null"],"minimum":0,"maximum":200}
		},
		"required":["name","age"],
		"additionalProperties":false
	}`)
	if err := assistant.ValidatePortableJSONSchema(valid); err != nil {
		t.Fatalf("ValidatePortableJSONSchema(valid) = %v", err)
	}

	invalidKeyword := json.RawMessage(`{"type":"string","pattern":".*"}`)
	assertAssistantCode(t, assistant.ValidatePortableJSONSchema(invalidKeyword), assistant.CodeInvalidSchema)

	optionalProperty := json.RawMessage(`{
		"type":"object",
		"properties":{"name":{"type":"string"},"age":{"type":"integer"}},
		"required":["name"],
		"additionalProperties":false
	}`)
	assertAssistantCode(t, assistant.ValidatePortableJSONSchema(optionalProperty), assistant.CodeInvalidSchema)
}

func TestToolDefinitionUsesPortableSchema(t *testing.T) {
	t.Parallel()

	definition := assistant.ToolDefinition{
		Name: "lookup",
		InputSchema: json.RawMessage(`{
			"type":"object",
			"properties":{"query":{"type":"string"}},
			"required":["query"],
			"additionalProperties":false
		}`),
	}
	if err := assistant.ValidateToolDefinition(definition); err != nil {
		t.Fatalf("ValidateToolDefinition(valid) = %v", err)
	}
}
