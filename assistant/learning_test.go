package assistant_test

import (
	"reflect"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
)

func TestSkillLearningProposal(t *testing.T) {
	t.Parallel()

	task := assistant.CanonicalTaskSignature("document_research", 1, []string{"pt_br", "legal", "pt_br"})
	wantTraits := []string{"legal", "pt_br"}
	if !reflect.DeepEqual(task.Traits, wantTraits) {
		t.Fatalf("traits = %#v, want %#v", task.Traits, wantTraits)
	}

	proposal := assistant.SkillLearningProposal{
		ID:          "learn/retrieval_01",
		AssistantID: "multimodal_researcher",
		Task:        task,
		Scope:       assistant.LearningRetrievalHint,
		Summary:     "Prefer hybrid retrieval for this task family before reranking.",
		Revision:    1,
		Evidence: []assistant.LearningEvidence{
			{RunID: "run-001", Outcome: assistant.LearningImproved, Origin: assistant.LearningOriginEvaluation},
			{RunID: "run-002", Outcome: assistant.LearningSuccess, Origin: assistant.LearningOriginRuntime},
		},
	}
	if err := assistant.ValidateSkillLearningProposal(proposal); err != nil {
		t.Fatalf("ValidateSkillLearningProposal = %v", err)
	}
	eligible, err := assistant.LearningEligible(assistant.DefaultLearningPolicy(), proposal)
	if err != nil || !eligible {
		t.Fatalf("LearningEligible = %v, %v", eligible, err)
	}

	auto := assistant.LearningPolicy{
		Mode:              assistant.LearningAutoPromote,
		MinEvidence:       2,
		MaxEvidence:       32,
		AutoPromoteScopes: []assistant.LearningScope{assistant.LearningRetrievalHint},
	}
	eligible, err = assistant.LearningEligible(auto, proposal)
	if err != nil || !eligible {
		t.Fatalf("auto LearningEligible = %v, %v", eligible, err)
	}

	auto.AutoPromoteScopes = []assistant.LearningScope{assistant.LearningSkillHint}
	assertAssistantCode(t, assistant.ValidateLearningPolicy(auto), assistant.CodeInvalidLearning)
}

func TestQuotaAndUsageLedger(t *testing.T) {
	t.Parallel()

	credential := assistant.CredentialRef{
		ID:         "key_01",
		Provider:   "openai",
		Mode:       assistant.CredentialBYOK,
		Reference:  "session:key_01",
		QuotaScope: "project:primary",
	}
	identity := credential.Identity()
	quota := assistant.QuotaState{
		Provider:   "openai",
		Credential: &identity,
		Known:      true,
		Limits: []assistant.QuotaWindow{{
			Dimension:     assistant.QuotaTotalTokens,
			Limit:         1000,
			Used:          250,
			Remaining:     750,
			WindowSeconds: 60,
		}},
	}
	if err := assistant.ValidateQuotaState(quota); err != nil {
		t.Fatalf("ValidateQuotaState = %v", err)
	}

	entry := assistant.UsageLedgerEntry{
		RunID:       "run-001",
		Attempt:     0,
		AssistantID: "coding_assistant",
		Credential:  &identity,
		Model:       assistant.ModelRef{Provider: "openai", Model: "model/latest"},
		Usage:       assistant.Usage{InputTokens: 100, OutputTokens: 20},
	}
	if err := assistant.ValidateUsageLedgerEntry(entry); err != nil {
		t.Fatalf("ValidateUsageLedgerEntry = %v", err)
	}
	if entry.Credential == nil || entry.Credential.ID != "key_01" {
		t.Fatalf("credential identity = %#v", entry.Credential)
	}
}
