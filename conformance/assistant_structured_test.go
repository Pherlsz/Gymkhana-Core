package conformance

import (
	"encoding/json"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
)

func TestAssistantStructuredConformance(t *testing.T) {
	runSuite(t, "assistant-structured.json", executeAssistantStructured)
}

func executeAssistantStructured(testCase vector) (any, string) {
	switch testCase.Operation {
	case "assistant.tool_arguments.validate":
		var value toolArgumentsInput
		if err := json.Unmarshal(testCase.Input, &value); err != nil {
			return nil, "invalid_conformance_input"
		}
		return assistantValidationResult(assistant.ValidateToolArguments(value.Definition, value.Call))
	case "assistant.model.next":
		var value modelFallbackInput
		if err := json.Unmarshal(testCase.Input, &value); err != nil {
			return nil, "invalid_conformance_input"
		}
		model, index, err := assistant.NextModelCandidate(value.Candidates, value.Current, value.Attempts, value.Failure, value.Policy)
		if err != nil {
			return assistantValidationResult(err)
		}
		return modelFallbackOutput{Model: model, Index: index}, ""
	default:
		return nil, "unsupported_operation"
	}
}

type toolArgumentsInput struct {
	Definition assistant.ToolDefinition `json:"definition"`
	Call       assistant.ToolCall       `json:"call"`
}

type modelFallbackInput struct {
	Candidates []assistant.ModelRef            `json:"candidates"`
	Current    int                             `json:"current"`
	Attempts   int64                           `json:"attempts"`
	Failure    assistant.FailureClass          `json:"failure"`
	Policy     assistant.RoutingFallbackPolicy `json:"policy"`
}

type modelFallbackOutput struct {
	Model assistant.ModelRef `json:"model"`
	Index int                `json:"index"`
}
