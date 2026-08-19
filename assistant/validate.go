package assistant

import (
	"strings"
	"unicode/utf8"
)

const (
	maxPortableJSONInteger  int64 = 1<<53 - 1
	maxConversationMessages       = 4096
	maxMessageParts               = 1024
	maxTextPartBytes              = 1 << 20
	maxToolResultParts            = 1024
)

func ValidateMessage(message Message) error {
	if !message.Role.Valid() {
		return validationError(CodeInvalidRole, "role")
	}
	if len(message.Content) == 0 {
		return validationError(CodeEmpty, "content")
	}
	if len(message.Content) > maxMessageParts {
		return validationError(CodeInvalidContent, "content")
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

func ValidateConversation(messages []Message) error {
	if len(messages) == 0 {
		return validationError(CodeEmpty, "messages")
	}
	if len(messages) > maxConversationMessages {
		return validationError(CodeInvalidContent, "messages")
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

func ValidateContentPart(part ContentPart) error {
	switch part.Type {
	case PartText:
		if part.Text == "" {
			return validationError(CodeEmpty, "text")
		}
		if !utf8.ValidString(part.Text) || len(part.Text) > maxTextPartBytes || part.Media != nil || part.ToolCall != nil || part.ToolResult != nil {
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

func ValidateToolDefinition(definition ToolDefinition) error {
	if err := ValidateToolName(definition.Name); err != nil {
		return err
	}
	if !utf8.ValidString(definition.Description) || utf8.RuneCountInString(definition.Description) > 4096 {
		return validationError(CodeInvalidContent, "description")
	}
	if err := ValidatePortableToolSchema(definition.InputSchema); err != nil {
		return validationError(CodeInvalidSchema, "input_schema")
	}
	return nil
}

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

func ValidateToolResult(result ToolResult) error {
	if !validCallID(result.CallID) {
		return validationError(CodeInvalidToolResult, "call_id")
	}
	if len(result.Content) == 0 {
		return validationError(CodeEmpty, "content")
	}
	if len(result.Content) > maxToolResultParts {
		return validationError(CodeInvalidToolResult, "content")
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

func ValidateFinishReason(reason FinishReason) error {
	if !reason.Valid() {
		return validationError(CodeInvalidFinishReason, "finish_reason")
	}
	return nil
}

func ValidateUsage(usage Usage) error {
	values := [...]int64{usage.InputTokens, usage.OutputTokens, usage.CachedInputTokens, usage.ReasoningTokens}
	for _, value := range values {
		if value < 0 || value > maxPortableJSONInteger {
			return validationError(CodeInvalidUsage, "usage")
		}
	}
	return nil
}

func ValidateCapabilities(capabilities []Capability) error {
	if len(capabilities) > 64 {
		return validationError(CodeInvalidCapability, "capabilities")
	}
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
