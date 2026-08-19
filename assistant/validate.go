package assistant

import (
	"bytes"
	"encoding/json"
	"unicode/utf8"
)

const maxPortableJSONInteger int64 = 1<<53 - 1

// ValidateMessage validates portable message semantics.
func ValidateMessage(message Message) error {
	if !message.Role.Valid() {
		return validationError(CodeInvalidRole, "role")
	}
	if len(message.Content) == 0 {
		return validationError(CodeEmpty, "content")
	}
	for _, part := range message.Content {
		if err := ValidateContentPart(part); err != nil {
			return err
		}
	}
	return nil
}

// ValidateContentPart validates the ContentPart discriminated union.
func ValidateContentPart(part ContentPart) error {
	switch part.Type {
	case PartText:
		if part.Text == "" {
			return validationError(CodeEmpty, "text")
		}
		if !utf8.ValidString(part.Text) || part.Media != nil || part.ToolCall != nil || part.ToolResult != nil {
			return validationError(CodeInvalidContent, "content")
		}
		return nil
	case PartImage, PartFile:
		if part.Text != "" || part.Media == nil || part.ToolCall != nil || part.ToolResult != nil {
			return validationError(CodeInvalidContent, "content")
		}
		return validateMedia(*part.Media)
	case PartToolCall:
		if part.Text != "" || part.Media != nil || part.ToolCall == nil || part.ToolResult != nil {
			return validationError(CodeInvalidContent, "content")
		}
		return ValidateToolCall(*part.ToolCall)
	case PartToolResult:
		if part.Text != "" || part.Media != nil || part.ToolCall != nil || part.ToolResult == nil {
			return validationError(CodeInvalidContent, "content")
		}
		return ValidateToolResult(*part.ToolResult)
	default:
		return validationError(CodeInvalidContentType, "type")
	}
}

func validateMedia(media MediaRef) error {
	if media.URI == "" || !utf8.ValidString(media.URI) {
		return validationError(CodeInvalidMedia, "media.uri")
	}
	if media.MediaType != "" && !utf8.ValidString(media.MediaType) {
		return validationError(CodeInvalidMedia, "media.media_type")
	}
	if media.Name != "" && !utf8.ValidString(media.Name) {
		return validationError(CodeInvalidMedia, "media.name")
	}
	return nil
}

// ValidateToolName validates the portable cross-provider tool-name subset.
func ValidateToolName(name string) error {
	if len(name) == 0 || len(name) > 64 || !utf8.ValidString(name) {
		return validationError(CodeInvalidToolName, "name")
	}
	for i := 0; i < len(name); i++ {
		b := name[i]
		if i == 0 {
			if !asciiLetter(b) && b != '_' {
				return validationError(CodeInvalidToolName, "name")
			}
			continue
		}
		if !asciiLetter(b) && (b < '0' || b > '9') && b != '_' && b != '-' {
			return validationError(CodeInvalidToolName, "name")
		}
	}
	return nil
}

func asciiLetter(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z'
}

// ValidateToolDefinition validates portable tool metadata and the top-level
// shape of its input schema.
func ValidateToolDefinition(definition ToolDefinition) error {
	if err := ValidateToolName(definition.Name); err != nil {
		return err
	}
	if !utf8.ValidString(definition.Description) {
		return validationError(CodeInvalidContent, "description")
	}
	if !validJSONObject(definition.InputSchema) {
		return validationError(CodeInvalidJSON, "input_schema")
	}
	return nil
}

// ValidateToolCall validates call identity, tool name, and structured arguments.
func ValidateToolCall(call ToolCall) error {
	if !validCallID(call.ID) {
		return validationError(CodeInvalidToolCall, "id")
	}
	if err := ValidateToolName(call.Name); err != nil {
		return err
	}
	if !validJSONObject(call.Arguments) {
		return validationError(CodeInvalidJSON, "arguments")
	}
	return nil
}

// ValidateToolResult validates a tool result and forbids recursive tool
// call/result content in the result body.
func ValidateToolResult(result ToolResult) error {
	if !validCallID(result.CallID) {
		return validationError(CodeInvalidToolResult, "call_id")
	}
	if len(result.Content) == 0 {
		return validationError(CodeEmpty, "content")
	}
	for _, part := range result.Content {
		if part.Type == PartToolCall || part.Type == PartToolResult {
			return validationError(CodeInvalidToolResult, "content")
		}
		if err := ValidateContentPart(part); err != nil {
			return err
		}
	}
	return nil
}

func validCallID(value string) bool {
	return value != "" && utf8.ValidString(value) && utf8.RuneCountInString(value) <= 256
}

func validJSONObject(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) < 2 || trimmed[0] != '{' || trimmed[len(trimmed)-1] != '}' || !json.Valid(trimmed) {
		return false
	}
	var object map[string]json.RawMessage
	return json.Unmarshal(trimmed, &object) == nil && object != nil
}

// ValidateFinishReason validates one portable completion reason.
func ValidateFinishReason(reason FinishReason) error {
	if !reason.Valid() {
		return validationError(CodeInvalidFinishReason, "finish_reason")
	}
	return nil
}

// ValidateUsage rejects counters outside the non-negative JSON safe-integer
// range so serialized values remain exact in JavaScript/TypeScript runtimes.
func ValidateUsage(usage Usage) error {
	values := [...]int64{usage.InputTokens, usage.OutputTokens, usage.CachedInputTokens, usage.ReasoningTokens}
	for _, value := range values {
		if value < 0 || value > maxPortableJSONInteger {
			return validationError(CodeInvalidUsage, "usage")
		}
	}
	return nil
}

// ValidateCapabilities validates known portable capability names and rejects
// duplicates so serialized capability sets remain unambiguous.
func ValidateCapabilities(capabilities []Capability) error {
	seen := make(map[Capability]struct{}, len(capabilities))
	for _, capability := range capabilities {
		if !capability.Valid() {
			return validationError(CodeInvalidCapability, "capabilities")
		}
		if _, exists := seen[capability]; exists {
			return validationError(CodeDuplicateCapability, "capabilities")
		}
		seen[capability] = struct{}{}
	}
	return nil
}
