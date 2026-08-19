package assistant

import (
	"encoding/json"
	"strings"
	"unicode/utf8"
)

const maxPortableJSONInteger int64 = 1<<53 - 1

// ValidateMessage validates portable message semantics including the allowed
// tool-content relationship for each semantic role.
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
		switch message.Role {
		case RoleSystem, RoleDeveloper, RoleUser:
			if part.Type == PartToolCall || part.Type == PartToolResult {
				return validationError(CodeInvalidRole, "content.type")
			}
		case RoleAssistant:
			if part.Type == PartToolResult {
				return validationError(CodeInvalidRole, "content.type")
			}
		case RoleTool:
			if part.Type != PartToolResult {
				return validationError(CodeInvalidRole, "content.type")
			}
		}
	}
	return nil
}

// ValidateConversation validates cross-message tool-call/result linkage. Tool
// call IDs are unique within the request, every tool result resolves one prior
// call exactly once, and no unresolved call remains when a new generation is
// requested.
func ValidateConversation(messages []Message) error {
	if len(messages) == 0 {
		return validationError(CodeEmpty, "messages")
	}
	pending := make(map[string]struct{})
	seenCalls := make(map[string]struct{})
	for _, message := range messages {
		if err := ValidateMessage(message); err != nil {
			return err
		}
		for _, part := range message.Content {
			switch part.Type {
			case PartToolCall:
				id := part.ToolCall.ID
				if _, duplicate := seenCalls[id]; duplicate {
					return validationError(CodeInvalidToolCall, "messages.tool_call.id")
				}
				seenCalls[id] = struct{}{}
				pending[id] = struct{}{}
			case PartToolResult:
				id := part.ToolResult.CallID
				if _, exists := pending[id]; !exists {
					return validationError(CodeInvalidToolResult, "messages.tool_result.call_id")
				}
				delete(pending, id)
			}
		}
	}
	if len(pending) > 0 {
		return validationError(CodeUnresolvedToolCall, "messages")
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
	case PartImage, PartAudio, PartVideo, PartFile:
		if part.Text != "" || part.Media == nil || part.ToolCall != nil || part.ToolResult != nil {
			return validationError(CodeInvalidContent, "content")
		}
		return validateMedia(part.Type, *part.Media)
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

func validateMedia(partType PartType, media MediaRef) error {
	if media.URI == "" || !utf8.ValidString(media.URI) || utf8.RuneCountInString(media.URI) > 8192 {
		return validationError(CodeInvalidMedia, "media.uri")
	}
	if media.MediaType != "" {
		if !utf8.ValidString(media.MediaType) || len(media.MediaType) > 256 {
			return validationError(CodeInvalidMedia, "media.media_type")
		}
		lower := strings.ToLower(media.MediaType)
		switch partType {
		case PartImage:
			if !strings.HasPrefix(lower, "image/") {
				return validationError(CodeInvalidMedia, "media.media_type")
			}
		case PartAudio:
			if !strings.HasPrefix(lower, "audio/") {
				return validationError(CodeInvalidMedia, "media.media_type")
			}
		case PartVideo:
			if !strings.HasPrefix(lower, "video/") {
				return validationError(CodeInvalidMedia, "media.media_type")
			}
		}
	}
	if media.Name != "" && (!utf8.ValidString(media.Name) || utf8.RuneCountInString(media.Name) > 512) {
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

// ValidateToolDefinition validates portable tool metadata and the conservative
// portable JSON-schema profile used across provider adapters.
func ValidateToolDefinition(definition ToolDefinition) error {
	if err := ValidateToolName(definition.Name); err != nil {
		return err
	}
	if !utf8.ValidString(definition.Description) || utf8.RuneCountInString(definition.Description) > 4096 {
		return validationError(CodeInvalidContent, "description")
	}
	if err := ValidatePortableJSONSchema(definition.InputSchema); err != nil {
		return validationError(CodeInvalidSchema, "input_schema")
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
	if err := ValidatePortableJSONObject(call.Arguments); err != nil {
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
	return ValidatePortableJSONObject(raw) == nil
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
