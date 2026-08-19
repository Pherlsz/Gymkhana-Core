package conformance

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
)

func TestAssistantConformance(t *testing.T) {
	runSuite(t, "assistant.json", executeAssistant)
}

func executeAssistant(testCase vector) (any, string) {
	switch testCase.Operation {
	case "assistant.message.validate":
		var value assistant.Message
		if err := json.Unmarshal(testCase.Input, &value); err != nil {
			return nil, "invalid_conformance_input"
		}
		return assistantValidationResult(assistant.ValidateMessage(value))
	case "assistant.tool_definition.validate":
		var value assistant.ToolDefinition
		if err := json.Unmarshal(testCase.Input, &value); err != nil {
			return nil, "invalid_conformance_input"
		}
		return assistantValidationResult(assistant.ValidateToolDefinition(value))
	case "assistant.finish_reason.validate":
		input, ok := stringInput(testCase.Input)
		if !ok {
			return nil, "invalid_conformance_input"
		}
		return assistantValidationResult(assistant.ValidateFinishReason(assistant.FinishReason(input)))
	case "assistant.usage.validate":
		var value assistant.Usage
		if err := json.Unmarshal(testCase.Input, &value); err != nil {
			return nil, "invalid_conformance_input"
		}
		return assistantValidationResult(assistant.ValidateUsage(value))
	case "assistant.capabilities.validate":
		var value []assistant.Capability
		if err := json.Unmarshal(testCase.Input, &value); err != nil {
			return nil, "invalid_conformance_input"
		}
		return assistantValidationResult(assistant.ValidateCapabilities(value))
	default:
		return nil, "unsupported_operation"
	}
}

func assistantValidationResult(err error) (any, string) {
	if err == nil {
		return true, ""
	}
	var validation *assistant.ValidationError
	if errors.As(err, &validation) {
		return nil, string(validation.Code)
	}
	return nil, "unknown_error"
}
