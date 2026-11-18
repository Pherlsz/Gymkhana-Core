package assistant_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
)

func TestEffectiveModulesAndRunModelOverride(t *testing.T) {
	t.Parallel()

	definition := validAssistantDefinition()
	definition.OptionalModules = []assistant.ModuleID{assistant.ModuleAudioInput}
	overrideModel := assistant.ModelRef{Provider: "gateway", Model: "audio/model:paid"}
	overrides := assistant.RunOverrides{
		EnableModules:  []assistant.ModuleID{assistant.ModuleAudioInput},
		DisableModules: []assistant.ModuleID{assistant.ModuleVision},
		Model:          &overrideModel,
		Credentials: []assistant.CredentialRef{{
			ID:        "key_02",
			Provider:  "gateway",
			Mode:      assistant.CredentialBYOK,
			Reference: "session:key_02",
		}},
	}

	modules, err := assistant.EffectiveModules(definition, overrides)
	if err != nil {
		t.Fatalf("EffectiveModules = %v", err)
	}
	wantModules := []assistant.ModuleID{assistant.ModuleText, assistant.ModuleTools, assistant.ModuleRetrieval, assistant.ModuleAudioInput}
	if !reflect.DeepEqual(modules, wantModules) {
		t.Fatalf("modules = %#v, want %#v", modules, wantModules)
	}

	chain := assistant.CredentialsForProvider(overrides, "gateway")
	if len(chain) != 1 || chain[0].Reference != "session:key_02" {
		t.Fatalf("CredentialsForProvider = %#v", chain)
	}

	catalog := []assistant.ModelDescriptor{{
		Ref:    overrideModel,
		Access: assistant.AccessPaid,
		Roles:  []assistant.ModelRole{assistant.ModelRoleGeneration},
		Capabilities: []assistant.Capability{
			assistant.CapabilityText,
			assistant.CapabilityAudioInput,
			assistant.CapabilityToolCalling,
		},
	}}
	resolved, err := assistant.ResolveAssistantModelsForRun(definition, overrides, catalog)
	if err != nil {
		t.Fatalf("ResolveAssistantModelsForRun = %v", err)
	}
	if len(resolved) != 1 || resolved[0] != overrideModel {
		t.Fatalf("resolved = %#v", resolved)
	}
}

func TestOrderedCredentialFallbackSkipsSharedQuotaScope(t *testing.T) {
	t.Parallel()
	definition := validAssistantDefinition()
	overrides := assistant.RunOverrides{Credentials: []assistant.CredentialRef{
		{ID: "key_01", Provider: "gateway", Mode: assistant.CredentialBYOK, Reference: "session:key_01", QuotaScope: "project:shared"},
		{ID: "key_02", Provider: "gateway", Mode: assistant.CredentialBYOK, Reference: "session:key_02", QuotaScope: "project:shared"},
		{ID: "key_03", Provider: "gateway", Mode: assistant.CredentialBYOK, Reference: "session:key_03", QuotaScope: "project:other"},
	}}
	if err := assistant.ValidateRunOverrides(definition, overrides); err != nil {
		t.Fatalf("ValidateRunOverrides = %v", err)
	}

	first, index, err := assistant.NextCredential(overrides, "gateway", -1, assistant.FailureUnknown)
	if err != nil || index != 0 || first.ID != "key_01" {
		t.Fatalf("first = %#v index=%d err=%v", first, index, err)
	}
	second, index, err := assistant.NextCredential(overrides, "gateway", index, assistant.FailureQuota)
	if err != nil || index != 2 || second.ID != "key_03" {
		t.Fatalf("quota-scope fallback = %#v index=%d err=%v", second, index, err)
	}
	if _, _, err := assistant.NextCredential(overrides, "gateway", 0, assistant.FailureAuth); !assistant.IsCode(err, assistant.CodeCredentialExhausted) {
		t.Fatalf("auth must not fallback, got %v", err)
	}
}

func TestEffectiveModulesRejectsEmptyRun(t *testing.T) {
	t.Parallel()
	definition := validAssistantDefinition()
	definition.Modules = []assistant.ModuleID{assistant.ModuleText}
	definition.Tools = nil
	definition.ToolPolicies = nil
	definition.Retrieval = nil
	definition.OptionalModules = nil
	definition.ModelPolicy = assistant.ModelPolicy{Mode: assistant.ModelDynamic}
	_, err := assistant.EffectiveModules(definition, assistant.RunOverrides{DisableModules: []assistant.ModuleID{assistant.ModuleText}})
	assertAssistantCode(t, err, assistant.CodeInvalidModule)
}

func TestRunOverrideCannotEnableUndeclaredModule(t *testing.T) {
	t.Parallel()
	definition := validAssistantDefinition()
	overrides := assistant.RunOverrides{EnableModules: []assistant.ModuleID{assistant.ModuleAudioInput}}
	assertAssistantCode(t, assistant.ValidateRunOverrides(definition, overrides), assistant.CodeInvalidModule)
}

func TestRunOverrideRejectsDuplicateCredentialIdentity(t *testing.T) {
	t.Parallel()
	definition := validAssistantDefinition()
	first := assistant.CredentialRef{ID: "key_01", Provider: "gateway", Mode: assistant.CredentialBYOK, Reference: "session:first"}
	second := assistant.CredentialRef{ID: "key_01", Provider: "gateway", Mode: assistant.CredentialBYOK, Reference: "session:second"}
	overrides := assistant.RunOverrides{Credentials: []assistant.CredentialRef{first, second}}
	assertAssistantCode(t, assistant.ValidateRunOverrides(definition, overrides), assistant.CodeInvalidCredential)
}

func TestAdapterRegistry(t *testing.T) {
	t.Parallel()
	registry := assistant.NewAdapterRegistry()
	adapter := fakeAdapter{provider: assistant.ProviderOpenAI}
	if err := registry.Register(adapter); err != nil {
		t.Fatalf("Register = %v", err)
	}
	if err := registry.Register(adapter); err == nil || !assistant.IsCode(err, assistant.CodeDuplicateProvider) {
		t.Fatalf("duplicate Register error = %v", err)
	}
	if got := registry.Providers(); !reflect.DeepEqual(got, []assistant.ProviderID{assistant.ProviderOpenAI}) {
		t.Fatalf("Providers = %#v", got)
	}
	if _, ok := registry.Get(assistant.ProviderOpenAI); !ok {
		t.Fatal("registered provider not found")
	}
}

func TestAdapterRegistryRejectsTypedNil(t *testing.T) {
	t.Parallel()
	registry := assistant.NewAdapterRegistry()
	var adapter *pointerFakeAdapter
	assertAssistantCode(t, registry.Register(adapter), assistant.CodeInvalidProvider)
}

func TestNilAdapterRegistryReceiverSafety(t *testing.T) {
	t.Parallel()

	var registry *assistant.AdapterRegistry
	assertAssistantCode(t, registry.Register(fakeAdapter{provider: assistant.ProviderOpenAI}), assistant.CodeInvalidProvider)
	if adapter, ok := registry.Get(assistant.ProviderOpenAI); ok || adapter != nil {
		t.Fatalf("nil registry Get = (%#v, %v), want (nil, false)", adapter, ok)
	}
	if providers := registry.Providers(); providers != nil {
		t.Fatalf("nil registry Providers = %#v, want nil", providers)
	}
	if got := registry.ClassifyFailure(assistant.ProviderOpenAI, context.Canceled); got != assistant.FailureCancelled {
		t.Fatalf("nil registry ClassifyFailure(Canceled) = %q, want %q", got, assistant.FailureCancelled)
	}
	if got := registry.ClassifyFailure(assistant.ProviderOpenAI, context.DeadlineExceeded); got != assistant.FailureTimeout {
		t.Fatalf("nil registry ClassifyFailure(DeadlineExceeded) = %q, want %q", got, assistant.FailureTimeout)
	}
	if got := registry.ClassifyFailure(assistant.ProviderOpenAI, &assistant.ValidationError{Code: assistant.CodeInvalidProvider}); got != assistant.FailureUnknown {
		t.Fatalf("nil registry ClassifyFailure(other) = %q, want %q", got, assistant.FailureUnknown)
	}
	_, err := registry.ListModels(context.Background(), assistant.ProviderOpenAI, nil)
	assertAssistantCode(t, err, assistant.CodeInvalidProvider)

	_, err = registry.Generate(context.Background(), assistant.GenerationRequest{
		Model:    assistant.ModelRef{Provider: assistant.ProviderOpenAI, Model: "gpt-test"},
		Messages: []assistant.Message{{Role: assistant.RoleUser, Content: []assistant.ContentPart{{Type: assistant.PartText, Text: "hi"}}}},
	})
	assertAssistantCode(t, err, assistant.CodeInvalidProvider)

	_, err = registry.Quota(context.Background(), assistant.ProviderOpenAI, nil, nil)
	assertAssistantCode(t, err, assistant.CodeInvalidProvider)
}

func TestUsageLedgerConcurrencyAndIdempotency(t *testing.T) {
	t.Parallel()

	var nilLedger *assistant.UsageLedger
	if ok, err := nilLedger.Record(assistant.UsageLedgerEntry{}); ok || !assistant.IsCode(err, assistant.CodeInvalidQuota) {
		t.Fatalf("nil UsageLedger.Record = (%v, %v)", ok, err)
	}
	if entries := nilLedger.Entries(); entries != nil {
		t.Fatalf("nil UsageLedger.Entries = %#v, want nil", entries)
	}
	if total, err := nilLedger.TotalUsageForAssistant("coding_assistant"); err != nil || total != (assistant.Usage{}) {
		t.Fatalf("nil UsageLedger.TotalUsageForAssistant = (%#v, %v)", total, err)
	}

	ledger := assistant.NewUsageLedger()
	cred := assistant.CredentialIdentity{ID: "key_01", Provider: "openai", Mode: assistant.CredentialBYOK, QuotaScope: "org:main"}
	entry := assistant.UsageLedgerEntry{
		RunID:       "run-100",
		Attempt:     0,
		AssistantID: "coding_assistant",
		Credential:  &cred,
		Model:       assistant.ModelRef{Provider: "openai", Model: "gpt-test"},
		Usage:       assistant.Usage{InputTokens: 50, OutputTokens: 10},
	}

	inserted, err := ledger.Record(entry)
	if err != nil || !inserted {
		t.Fatalf("first Record = (%v, %v), want (true, nil)", inserted, err)
	}
	// Mutating caller's credential after Record must not mutate ledger state.
	cred.QuotaScope = "org:mutated"

	replayCred := assistant.CredentialIdentity{ID: "key_01", Provider: "openai", Mode: assistant.CredentialBYOK, QuotaScope: "org:main"}
	replay := entry
	replay.Credential = &replayCred
	inserted, err = ledger.Record(replay)
	if err != nil || inserted {
		t.Fatalf("idempotent replay Record = (%v, %v), want (false, nil)", inserted, err)
	}

	conflict := replay
	conflict.Usage.OutputTokens = 999
	if _, err := ledger.Record(conflict); !assistant.IsCode(err, assistant.CodeInvalidQuota) {
		t.Fatalf("conflicting replay error = %v, want CodeInvalidQuota", err)
	}

	entries := ledger.Entries()
	if len(entries) != 1 || entries[0].Credential.QuotaScope != "org:main" {
		t.Fatalf("Entries = %#v", entries)
	}
	entries[0].Credential.QuotaScope = "org:tampered"
	if ledger.Entries()[0].Credential.QuotaScope != "org:main" {
		t.Fatal("Entries() leaked internal pointer aliasing")
	}

	total, err := ledger.TotalUsageForAssistant("coding_assistant")
	if err != nil || total.InputTokens != 50 || total.OutputTokens != 10 {
		t.Fatalf("TotalUsageForAssistant = (%#v, %v)", total, err)
	}
}

type fakeAdapter struct {
	provider assistant.ProviderID
}

func (adapter fakeAdapter) Descriptor() assistant.ProviderDescriptor {
	return assistant.ProviderDescriptor{
		ID:              adapter.provider,
		DisplayName:     "Fake provider",
		CredentialModes: []assistant.CredentialMode{assistant.CredentialManaged, assistant.CredentialBYOK},
	}
}
func (adapter fakeAdapter) ListModels(context.Context, *assistant.CredentialRef) ([]assistant.ModelDescriptor, error) {
	return nil, nil
}
func (adapter fakeAdapter) Generate(context.Context, assistant.GenerationRequest) (assistant.GenerationResponse, error) {
	return assistant.GenerationResponse{}, nil
}

type pointerFakeAdapter struct{}

func (*pointerFakeAdapter) Descriptor() assistant.ProviderDescriptor {
	return assistant.ProviderDescriptor{ID: "pointer", CredentialModes: []assistant.CredentialMode{assistant.CredentialNone}}
}
func (*pointerFakeAdapter) ListModels(context.Context, *assistant.CredentialRef) ([]assistant.ModelDescriptor, error) {
	return nil, nil
}
func (*pointerFakeAdapter) Generate(context.Context, assistant.GenerationRequest) (assistant.GenerationResponse, error) {
	return assistant.GenerationResponse{}, nil
}
