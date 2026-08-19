package assistant_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
)

func TestValidateMessage(t *testing.T) {
	t.Parallel()

	valid := assistant.Message{
		Role: assistant.RoleUser,
		Content: []assistant.ContentPart{
			{Type: assistant.PartText, Text: "Hello"},
			{Type: assistant.PartImage, Media: &assistant.MediaRef{URI: "https://example.test/image.png", MediaType: "image/png"}},
		},
	}
	if err := assistant.ValidateMessage(valid); err != nil {
		t.Fatalf("ValidateMessage(valid) = %v", err)
	}

	invalidRole := valid
	invalidRole.Role = "provider-specific-role"
	assertCode(t, assistant.ValidateMessage(invalidRole), assistant.CodeInvalidRole)

	ambiguous := assistant.Message{
		Role: assistant.RoleUser,
		Content: []assistant.ContentPart{{
			Type:  assistant.PartText,
			Text:  "hello",
			Media: &assistant.MediaRef{URI: "https://example.test/image.png"},
		}},
	}
	assertCode(t, assistant.ValidateMessage(ambiguous), assistant.CodeInvalidContent)
}

func TestToolContracts(t *testing.T) {
	t.Parallel()

	definition := assistant.ToolDefinition{
		Name:        "lookup_weather",
		Description: "Look up weather for one location.",
		InputSchema: json.RawMessage("{}"),
	}
	if err := assistant.ValidateToolDefinition(definition); err != nil {
		t.Fatalf("ValidateToolDefinition(valid) = %v", err)
	}

	definition.Name = "lookup.weather"
	assertCode(t, assistant.ValidateToolDefinition(definition), assistant.CodeInvalidToolName)

	call := assistant.ToolCall{
		ID:        "call_01",
		Name:      "lookup_weather",
		Arguments: json.RawMessage("{}"),
	}
	if err := assistant.ValidateToolCall(call); err != nil {
		t.Fatalf("ValidateToolCall(valid) = %v", err)
	}

	call.Arguments = json.RawMessage("[]")
	assertCode(t, assistant.ValidateToolCall(call), assistant.CodeInvalidJSON)

	result := assistant.ToolResult{
		CallID: "call_01",
		Content: []assistant.ContentPart{{
			Type: assistant.PartText,
			Text: "21 C",
		}},
	}
	if err := assistant.ValidateToolResult(result); err != nil {
		t.Fatalf("ValidateToolResult(valid) = %v", err)
	}

	result.Content = []assistant.ContentPart{{
		Type: assistant.PartToolCall,
		ToolCall: &assistant.ToolCall{
			ID:        "nested",
			Name:      "lookup_weather",
			Arguments: json.RawMessage("{}"),
		},
	}}
	assertCode(t, assistant.ValidateToolResult(result), assistant.CodeInvalidToolResult)
}

func TestPortableEnumsAndUsage(t *testing.T) {
	t.Parallel()

	if err := assistant.ValidateFinishReason(assistant.FinishToolCalls); err != nil {
		t.Fatalf("ValidateFinishReason(valid) = %v", err)
	}
	assertCode(t, assistant.ValidateFinishReason("provider_reason"), assistant.CodeInvalidFinishReason)

	if err := assistant.ValidateUsage(assistant.Usage{InputTokens: 10, OutputTokens: 4, CachedInputTokens: 2}); err != nil {
		t.Fatalf("ValidateUsage(valid) = %v", err)
	}
	assertCode(t, assistant.ValidateUsage(assistant.Usage{InputTokens: -1}), assistant.CodeInvalidUsage)

	capabilities := []assistant.Capability{
		assistant.CapabilityText,
		assistant.CapabilityToolCalling,
		assistant.CapabilityStreaming,
	}
	if err := assistant.ValidateCapabilities(capabilities); err != nil {
		t.Fatalf("ValidateCapabilities(valid) = %v", err)
	}
	assertCode(t, assistant.ValidateCapabilities([]assistant.Capability{assistant.CapabilityText, assistant.CapabilityText}), assistant.CodeDuplicateCapability)
	assertCode(t, assistant.ValidateCapabilities([]assistant.Capability{"provider_magic"}), assistant.CodeInvalidCapability)
}

func TestMessageJSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := assistant.Message{
		Role: assistant.RoleAssistant,
		Content: []assistant.ContentPart{{
			Type: assistant.PartToolCall,
			ToolCall: &assistant.ToolCall{
				ID:        "call_01",
				Name:      "lookup_weather",
				Arguments: json.RawMessage("{}"),
			},
		}},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal = %v", err)
	}
	var decoded assistant.Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal = %v", err)
	}
	if err := assistant.ValidateMessage(decoded); err != nil {
		t.Fatalf("round-trip message invalid: %v", err)
	}
}

func assertCode(t *testing.T, err error, code assistant.ErrorCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error code %q", code)
	}
	var validation *assistant.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("error %T is not ValidationError: %v", err, err)
	}
	if validation.Code != code {
		t.Fatalf("error code = %q, want %q", validation.Code, code)
	}
	if !assistant.IsCode(err, code) {
		t.Fatalf("IsCode(%q) = false", code)
	}
}
