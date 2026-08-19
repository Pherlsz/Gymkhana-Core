package conformance

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
)

func TestAssistantConformance(t *testing.T) { runSuite(t, "assistant.json", executeAssistant) }
func TestAssistantMultimodalConformance(t *testing.T) {
	runSuite(t, "assistant-multimodal.json", executeAssistant)
}
func TestAssistantConfigConformance(t *testing.T) {
	runSuite(t, "assistant-config.json", executeAssistantConfig)
}
func TestAssistantHardeningConformance(t *testing.T) {
	runSuite(t, "assistant-hardening.json", executeAssistantHardening)
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
	case "assistant.skill.token_economy.builtin":
		return assistant.BuiltinTokenEconomySkill(), ""
	case "assistant.skills.validate":
		var value []assistant.Skill
		if err := json.Unmarshal(testCase.Input, &value); err != nil {
			return nil, "invalid_conformance_input"
		}
		return assistantValidationResult(assistant.ValidateSkills(value))
	default:
		return nil, "unsupported_operation"
	}
}

func executeAssistantConfig(testCase vector) (any, string) {
	switch testCase.Operation {
	case "assistant.definition.validate":
		var value assistant.AssistantDefinition
		if err := json.Unmarshal(testCase.Input, &value); err != nil {
			return nil, "invalid_conformance_input"
		}
		return assistantValidationResult(assistant.ValidateAssistantDefinition(value))
	case "assistant.credential_ref.validate":
		var value assistant.CredentialRef
		if err := json.Unmarshal(testCase.Input, &value); err != nil {
			return nil, "invalid_conformance_input"
		}
		return assistantValidationResult(assistant.ValidateCredentialRef(value))
	case "assistant.retrieval.validate":
		var value assistant.RetrievalPolicy
		if err := json.Unmarshal(testCase.Input, &value); err != nil {
			return nil, "invalid_conformance_input"
		}
		return assistantValidationResult(assistant.ValidateRetrievalPolicy(value))
	case "assistant.model.resolve":
		var value modelResolutionInput
		if err := json.Unmarshal(testCase.Input, &value); err != nil {
			return nil, "invalid_conformance_input"
		}
		resolved, err := assistant.ResolveModelCandidates(value.Policy, value.Catalog)
		if err != nil {
			return assistantValidationResult(err)
		}
		return resolved, ""
	case "assistant.definition.resolve_models":
		var value assistantResolutionInput
		if err := json.Unmarshal(testCase.Input, &value); err != nil {
			return nil, "invalid_conformance_input"
		}
		resolved, err := assistant.ResolveAssistantModels(value.Assistant, value.Catalog)
		if err != nil {
			return assistantValidationResult(err)
		}
		return resolved, ""
	default:
		return nil, "unsupported_operation"
	}
}

func executeAssistantHardening(testCase vector) (any, string) {
	switch testCase.Operation {
	case "assistant.conversation.validate":
		var value []assistant.Message
		if err := json.Unmarshal(testCase.Input, &value); err != nil {
			return nil, "invalid_conformance_input"
		}
		return assistantValidationResult(assistant.ValidateConversation(value))
	case "assistant.tool_policy.validate":
		var value assistant.ToolPolicy
		if err := json.Unmarshal(testCase.Input, &value); err != nil {
			return nil, "invalid_conformance_input"
		}
		return assistantValidationResult(assistant.ValidateToolPolicy(value))
	case "assistant.quota.validate":
		var value assistant.QuotaState
		if err := json.Unmarshal(testCase.Input, &value); err != nil {
			return nil, "invalid_conformance_input"
		}
		return assistantValidationResult(assistant.ValidateQuotaState(value))
	case "assistant.usage_ledger.validate":
		var value assistant.UsageLedgerEntry
		if err := json.Unmarshal(testCase.Input, &value); err != nil {
			return nil, "invalid_conformance_input"
		}
		return assistantValidationResult(assistant.ValidateUsageLedgerEntry(value))
	case "assistant.trace.validate":
		var value assistant.RunTrace
		if err := json.Unmarshal(testCase.Input, &value); err != nil {
			return nil, "invalid_conformance_input"
		}
		return assistantValidationResult(assistant.ValidateRunTrace(value))
	case "assistant.retrieval_evidence.validate":
		var value assistant.RetrievalEvidence
		if err := json.Unmarshal(testCase.Input, &value); err != nil {
			return nil, "invalid_conformance_input"
		}
		return assistantValidationResult(assistant.ValidateRetrievalEvidence(value))
	case "assistant.credential.next":
		var value credentialFallbackInput
		if err := json.Unmarshal(testCase.Input, &value); err != nil {
			return nil, "invalid_conformance_input"
		}
		credential, index, err := assistant.NextCredential(value.Overrides, value.Provider, value.Current, value.Failure)
		if err != nil {
			return assistantValidationResult(err)
		}
		return credentialFallbackOutput{Credential: credential, Index: index}, ""
	default:
		return nil, "unsupported_operation"
	}
}

type modelResolutionInput struct {
	Policy  assistant.ModelPolicy       `json:"policy"`
	Catalog []assistant.ModelDescriptor `json:"catalog"`
}
type assistantResolutionInput struct {
	Assistant assistant.AssistantDefinition `json:"assistant"`
	Catalog   []assistant.ModelDescriptor   `json:"catalog"`
}
type credentialFallbackInput struct {
	Overrides assistant.RunOverrides `json:"overrides"`
	Provider  assistant.ProviderID   `json:"provider"`
	Current   int                    `json:"current"`
	Failure   assistant.FailureClass `json:"failure"`
}
type credentialFallbackOutput struct {
	Credential assistant.CredentialRef `json:"credential"`
	Index      int                     `json:"index"`
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
