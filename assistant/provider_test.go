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
			Provider:  "openai",
			Mode:      assistant.CredentialBYOK,
			Reference: "session/key_01",
		},
	}
	if err := assistant.ValidateGenerationRequest(request); err != nil {
		t.Fatalf("ValidateGenerationRequest(valid) = %v", err)
	}

	badProvider := request
	badProvider.Credential = &assistant.CredentialRef{Provider: "google", Mode: assistant.CredentialBYOK, Reference: "session/key_02"}
	assertAssistantCode(t, assistant.ValidateGenerationRequest(badProvider), assistant.CodeInvalidCredential)
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
