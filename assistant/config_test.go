package assistant_test

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
)

func TestValidateAssistantDefinition(t *testing.T) {
	t.Parallel()

	definition := validAssistantDefinition()
	if err := assistant.ValidateAssistantDefinition(definition); err != nil {
		t.Fatalf("ValidateAssistantDefinition(valid) = %v", err)
	}

	duplicateModule := definition
	duplicateModule.Modules = append(append([]assistant.ModuleID{}, definition.Modules...), assistant.ModuleText)
	assertAssistantCode(t, assistant.ValidateAssistantDefinition(duplicateModule), assistant.CodeDuplicateModule)

	missingRetrievalModule := definition
	missingRetrievalModule.Modules = []assistant.ModuleID{assistant.ModuleText, assistant.ModuleTools}
	assertAssistantCode(t, assistant.ValidateAssistantDefinition(missingRetrievalModule), assistant.CodeInvalidAssistant)

	badInstruction := definition
	badInstruction.Instructions = []assistant.InstructionBlock{{Role: assistant.RoleUser, Text: "not allowed"}}
	assertAssistantCode(t, assistant.ValidateAssistantDefinition(badInstruction), assistant.CodeInvalidAssistant)
}

func TestValidateAssistantCatalog(t *testing.T) {
	t.Parallel()

	first := validAssistantDefinition()
	second := validAssistantDefinition()
	second.ID = "researcher"
	second.Name = "Researcher"

	if err := assistant.ValidateAssistantCatalog(assistant.AssistantCatalog{Assistants: []assistant.AssistantDefinition{first, second}}); err != nil {
		t.Fatalf("ValidateAssistantCatalog(valid) = %v", err)
	}

	second.ID = first.ID
	assertAssistantCode(t, assistant.ValidateAssistantCatalog(assistant.AssistantCatalog{Assistants: []assistant.AssistantDefinition{first, second}}), assistant.CodeDuplicateAssistant)
}

func TestResolveAssistantModels(t *testing.T) {
	t.Parallel()

	definition := validAssistantDefinition()
	catalog := []assistant.ModelDescriptor{
		{
			Ref:          assistant.ModelRef{Provider: "gateway", Model: "basic/text:free"},
			Access:       assistant.AccessFree,
			Roles:        []assistant.ModelRole{assistant.ModelRoleGeneration},
			Capabilities: []assistant.Capability{assistant.CapabilityText},
		},
		{
			Ref:    assistant.ModelRef{Provider: "gateway", Model: "vision/pro:paid"},
			Access: assistant.AccessPaid,
			Roles:  []assistant.ModelRole{assistant.ModelRoleGeneration},
			Capabilities: []assistant.Capability{
				assistant.CapabilityText,
				assistant.CapabilityImageInput,
				assistant.CapabilityToolCalling,
			},
		},
		{
			Ref:    assistant.ModelRef{Provider: "gateway", Model: "vision/fast:free"},
			Access: assistant.AccessFree,
			Roles:  []assistant.ModelRole{assistant.ModelRoleGeneration},
			Capabilities: []assistant.Capability{
				assistant.CapabilityText,
				assistant.CapabilityImageInput,
				assistant.CapabilityToolCalling,
			},
		},
	}

	resolved, err := assistant.ResolveAssistantModels(definition, catalog)
	if err != nil {
		t.Fatalf("ResolveAssistantModels = %v", err)
	}
	want := []assistant.ModelRef{
		{Provider: "gateway", Model: "vision/fast:free"},
		{Provider: "gateway", Model: "vision/pro:paid"},
	}
	if !reflect.DeepEqual(resolved, want) {
		t.Fatalf("resolved = %#v, want %#v", resolved, want)
	}
}

func TestResolveAssistantModelsConcurrentNoMutation(t *testing.T) {
	t.Parallel()

	required := make([]assistant.Capability, 1, 8)
	required[0] = assistant.CapabilityText

	definition := validAssistantDefinition()
	definition.OptionalModules = []assistant.ModuleID{assistant.ModuleAudioInput}
	definition.ModelPolicy.RequiredCapabilities = required

	overrides := assistant.RunOverrides{
		EnableModules: []assistant.ModuleID{assistant.ModuleAudioInput},
	}
	catalog := []assistant.ModelDescriptor{{
		Ref:    assistant.ModelRef{Provider: "gateway", Model: "multimodal/full:free"},
		Access: assistant.AccessFree,
		Roles:  []assistant.ModelRole{assistant.ModelRoleGeneration},
		Capabilities: []assistant.Capability{
			assistant.CapabilityText,
			assistant.CapabilityImageInput,
			assistant.CapabilityAudioInput,
			assistant.CapabilityToolCalling,
		},
	}}

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			if _, err := assistant.ResolveAssistantModels(definition, catalog); err != nil {
				t.Errorf("ResolveAssistantModels = %v", err)
			}
		}()
		go func() {
			defer wg.Done()
			if _, err := assistant.ResolveAssistantModelsForRun(definition, overrides, catalog); err != nil {
				t.Errorf("ResolveAssistantModelsForRun = %v", err)
			}
		}()
	}
	wg.Wait()

	if len(definition.ModelPolicy.RequiredCapabilities) != 1 || definition.ModelPolicy.RequiredCapabilities[0] != assistant.CapabilityText {
		t.Fatalf("definition.ModelPolicy.RequiredCapabilities mutated: %#v", definition.ModelPolicy.RequiredCapabilities)
	}
	if full := required[:cap(required)]; full[1] != "" {
		t.Fatalf("spare capacity in caller slice was overwritten: %#v", full)
	}
}

func TestResolveModelRoleValidatesEntireCatalogAndRejectsDuplicates(t *testing.T) {
	t.Parallel()

	ref := assistant.ModelRef{Provider: "gateway", Model: "embed/v1"}
	validEmbed := assistant.ModelDescriptor{
		Ref:    ref,
		Access: assistant.AccessPaid,
		Roles:  []assistant.ModelRole{assistant.ModelRoleEmbedding},
	}

	duplicateCatalog := []assistant.ModelDescriptor{validEmbed, validEmbed}
	assertAssistantCode(t, func() error {
		_, err := assistant.ResolveModelRole(ref, assistant.ModelRoleEmbedding, duplicateCatalog)
		return err
	}(), assistant.CodeInvalidModel)

	trailingInvalidCatalog := []assistant.ModelDescriptor{
		validEmbed,
		{
			Ref:    assistant.ModelRef{Provider: "gateway", Model: "broken model id"},
			Access: assistant.AccessPaid,
			Roles:  []assistant.ModelRole{assistant.ModelRoleGeneration},
		},
	}
	assertAssistantCode(t, func() error {
		_, err := assistant.ResolveModelRole(ref, assistant.ModelRoleEmbedding, trailingInvalidCatalog)
		return err
	}(), assistant.CodeInvalidModel)

	resolved, err := assistant.ResolveModelRole(ref, assistant.ModelRoleEmbedding, []assistant.ModelDescriptor{validEmbed})
	if err != nil {
		t.Fatalf("ResolveModelRole(valid) = %v", err)
	}
	resolved.Roles[0] = assistant.ModelRoleGeneration
	if validEmbed.Roles[0] != assistant.ModelRoleEmbedding {
		t.Fatal("ResolveModelRole returned aliased Roles slice")
	}
}

func TestModelIdentityDoesNotCollideOnSlash(t *testing.T) {
	t.Parallel()

	policy := assistant.ModelPolicy{Mode: assistant.ModelDynamic}
	catalog := []assistant.ModelDescriptor{
		{
			Ref:          assistant.ModelRef{Provider: "a", Model: "b/c"},
			Access:       assistant.AccessFree,
			Roles:        []assistant.ModelRole{assistant.ModelRoleGeneration},
			Capabilities: []assistant.Capability{assistant.CapabilityText},
		},
		{
			Ref:          assistant.ModelRef{Provider: "a/b", Model: "c"},
			Access:       assistant.AccessFree,
			Roles:        []assistant.ModelRole{assistant.ModelRoleGeneration},
			Capabilities: []assistant.Capability{assistant.CapabilityText},
		},
	}
	resolved, err := assistant.ResolveModelCandidates(policy, catalog)
	if err != nil {
		t.Fatalf("ResolveModelCandidates = %v", err)
	}
	if len(resolved) != 2 || resolved[0] == resolved[1] {
		t.Fatalf("collision in resolved models: %#v", resolved)
	}
}

func TestManualModelSelection(t *testing.T) {
	t.Parallel()

	policy := assistant.ModelPolicy{
		Mode:   assistant.ModelManual,
		Manual: &assistant.ModelRef{Provider: "local", Model: "model:latest"},
	}
	catalog := []assistant.ModelDescriptor{{
		Ref:          assistant.ModelRef{Provider: "local", Model: "model:latest"},
		Access:       assistant.AccessLocal,
		Roles:        []assistant.ModelRole{assistant.ModelRoleGeneration},
		Capabilities: []assistant.Capability{assistant.CapabilityText},
	}}
	resolved, err := assistant.ResolveModelCandidates(policy, catalog)
	if err != nil {
		t.Fatalf("ResolveModelCandidates = %v", err)
	}
	if len(resolved) != 1 || resolved[0] != *policy.Manual {
		t.Fatalf("resolved = %#v", resolved)
	}
}

func TestCredentialAndRetrievalValidation(t *testing.T) {
	t.Parallel()

	credentials := assistant.CredentialPolicy{AllowedModes: []assistant.CredentialMode{assistant.CredentialManaged, assistant.CredentialBYOK}}
	if err := assistant.ValidateCredentialPolicy(credentials); err != nil {
		t.Fatalf("ValidateCredentialPolicy = %v", err)
	}
	validCredential := assistant.CredentialRef{
		ID:        "key_01",
		Provider:  "openai",
		Mode:      assistant.CredentialBYOK,
		Reference: "session:key_01",
	}
	if err := assistant.ValidateCredentialRef(validCredential); err != nil {
		t.Fatalf("ValidateCredentialRef = %v", err)
	}

	rawLooking := validCredential
	rawLooking.Reference = "bare-secret-token"
	assertAssistantCode(t, assistant.ValidateCredentialRef(rawLooking), assistant.CodeInvalidCredential)

	retrieval := validRetrievalPolicy()
	if err := assistant.ValidateRetrievalPolicy(retrieval); err != nil {
		t.Fatalf("ValidateRetrievalPolicy = %v", err)
	}
	retrieval.ContextLimit = retrieval.CandidateLimit + 1
	assertAssistantCode(t, assistant.ValidateRetrievalPolicy(retrieval), assistant.CodeInvalidRetrieval)
}

func validAssistantDefinition() assistant.AssistantDefinition {
	return assistant.AssistantDefinition{
		ID:          "multimodal_researcher",
		Name:        "Multimodal Researcher",
		Description: "Researches grounded answers across supported media.",
		Instructions: []assistant.InstructionBlock{{
			Role: assistant.RoleDeveloper,
			Text: "Ground answers in retrieved evidence when retrieval is enabled.",
		}},
		Modules: []assistant.ModuleID{
			assistant.ModuleText,
			assistant.ModuleVision,
			assistant.ModuleTools,
			assistant.ModuleRetrieval,
		},
		Skills: []assistant.Skill{assistant.BuiltinTokenEconomySkill()},
		Tools:  []string{"search_docs"},
		ToolPolicies: []assistant.ToolPolicy{{
			Name:         "search_docs",
			Effect:       assistant.ToolEffectReadOnly,
			Confirmation: assistant.ToolConfirmNever,
			ParallelSafe: true,
			Idempotent:   true,
		}},
		ModelPolicy: assistant.ModelPolicy{
			Mode:             assistant.ModelDynamic,
			AllowedProviders: []assistant.ProviderID{"gateway"},
			AllowedAccess:    []assistant.AccessTier{assistant.AccessFree, assistant.AccessPaid},
			PreferredAccess:  []assistant.AccessTier{assistant.AccessFree, assistant.AccessPaid},
		},
		Credentials: assistant.CredentialPolicy{AllowedModes: []assistant.CredentialMode{assistant.CredentialManaged, assistant.CredentialBYOK}},
		Retrieval:   retrievalPolicyPtr(validRetrievalPolicy()),
		Budget: assistant.ExecutionBudget{
			MaxTurns:           8,
			MaxToolCalls:       12,
			MaxRetrievalRounds: 3,
			MaxOutputTokens:    4096,
		},
	}
}

func validRetrievalPolicy() assistant.RetrievalPolicy {
	return assistant.RetrievalPolicy{
		Mode:             assistant.RetrievalOnDemand,
		Search:           assistant.SearchHybrid,
		QueryTransform:   assistant.QueryRewrite,
		CandidateLimit:   20,
		ContextLimit:     6,
		Rerank:           true,
		Grounding:        assistant.GroundingRequired,
		RequireCitations: true,
	}
}

func retrievalPolicyPtr(value assistant.RetrievalPolicy) *assistant.RetrievalPolicy {
	return &value
}

func assertAssistantCode(t *testing.T, err error, code assistant.ErrorCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error code %q", code)
	}
	var validation *assistant.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("error %T is not ValidationError: %v", err, err)
	}
	if validation.Code != code {
		t.Fatalf("error code = %q, want %q", validation.Code, code)
	}
}
