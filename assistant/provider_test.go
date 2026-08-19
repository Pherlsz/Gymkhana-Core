package assistant_test

import (
	"encoding/json"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
)

func TestValidateGenerationRequest(t *testing.T) {
	t.Parallel()

	request := assistant.GenerationRequest{
		Model: assistant.ModelRef{Provider: "openai", Model: "model/latest"},
		Messages: []assistant.Message{{
			Role:    assistant.RoleUser,
			Content: []assistant.ContentPart{{Type: assistant.PartText, Text: "hello"}},
		}},
		Tools: []assistant.ToolDefinition{{
			Name:        "lookup_weather",
			InputSchema: json.RawMessage("{}"),
		}},
		Credential: &assistant.CredentialRef{
			ID:        "key_01",
			Provider:  "openai",
			Mode:      assistant.CredentialBYOK,
			Reference: "session:key_01",
		},
	}
	if err := assistant.ValidateGenerationRequest(request); err != nil {
		t.Fatalf("ValidateGenerationRequest(valid) = %v", err)
	}

	badProvider := request
	badProvider.Credential = &assistant.CredentialRef{ID: "key_02", Provider: "google", Mode: assistant.CredentialBYOK, Reference: "session:key_02"}
	assertAssistantCode(t, assistant.ValidateGenerationRequest(badProvider), assistant.CodeInvalidCredential)
}

func TestValidateConversationToolLifecycle(t *testing.T) {
	t.Parallel()

	messages := []assistant.Message{
		{
			Role:    assistant.RoleUser,
			Content: []assistant.ContentPart{{Type: assistant.PartText, Text: "weather?"}},
		},
		{
			Role: assistant.RoleAssistant,
			Content: []assistant.ContentPart{{
				Type: assistant.PartToolCall,
				ToolCall: &assistant.ToolCall{
					ID:        "call_01",
					Name:      "lookup_weather",
					Arguments: json.RawMessage(`{"city":"Porto Alegre"}`),
				},
			}},
		},
		{
			Role: assistant.RoleTool,
			Content: []assistant.ContentPart{{
				Type: assistant.PartToolResult,
				ToolResult: &assistant.ToolResult{
					CallID:  "call_01",
					Content: []assistant.ContentPart{{Type: assistant.PartText, Text: "21 C"}},
				},
			}},
		},
	}
	if err := assistant.ValidateConversation(messages); err != nil {
		t.Fatalf("ValidateConversation(valid) = %v", err)
	}

	unresolved := messages[:2]
	assertAssistantCode(t, assistant.ValidateConversation(unresolved), assistant.CodeUnresolvedToolCall)

	userToolCall := assistant.Message{
		Role: assistant.RoleUser,
		Content: []assistant.ContentPart{{
			Type:     assistant.PartToolCall,
			ToolCall: &assistant.ToolCall{ID: "call_02", Name: "lookup_weather", Arguments: json.RawMessage("{}")},
		}},
	}
	assertAssistantCode(t, assistant.ValidateMessage(userToolCall), assistant.CodeInvalidRole)
}

func TestValidateGenerationResponse(t *testing.T) {
	t.Parallel()

	response := assistant.GenerationResponse{
		Model: assistant.ModelRef{Provider: "openai", Model: "model/latest"},
		Message: assistant.Message{
			Role:    assistant.RoleAssistant,
			Content: []assistant.ContentPart{{Type: assistant.PartText, Text: "hello"}},
		},
		FinishReason: assistant.FinishStop,
		Usage:        assistant.Usage{InputTokens: 10, OutputTokens: 2},
	}
	if err := assistant.ValidateGenerationResponse(response); err != nil {
		t.Fatalf("ValidateGenerationResponse(valid) = %v", err)
	}

	response.Message.Role = assistant.RoleTool
	assertAssistantCode(t, assistant.ValidateGenerationResponse(response), assistant.CodeInvalidProvider)
}

func TestValidateGenerationExchangeToolAllowlist(t *testing.T) {
	t.Parallel()

	request := assistant.GenerationRequest{
		Model:    assistant.ModelRef{Provider: "openai", Model: "model/latest"},
		Messages: []assistant.Message{{Role: assistant.RoleUser, Content: []assistant.ContentPart{{Type: assistant.PartText, Text: "weather?"}}}},
		Tools:    []assistant.ToolDefinition{{Name: "lookup_weather", InputSchema: json.RawMessage("{}")}},
	}
	response := assistant.GenerationResponse{
		Model: assistant.ModelRef{Provider: "openai", Model: "model/latest"},
		Message: assistant.Message{Role: assistant.RoleAssistant, Content: []assistant.ContentPart{{
			Type: assistant.PartToolCall,
			ToolCall: &assistant.ToolCall{ID: "call_01", Name: "lookup_weather", Arguments: json.RawMessage("{}")},
		}}},
		FinishReason: assistant.FinishToolCalls,
	}
	if err := assistant.ValidateGenerationExchange(request, response); err != nil {
		t.Fatalf("ValidateGenerationExchange(valid) = %v", err)
	}

	response.Message.Content[0].ToolCall.Name = "delete_everything"
	assertAssistantCode(t, assistant.ValidateGenerationExchange(request, response), assistant.CodeInvalidToolCall)
}
