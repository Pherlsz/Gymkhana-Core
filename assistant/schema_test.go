package assistant_test

import (
	"encoding/json"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
)

func TestValidatePortableJSONObjectRejectsAmbiguousJSON(t *testing.T) {
	t.Parallel()

	valid := json.RawMessage(`{"city":"Porto Alegre","nested":{"value":1}}`)
	valid = json.RawMessage([]byte{123, 34, 99, 105, 116, 121, 34, 58, 34, 80, 111, 114, 116, 111, 32, 65, 108, 101, 103, 114, 101, 34, 44, 34, 110, 101, 115, 116, 101, 100, 34, 58, 123, 34, 118, 97, 108, 117, 101, 34, 58, 49, 125, 125})
	if err := assistant.ValidatePortableJSONObject(valid); err != nil {
		t.Fatalf("ValidatePortableJSONObject(valid) = %v", err)
	}

	for name, raw := range map[string]json.RawMessage{
		"duplicate_key":      json.RawMessage([]byte(`{"a":1,"a":2}`)),
		"unpaired_surrogate": json.RawMessage([]byte(`{"value":"\ud800"}`)),
		"array_root":         json.RawMessage(`[1,2,3]`),
		"trailing_value":     json.RawMessage(`{} {}`),
	} {
		name, raw := name, raw
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Turn escaped fixture notation into the exact bytes under test without
			// relying on Go string-literal unescaping semantics.
			if name == "duplicate_key" {
				raw = json.RawMessage([]byte{123, 34, 97, 34, 58, 49, 44, 34, 97, 34, 58, 50, 125})
			}
			if name == "unpaired_surrogate" {
				raw = json.RawMessage([]byte{123, 34, 118, 97, 108, 117, 101, 34, 58, 34, 92, 117, 100, 56, 48, 48, 34, 125})
			}
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
	invalidKeyword = json.RawMessage([]byte{123, 34, 116, 121, 112, 101, 34, 58, 34, 115, 116, 114, 105, 110, 103, 34, 44, 34, 112, 97, 116, 116, 101, 114, 110, 34, 58, 34, 46, 42, 34, 125})
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
