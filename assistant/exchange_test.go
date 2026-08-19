package assistant_test

import (
	"encoding/json"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
)

func TestValidateGenerationToolArguments(t *testing.T) {
	t.Parallel()

	request := assistant.GenerationRequest{
		Model:    assistant.ModelRef{Provider: "openai", Model: "model/latest"},
		Messages: []assistant.Message{{Role: assistant.RoleUser, Content: []assistant.ContentPart{{Type: assistant.PartText, Text: "Weather?"}}}},
		Tools: []assistant.ToolDefinition{{
			Name: "lookup_weather",
			InputSchema: json.RawMessage(`{
				"type":"object",
				"properties":{"city":{"type":"string"}},
				"required":["city"],
				"additionalProperties":false
			}`),
		}},
	}
	response := assistant.GenerationResponse{
		Model: assistant.ModelRef{Provider: "openai", Model: "model/latest"},
		Message: assistant.Message{Role: assistant.RoleAssistant, Content: []assistant.ContentPart{{
			Type: assistant.PartToolCall,
			ToolCall: &assistant.ToolCall{
				ID:        "call_01",
				Name:      "lookup_weather",
				Arguments: json.RawMessage(`{"city":"Porto Alegre"}`),
			},
		}}},
		FinishReason: assistant.FinishToolCalls,
	}
	if err := assistant.ValidateGenerationToolArguments(request, response); err != nil {
		t.Fatalf("ValidateGenerationToolArguments(valid) = %v", err)
	}

	response.Message.Content[0].ToolCall.Arguments = json.RawMessage(`{"city":"Porto Alegre","secret_extra":true}`)
	assertAssistantCode(t, assistant.ValidateGenerationToolArguments(request, response), assistant.CodeInvalidToolCall)
}
