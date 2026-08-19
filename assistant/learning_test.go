package assistant_test

import (
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
)

func TestSkillLearningProposal(t *testing.T) {
	t.Parallel()
	proposal := assistant.SkillLearningProposal{
		ID:          "learn/retrieval_01",
		AssistantID: "multimodal_researcher",
		Scope:       assistant.LearningRetrievalHint,
		Summary:     "Prefer hybrid retrieval for this task family before reranking.",
		Revision:    1,
		Evidence: []assistant.LearningEvidence{{
			RunID:   "run-001",
			Outcome: "success",
		}},
	}
	if err := assistant.ValidateSkillLearningProposal(proposal); err != nil {
		t.Fatalf("ValidateSkillLearningProposal = %v", err)
	}
	proposal.Evidence = nil
	assertAssistantCode(t, assistant.ValidateSkillLearningProposal(proposal), assistant.CodeInvalidLearning)
}

func TestQuotaAndUsageLedger(t *testing.T) {
	t.Parallel()
	credential := assistant.CredentialRef{Provider: "openai", Mode: assistant.CredentialBYOK, Reference: "session/key_01"}
	quota := assistant.QuotaState{Credential: credential, AssistantID: "coding_assistant", Known: true, TokenLimit: 1000, TokensUsed: 250, TokensRemain: 750}
	if err := assistant.ValidateQuotaState(quota); err != nil {
		t.Fatalf("ValidateQuotaState = %v", err)
	}
	entry := assistant.UsageLedgerEntry{
		AssistantID: "coding_assistant",
		Credential:  credential,
		Model:       assistant.ModelRef{Provider: "openai", Model: "model/latest"},
		Usage:       assistant.Usage{InputTokens: 100, OutputTokens: 20},
	}
	if err := assistant.ValidateUsageLedgerEntry(entry); err != nil {
		t.Fatalf("ValidateUsageLedgerEntry = %v", err)
	}
}
