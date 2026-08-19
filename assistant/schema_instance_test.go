package assistant_test

import (
	"encoding/json"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
)

func TestValidatePortableJSONInstance(t *testing.T) {
	t.Parallel()

	schema := json.RawMessage("{\"type\":\"object\",\"properties\":{\"name\":{\"type\":\"string\"},\"age\":{\"type\":[\"integer\",\"null\"],\"minimum\":0,\"maximum\":200},\"tags\":{\"type\":\"array\",\"items\":{\"type\":\"string\"},\"minItems\":0,\"maxItems\":3}},\"required\":[\"name\",\"age\",\"tags\"],\"additionalProperties\":false}")
	valid := json.RawMessage("{\"name\":\"Ana\",\"age\":32,\"tags\":[\"admin\"]}")
	if err := assistant.ValidatePortableJSONInstance(schema, valid); err != nil {
		t.Fatalf("ValidatePortableJSONInstance(valid) = %v", err)
	}

	cases := []json.RawMessage{
		json.RawMessage("{\"name\":\"Ana\",\"tags\":[]}"),
		json.RawMessage("{\"name\":\"Ana\",\"age\":201,\"tags\":[]}"),
		json.RawMessage("{\"name\":\"Ana\",\"age\":32,\"tags\":[],\"extra\":true}"),
		json.RawMessage("{\"name\":\"Ana\",\"age\":9007199254740992,\"tags\":[]}"),
	}
	for _, value := range cases {
		if err := assistant.ValidatePortableJSONInstance(schema, value); err == nil {
			t.Fatalf("ValidatePortableJSONInstance(%s) unexpectedly succeeded", value)
		}
	}
}

func TestValidateToolArguments(t *testing.T) {
	t.Parallel()

	definition := assistant.ToolDefinition{
		Name:        "lookup_weather",
		InputSchema: json.RawMessage("{\"type\":\"object\",\"properties\":{\"city\":{\"type\":\"string\"}},\"required\":[\"city\"],\"additionalProperties\":false}"),
	}
	call := assistant.ToolCall{
		ID:        "call_01",
		Name:      "lookup_weather",
		Arguments: json.RawMessage("{\"city\":\"Porto Alegre\"}"),
	}
	if err := assistant.ValidateToolArguments(definition, call); err != nil {
		t.Fatalf("ValidateToolArguments(valid) = %v", err)
	}

	call.Arguments = json.RawMessage("{\"city\":\"Porto Alegre\",\"units\":\"metric\"}")
	assertAssistantCode(t, assistant.ValidateToolArguments(definition, call), assistant.CodeInvalidToolCall)

	call.Name = "other_tool"
	call.Arguments = json.RawMessage("{\"city\":\"Porto Alegre\"}")
	assertAssistantCode(t, assistant.ValidateToolArguments(definition, call), assistant.CodeInvalidToolCall)
}

func TestValidatePortableJSONValueRejectsDuplicateKeys(t *testing.T) {
	t.Parallel()

	if err := assistant.ValidatePortableJSONValue(json.RawMessage("[1,{\"a\":1,\"a\":2}]")); err == nil {
		t.Fatal("duplicate nested key unexpectedly accepted")
	}
}
