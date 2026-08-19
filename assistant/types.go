package assistant

import "encoding/json"

// Role identifies the semantic authority/source of a message independently of
// any provider's native role names.
type Role string

const (
	RoleSystem    Role = "system"
	RoleDeveloper Role = "developer"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

func (r Role) Valid() bool {
	switch r {
	case RoleSystem, RoleDeveloper, RoleUser, RoleAssistant, RoleTool:
		return true
	default:
		return false
	}
}

// PartType identifies the payload carried by a ContentPart.
type PartType string

const (
	PartText       PartType = "text"
	PartImage      PartType = "image"
	PartAudio      PartType = "audio"
	PartVideo      PartType = "video"
	PartFile       PartType = "file"
	PartToolCall   PartType = "tool_call"
	PartToolResult PartType = "tool_result"
)

// Message is an ordered provider-neutral assistant message.
type Message struct {
	Role    Role          `json:"role"`
	Content []ContentPart `json:"content"`
}

// ContentPart is a discriminated union. Exactly the payload matching Type is
// valid; ValidateContentPart enforces that invariant.
type ContentPart struct {
	Type       PartType    `json:"type"`
	Text       string      `json:"text,omitempty"`
	Media      *MediaRef   `json:"media,omitempty"`
	ToolCall   *ToolCall   `json:"tool_call,omitempty"`
	ToolResult *ToolResult `json:"tool_result,omitempty"`
}

// MediaRef is an opaque media location plus optional portable metadata. Core
// does not fetch or authorize the URI.
type MediaRef struct {
	URI       string `json:"uri"`
	MediaType string `json:"media_type,omitempty"`
	Name      string `json:"name,omitempty"`
}

// ToolDefinition describes one callable tool. InputSchema uses the conservative
// portable schema profile defined by ValidatePortableJSONSchema.
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// ToolCall is a provider-neutral request to invoke a tool.
type ToolCall struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ToolResult is the result of a prior tool call. Result content may contain
// text/image/audio/video/file parts but may not recursively contain tool calls/results.
type ToolResult struct {
	CallID  string        `json:"call_id"`
	Content []ContentPart `json:"content"`
	IsError bool          `json:"is_error,omitempty"`
}

// FinishReason is a portable completion reason. Refusal is distinct from a
// transport/provider error, and Paused represents a resumable provider turn.
type FinishReason string

const (
	FinishStop          FinishReason = "stop"
	FinishLength        FinishReason = "length"
	FinishToolCalls     FinishReason = "tool_calls"
	FinishRefusal       FinishReason = "refusal"
	FinishPaused        FinishReason = "paused"
	FinishContentFilter FinishReason = "content_filter"
	FinishError         FinishReason = "error"
	FinishOther         FinishReason = "other"
)

func (r FinishReason) Valid() bool {
	switch r {
	case FinishStop, FinishLength, FinishToolCalls, FinishRefusal, FinishPaused, FinishContentFilter, FinishError, FinishOther:
		return true
	default:
		return false
	}
}

// Usage contains provider-reported token counters. Tokenization semantics are
// provider/model-defined; Core only standardizes the portable counter names.
type Usage struct {
	InputTokens       int64 `json:"input_tokens,omitempty"`
	OutputTokens      int64 `json:"output_tokens,omitempty"`
	CachedInputTokens int64 `json:"cached_input_tokens,omitempty"`
	ReasoningTokens   int64 `json:"reasoning_tokens,omitempty"`
}

// Capability identifies a portable model/provider capability.
type Capability string

const (
	CapabilityText             Capability = "text"
	CapabilityImageInput       Capability = "image_input"
	CapabilityImageOutput      Capability = "image_output"
	CapabilityAudioInput       Capability = "audio_input"
	CapabilityAudioOutput      Capability = "audio_output"
	CapabilityVideoInput       Capability = "video_input"
	CapabilityVideoOutput      Capability = "video_output"
	CapabilityFileInput        Capability = "file_input"
	CapabilityFileOutput       Capability = "file_output"
	CapabilityToolCalling      Capability = "tool_calling"
	CapabilityStructuredOutput Capability = "structured_output"
	CapabilityStreaming        Capability = "streaming"
)

func (c Capability) Valid() bool {
	switch c {
	case CapabilityText, CapabilityImageInput, CapabilityImageOutput, CapabilityAudioInput, CapabilityAudioOutput, CapabilityVideoInput, CapabilityVideoOutput, CapabilityFileInput, CapabilityFileOutput, CapabilityToolCalling, CapabilityStructuredOutput, CapabilityStreaming:
		return true
	default:
		return false
	}
}
