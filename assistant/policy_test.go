package assistant_test

import (
	"math"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
)

func TestToolPolicySafetyDefaults(t *testing.T) {
	t.Parallel()

	policy, err := assistant.ResolveToolPolicy("delete_account", nil)
	if err != nil {
		t.Fatalf("ResolveToolPolicy = %v", err)
	}
	if policy.Effect != assistant.ToolEffectUnknown || policy.Confirmation != assistant.ToolConfirmAlways {
		t.Fatalf("unsafe default policy: %#v", policy)
	}

	unsafe := assistant.ToolPolicy{Name: "delete_account", Effect: assistant.ToolEffectDestructive, Confirmation: assistant.ToolConfirmNever}
	assertAssistantCode(t, assistant.ValidateToolPolicy(unsafe), assistant.CodeInvalidToolPolicy)
}

func TestNextModelCandidateDistinguishesModelAndProviderFallback(t *testing.T) {
	t.Parallel()

	candidates := []assistant.ModelRef{{Provider: "openai", Model: "model/a"}, {Provider: "openai", Model: "model/b"}, {Provider: "google", Model: "model/c"}}
	policy := assistant.RoutingFallbackPolicy{ModelOn: []assistant.FailureClass{assistant.FailureContextLimit}, ProviderOn: []assistant.FailureClass{assistant.FailureRateLimit}, MaxAttempts: 4}
	first, index, err := assistant.NextModelCandidate(candidates, -1, 0, assistant.FailureUnknown, policy)
	if err != nil || index != 0 || first != candidates[0] {
		t.Fatalf("first = %#v index=%d err=%v", first, index, err)
	}
	next, index, err := assistant.NextModelCandidate(candidates, 0, 1, assistant.FailureContextLimit, policy)
	if err != nil || index != 1 || next != candidates[1] {
		t.Fatalf("model fallback = %#v index=%d err=%v", next, index, err)
	}
	next, index, err = assistant.NextModelCandidate(candidates, 1, 2, assistant.FailureRateLimit, policy)
	if err != nil || index != 2 || next != candidates[2] {
		t.Fatalf("provider fallback = %#v index=%d err=%v", next, index, err)
	}
	if _, _, err := assistant.NextModelCandidate(candidates, 0, 1, assistant.FailureSafety, policy); !assistant.IsCode(err, assistant.CodeModelUnavailable) {
		t.Fatalf("safety failure must not fallback, got %v", err)
	}
}

func TestRunTraceTotalUsage(t *testing.T) {
	t.Parallel()

	credential := assistant.CredentialIdentity{ID: "key_01", Provider: "openai", Mode: assistant.CredentialBYOK, QuotaScope: "project:primary"}
	trace := assistant.RunTrace{
		RunID: "run-001", AssistantID: "coding_assistant", AssistantRevision: 3,
		EffectiveModules: []assistant.ModuleID{assistant.ModuleText, assistant.ModuleTools},
		Attempts: []assistant.AttemptTrace{
			{Index: 0, Model: assistant.ModelRef{Provider: "openai", Model: "model/a"}, Credential: &credential, Failure: assistant.FailureRateLimit, Usage: assistant.Usage{InputTokens: 10}},
			{Index: 1, Model: assistant.ModelRef{Provider: "openai", Model: "model/b"}, Credential: &credential, Usage: assistant.Usage{InputTokens: 20, OutputTokens: 5}},
		},
		FinishReason: assistant.FinishStop,
	}
	if err := assistant.ValidateRunTrace(trace); err != nil {
		t.Fatalf("ValidateRunTrace(valid) = %v", err)
	}
	total, err := trace.TotalUsage()
	if err != nil {
		t.Fatalf("TotalUsage = %v", err)
	}
	if total.InputTokens != 30 || total.OutputTokens != 5 {
		t.Fatalf("total = %#v", total)
	}
}

func TestRetrievalEvidenceValidation(t *testing.T) {
	t.Parallel()

	evidence := assistant.RetrievalEvidence{ID: "ev-001", Content: "bounded evidence", Source: "document:42", Score: 0.9, Metadata: map[string]string{"tenant": "example"}}
	if err := assistant.ValidateRetrievalEvidence(evidence); err != nil {
		t.Fatalf("ValidateRetrievalEvidence(valid) = %v", err)
	}
	evidence.Score = math.Inf(1)
	assertAssistantCode(t, assistant.ValidateRetrievalEvidence(evidence), assistant.CodeInvalidRetrieval)
}
