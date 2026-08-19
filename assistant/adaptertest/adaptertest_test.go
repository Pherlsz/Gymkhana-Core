package adaptertest

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
)

var errRateLimited = errors.New("rate limited")

type fakeAdapter struct {
	generateCalls int
	quotaCalls    int
}

func (adapter *fakeAdapter) Descriptor() assistant.ProviderDescriptor {
	return assistant.ProviderDescriptor{
		ID:              "test-provider",
		DisplayName:     "Test Provider",
		CredentialModes: []assistant.CredentialMode{assistant.CredentialNone},
	}
}

func (adapter *fakeAdapter) ListModels(context.Context, *assistant.CredentialRef) ([]assistant.ModelDescriptor, error) {
	return []assistant.ModelDescriptor{{
		Ref:          assistant.ModelRef{Provider: "test-provider", Model: "test-model"},
		DisplayName:  "Test Model",
		Access:       assistant.AccessLocal,
		Roles:        []assistant.ModelRole{assistant.ModelRoleGeneration},
		Capabilities: []assistant.Capability{assistant.CapabilityText, assistant.CapabilityStructuredOutput},
	}}, nil
}

func (adapter *fakeAdapter) Generate(_ context.Context, request assistant.GenerationRequest) (assistant.GenerationResponse, error) {
	adapter.generateCalls++
	return assistant.GenerationResponse{
		Model: request.Model,
		Message: assistant.Message{
			Role:    assistant.RoleAssistant,
			Content: []assistant.ContentPart{{Type: assistant.PartText, Text: "OK"}},
		},
		FinishReason: assistant.FinishStop,
		Usage:        assistant.Usage{InputTokens: 2, OutputTokens: 1},
	}, nil
}

func (adapter *fakeAdapter) ClassifyFailure(err error) assistant.FailureClass {
	if errors.Is(err, errRateLimited) {
		return assistant.FailureRateLimit
	}
	return assistant.FailureUnknown
}

func (adapter *fakeAdapter) Quota(_ context.Context, _ *assistant.CredentialRef, model *assistant.ModelRef) (assistant.QuotaState, error) {
	adapter.quotaCalls++
	return assistant.QuotaState{
		Provider: "test-provider",
		Model:    model,
		Known:    true,
	}, nil
}

func TestRunExercisesAdapterBoundaries(t *testing.T) {
	adapter := &fakeAdapter{}
	Run(t, Config{
		Adapter: adapter,
		Model:   assistant.ModelRef{Provider: "test-provider", Model: "test-model"},
		Probes: []Probe{{
			Name:                 "structured-output",
			RequiredCapabilities: []assistant.Capability{assistant.CapabilityStructuredOutput},
			Messages: []assistant.Message{{
				Role:    assistant.RoleUser,
				Content: []assistant.ContentPart{{Type: assistant.PartText, Text: "Return OK as JSON."}},
			}},
			ResponseSchema: json.RawMessage(`{"type":"string"}`),
		}},
		FailureProbes: []FailureProbe{{
			Name: "rate-limit",
			Err:  errRateLimited,
			Want: assistant.FailureRateLimit,
		}},
		ProbeQuota: true,
	})

	if adapter.generateCalls != 2 {
		t.Fatalf("Generate calls = %d, want 2", adapter.generateCalls)
	}
	if adapter.quotaCalls != 1 {
		t.Fatalf("Quota calls = %d, want 1", adapter.quotaCalls)
	}
}

func TestValidateConfigRejectsUnsupportedOptionalProbes(t *testing.T) {
	adapter := providerOnlyAdapter{}
	config := Config{
		Adapter:    &adapter,
		Model:      assistant.ModelRef{Provider: "test-provider", Model: "test-model"},
		ProbeQuota: true,
	}
	if err := validateConfig(config); err == nil {
		t.Fatal("validateConfig() accepted ProbeQuota without QuotaProvider")
	}

	config.ProbeQuota = false
	config.FailureProbes = []FailureProbe{{Name: "rate-limit", Err: errRateLimited, Want: assistant.FailureRateLimit}}
	if err := validateConfig(config); err == nil {
		t.Fatal("validateConfig() accepted FailureProbes without FailureClassifier")
	}
}

type providerOnlyAdapter struct {
	inner fakeAdapter
}

func (adapter *providerOnlyAdapter) Descriptor() assistant.ProviderDescriptor {
	return adapter.inner.Descriptor()
}

func (adapter *providerOnlyAdapter) ListModels(ctx context.Context, credential *assistant.CredentialRef) ([]assistant.ModelDescriptor, error) {
	return adapter.inner.ListModels(ctx, credential)
}

func (adapter *providerOnlyAdapter) Generate(ctx context.Context, request assistant.GenerationRequest) (assistant.GenerationResponse, error) {
	return adapter.inner.Generate(ctx, request)
}
