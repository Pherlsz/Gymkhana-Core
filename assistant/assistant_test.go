package assistant_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
)

func TestValidateMessage(t *testing.T) {
	t.Parallel()

	message := assistant.Message{
		Role: assistant.RoleUser,
		Content: []assistant.ContentPart{
			{Type: assistant.PartText, Text: "Compare these documents."},
			{
				Type: assistant.PartFile,
				Media: &assistant.MediaRef{
					URI:       "urn:document:42",
					MediaType: "application/pdf",
					Name:      "contract.pdf",
				},
			},
		},
	}
	if err := assistant.ValidateMessage(message); err != nil {
		t.Fatalf("ValidateMessage(valid) = %v", err)
	}

	message.Role = assistant.Role("provider_specific")
	assertAssistantCode(t, assistant.ValidateMessage(message), assistant.CodeInvalidRole)
}

func TestValidateContentParts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		part assistant.ContentPart
		code assistant.ErrorCode
	}{
		{
			name: "text",
			part: assistant.ContentPart{Type: assistant.PartText, Text: "hello"},
		},
		{
			name: "image",
			part: assistant.ContentPart{Type: assistant.PartImage, Media: &assistant.MediaRef{URI: "urn:image:1"}},
		},
		{
			name: "tool call",
			part: assistant.ContentPart{Type: assistant.PartToolCall, ToolCall: &assistant.ToolCall{
				ID:        "call_01",
				Name:      "lookup_weather",
				Arguments: json.RawMessage("{\"location\":\"Tokyo\"}"),
			}},
		},
		{
			name: "tool result",
			part: assistant.ContentPart{Type: assistant.PartToolResult, ToolResult: &assistant.ToolResult{
				CallID:  "call_01",
				Content: []assistant.ContentPart{{Type: assistant.PartText, Text: "21 C"}},
			}},
		},
		{
			name: "mixed payload rejected",
			part: assistant.ContentPart{Type: assistant.PartText, Text: "hello", Media: &assistant.MediaRef{URI: "urn:image:1"}},
			code: assistant.CodeInvalidContent,
		},
		{
			name: "unknown type rejected",
			part: assistant.ContentPart{Type: assistant.PartType("provider_magic"), Text: "hello"},
			code: assistant.CodeInvalidContentType,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := assistant.ValidateContentPart(test.part)
			if test.code == "" {
				if err != nil {
					t.Fatalf("ValidateContentPart(valid) = %v", err)
				}
				return
			}
			assertAssistantCode(t, err, test.code)
		})
	}
}

func TestValidateToolDefinition(t *testing.T) {
	t.Parallel()

	definition := assistant.ToolDefinition{
		Name:        "lookup_weather",
		Description: "Look up current weather.",
		InputSchema: json.RawMessage(`{}`),
	}
	if err := assistant.ValidateToolDefinition(definition); err != nil {
		t.Fatalf("ValidateToolDefinition(valid) = %v", err)
	}

	definition.Name = "lookup.weather"
	assertAssistantCode(t, assistant.ValidateToolDefinition(definition), assistant.CodeInvalidToolName)

	definition.Name = "lookup_weather"
	definition.InputSchema = json.RawMessage(`[]`)
	assertAssistantCode(t, assistant.ValidateToolDefinition(definition), assistant.CodeInvalidSchema)
}

func TestValidateToolCallRejectsNonObjectArguments(t *testing.T) {
	t.Parallel()

	call := assistant.ToolCall{ID: "call_01", Name: "lookup_weather", Arguments: json.RawMessage(`[]`)}
	assertAssistantCode(t, assistant.ValidateToolCall(call), assistant.CodeInvalidJSON)
}

func TestValidateToolCallIDUsesUnicodeScalarLimit(t *testing.T) {
	t.Parallel()

	call := assistant.ToolCall{
		ID:        strings.Repeat("界", 256),
		Name:      "lookup_weather",
		Arguments: json.RawMessage(`{}`),
	}
	if err := assistant.ValidateToolCall(call); err != nil {
		t.Fatalf("256-scalar call ID = %v", err)
	}
	call.ID += "界"
	assertAssistantCode(t, assistant.ValidateToolCall(call), assistant.CodeInvalidToolCall)
}

func TestValidateUsage(t *testing.T) {
	t.Parallel()

	if err := assistant.ValidateUsage(assistant.Usage{InputTokens: 10, OutputTokens: 3}); err != nil {
		t.Fatalf("ValidateUsage(valid) = %v", err)
	}
	assertAssistantCode(t, assistant.ValidateUsage(assistant.Usage{InputTokens: -1}), assistant.CodeInvalidUsage)
	assertAssistantCode(t, assistant.ValidateUsage(assistant.Usage{InputTokens: 1 << 53}), assistant.CodeInvalidUsage)
}

func TestValidateCapabilities(t *testing.T) {
	t.Parallel()

	if err := assistant.ValidateCapabilities([]assistant.Capability{assistant.CapabilityText, assistant.CapabilityToolCalling}); err != nil {
		t.Fatalf("ValidateCapabilities(valid) = %v", err)
	}
	assertAssistantCode(t, assistant.ValidateCapabilities([]assistant.Capability{assistant.CapabilityText, assistant.CapabilityText}), assistant.CodeDuplicateCapability)
	assertAssistantCode(t, assistant.ValidateCapabilities([]assistant.Capability{assistant.Capability("provider_magic")}), assistant.CodeInvalidCapability)
}

func TestValidationErrorDoesNotEchoSensitiveInput(t *testing.T) {
	t.Parallel()

	secret := "secret-user-value"
	message := assistant.Message{Role: assistant.RoleUser, Content: []assistant.ContentPart{{Type: assistant.PartText, Text: ""}}}
	err := assistant.ValidateMessage(message)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if contains(err.Error(), secret) {
		t.Fatalf("error leaked original input: %q", err)
	}
}

func contains(value, substring string) bool {
	if len(substring) == 0 {
		return true
	}
	for index := 0; index+len(substring) <= len(value); index++ {
		if value[index:index+len(substring)] == substring {
			return true
		}
	}
	return false
}
