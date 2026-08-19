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
