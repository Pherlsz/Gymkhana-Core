package assistant

import (
	"errors"
	"fmt"
)

// ErrorCode is a stable non-localized Assistant validation code.
type ErrorCode string

const (
	CodeEmpty               ErrorCode = "empty"
	CodeInvalidRole         ErrorCode = "invalid_role"
	CodeInvalidContent      ErrorCode = "invalid_content"
	CodeInvalidContentType  ErrorCode = "invalid_content_type"
	CodeInvalidMedia        ErrorCode = "invalid_media"
	CodeInvalidToolName     ErrorCode = "invalid_tool_name"
	CodeInvalidJSON         ErrorCode = "invalid_json"
	CodeInvalidToolCall     ErrorCode = "invalid_tool_call"
	CodeInvalidToolResult   ErrorCode = "invalid_tool_result"
	CodeInvalidFinishReason ErrorCode = "invalid_finish_reason"
	CodeInvalidUsage        ErrorCode = "invalid_usage"
	CodeInvalidCapability   ErrorCode = "invalid_capability"
	CodeDuplicateCapability ErrorCode = "duplicate_capability"
	CodeInvalidSkill        ErrorCode = "invalid_skill"
	CodeDuplicateSkill      ErrorCode = "duplicate_skill"
)

// ValidationError intentionally contains no original prompt, arguments, URI,
// or other user/provider data.
type ValidationError struct {
	Code  ErrorCode
	Field string
}

func (e *ValidationError) Error() string {
	if e == nil {
		return "assistant: validation error"
	}
	if e.Field == "" {
		return fmt.Sprintf("assistant: %s", e.Code)
	}
	return fmt.Sprintf("assistant: %s: %s", e.Field, e.Code)
}

func validationError(code ErrorCode, field string) error {
	return &ValidationError{Code: code, Field: field}
}

// IsCode reports whether err contains the requested stable validation code.
func IsCode(err error, code ErrorCode) bool {
	var validation *ValidationError
	return errors.As(err, &validation) && validation.Code == code
}
