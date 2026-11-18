package assistant_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
)

func BenchmarkResolveModelCandidates10000(b *testing.B) {
	catalog := make([]assistant.ModelDescriptor, 10000)
	for i := range catalog {
		access := assistant.AccessPaid
		if i%3 == 0 {
			access = assistant.AccessFree
		}
		catalog[i] = assistant.ModelDescriptor{
			Ref:          assistant.ModelRef{Provider: "gateway", Model: assistant.ModelID(fmt.Sprintf("vendor/model-%05d", i))},
			Access:       access,
			Roles:        []assistant.ModelRole{assistant.ModelRoleGeneration},
			Capabilities: []assistant.Capability{assistant.CapabilityText, assistant.CapabilityToolCalling},
		}
	}
	policy := assistant.ModelPolicy{
		Mode:                 assistant.ModelDynamic,
		AllowedAccess:        []assistant.AccessTier{assistant.AccessFree, assistant.AccessPaid},
		PreferredAccess:      []assistant.AccessTier{assistant.AccessFree, assistant.AccessPaid},
		RequiredCapabilities: []assistant.Capability{assistant.CapabilityText, assistant.CapabilityToolCalling},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		resolved, err := assistant.ResolveModelCandidates(policy, catalog)
		if err != nil {
			b.Fatal(err)
		}
		if len(resolved) != len(catalog) {
			b.Fatalf("resolved %d models, want %d", len(resolved), len(catalog))
		}
	}
}

func BenchmarkResolveAssistantModelsForRun(b *testing.B) {
	def := validAssistantDefinition()
	def.OptionalModules = []assistant.ModuleID{assistant.ModuleAudioInput}
	overrides := assistant.RunOverrides{
		EnableModules: []assistant.ModuleID{assistant.ModuleAudioInput},
	}
	catalog := []assistant.ModelDescriptor{
		{
			Ref:          assistant.ModelRef{Provider: "gateway", Model: "vision-audio/free"},
			Access:       assistant.AccessFree,
			Roles:        []assistant.ModelRole{assistant.ModelRoleGeneration},
			Capabilities: []assistant.Capability{assistant.CapabilityText, assistant.CapabilityImageInput, assistant.CapabilityAudioInput, assistant.CapabilityToolCalling},
		},
		{
			Ref:          assistant.ModelRef{Provider: "gateway", Model: "vision-audio/paid"},
			Access:       assistant.AccessPaid,
			Roles:        []assistant.ModelRole{assistant.ModelRoleGeneration},
			Capabilities: []assistant.Capability{assistant.CapabilityText, assistant.CapabilityImageInput, assistant.CapabilityAudioInput, assistant.CapabilityToolCalling},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		resolved, err := assistant.ResolveAssistantModelsForRun(def, overrides, catalog)
		if err != nil || len(resolved) != 2 {
			b.Fatalf("ResolveAssistantModelsForRun = (%v, %v)", resolved, err)
		}
	}
}

func BenchmarkValidateGenerationToolArguments(b *testing.B) {
	request := assistant.GenerationRequest{
		Model:    assistant.ModelRef{Provider: "openai", Model: "model/latest"},
		Messages: []assistant.Message{{Role: assistant.RoleUser, Content: []assistant.ContentPart{{Type: assistant.PartText, Text: "Weather?"}}}},
		Tools: []assistant.ToolDefinition{{
			Name:        "lookup_weather",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}},"required":["city"],"additionalProperties":false}`),
		}},
	}
	response := assistant.GenerationResponse{
		Model: assistant.ModelRef{Provider: "openai", Model: "model/latest"},
		Message: assistant.Message{Role: assistant.RoleAssistant, Content: []assistant.ContentPart{{
			Type:     assistant.PartToolCall,
			ToolCall: &assistant.ToolCall{ID: "call_01", Name: "lookup_weather", Arguments: json.RawMessage(`{"city":"Porto Alegre"}`)},
		}}},
		FinishReason: assistant.FinishToolCalls,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if err := assistant.ValidateGenerationToolArguments(request, response); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateRoutingFallbackPolicy(b *testing.B) {
	policy := assistant.DefaultRoutingFallbackPolicy()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if err := assistant.ValidateRoutingFallbackPolicy(policy); err != nil {
			b.Fatal(err)
		}
	}
}
