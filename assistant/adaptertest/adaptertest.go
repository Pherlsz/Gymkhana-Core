// Package adaptertest provides reusable integration checks for implementations
// of assistant.ProviderAdapter. It exercises the public registry boundary rather
// than provider SDK details so external adapters can share one conformance harness.
package adaptertest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Pherlsz/Gymkhana-Core/assistant"
)

const defaultTimeout = 30 * time.Second

// Config describes one adapter/model pair to exercise. Model must be a
// generation model returned by Adapter.ListModels and must advertise text
// capability because Run always executes one minimal text generation smoke test.
type Config struct {
	Adapter    assistant.ProviderAdapter
	Credential *assistant.CredentialRef
	Model      assistant.ModelRef
	Timeout    time.Duration

	// Probes exercise additional capability-specific request shapes such as tools,
	// structured output, or multimodal input. Run verifies RequiredCapabilities
	// are advertised by the selected model before invoking a probe.
	Probes []Probe

	// FailureProbes validate optional FailureClassifier mappings using
	// adapter-owned representative errors without making network calls.
	FailureProbes []FailureProbe

	// ProbeQuota calls the optional QuotaProvider boundary. Set it only when the
	// adapter is expected to expose normalized quota observations.
	ProbeQuota bool
}

// Probe is one additional generation request executed against Config.Model.
// Model and Credential are supplied by the harness; callers provide only the
// portable message/tool/schema shape and optional response assertion.
type Probe struct {
	Name                 string
	RequiredCapabilities []assistant.Capability
	Messages             []assistant.Message
	Tools                []assistant.ToolDefinition
	ResponseSchema       json.RawMessage
	Check                func(assistant.GenerationResponse) error
}

// FailureProbe is one expected mapping from an adapter/vendor error to the
// provider-neutral failure taxonomy used by Assistant fallback policy.
type FailureProbe struct {
	Name string
	Err  error
	Want assistant.FailureClass
}

// Run executes reusable adapter conformance checks. It is intended for adapter
// integration tests; Core itself does not embed provider SDKs or credentials.
func Run(t *testing.T, config Config) {
	t.Helper()

	if err := validateConfig(config); err != nil {
		t.Fatalf("adaptertest: invalid config: %v", err)
	}

	registry := assistant.NewAdapterRegistry()
	if err := registry.Register(config.Adapter); err != nil {
		t.Fatalf("adaptertest: register adapter: %v", err)
	}

	descriptor := config.Adapter.Descriptor()
	if !t.Run("descriptor", func(t *testing.T) {
		if err := assistant.ValidateProviderDescriptor(descriptor); err != nil {
			t.Fatalf("ValidateProviderDescriptor() = %v", err)
		}
		if descriptor.ID != config.Model.Provider {
			t.Fatalf("descriptor provider %q does not match model provider %q", descriptor.ID, config.Model.Provider)
		}
	}) {
		return
	}

	var selected assistant.ModelDescriptor
	if !t.Run("catalog", func(t *testing.T) {
		models, err := registry.ListModels(operationContext(t, config.Timeout), descriptor.ID, config.Credential)
		if err != nil {
			t.Fatalf("ListModels() = %v", err)
		}
		var ok bool
		selected, ok = findModel(models, config.Model)
		if !ok {
			t.Fatalf("configured model %+v not returned by ListModels", config.Model)
		}
		if !hasRole(selected, assistant.ModelRoleGeneration) {
			t.Fatalf("configured model %+v does not advertise generation role", config.Model)
		}
		if !hasCapability(selected, assistant.CapabilityText) {
			t.Fatalf("configured model %+v does not advertise text capability", config.Model)
		}
	}) {
		return
	}

	if !t.Run("generate/text", func(t *testing.T) {
		_, err := registry.Generate(operationContext(t, config.Timeout), assistant.GenerationRequest{
			Model: config.Model,
			Messages: []assistant.Message{{
				Role:    assistant.RoleUser,
				Content: []assistant.ContentPart{{Type: assistant.PartText, Text: "Reply with OK."}},
			}},
			Credential: config.Credential,
		})
		if err != nil {
			t.Fatalf("Generate(text) = %v", err)
		}
	}) {
		return
	}

	for _, probe := range config.Probes {
		probe := probe
		t.Run("probe/"+probe.Name, func(t *testing.T) {
			for _, capability := range probe.RequiredCapabilities {
				if !hasCapability(selected, capability) {
					t.Fatalf("configured model %+v does not advertise required capability %q", config.Model, capability)
				}
			}
			response, err := registry.Generate(operationContext(t, config.Timeout), assistant.GenerationRequest{
				Model:          config.Model,
				Messages:       probe.Messages,
				Tools:          probe.Tools,
				ResponseSchema: probe.ResponseSchema,
				Credential:     config.Credential,
			})
			if err != nil {
				t.Fatalf("Generate(%s) = %v", probe.Name, err)
			}
			if probe.Check != nil {
				if err := probe.Check(response); err != nil {
					t.Fatalf("probe check = %v", err)
				}
			}
		})
	}

	if config.ProbeQuota {
		t.Run("quota", func(t *testing.T) {
			state, err := registry.Quota(operationContext(t, config.Timeout), descriptor.ID, config.Credential, &config.Model)
			if err != nil {
				t.Fatalf("Quota() = %v", err)
			}
			if state.Provider != descriptor.ID {
				t.Fatalf("quota provider %q does not match descriptor provider %q", state.Provider, descriptor.ID)
			}
		})
	}

	for _, probe := range config.FailureProbes {
		probe := probe
		t.Run("failure/"+probe.Name, func(t *testing.T) {
			got := registry.ClassifyFailure(descriptor.ID, probe.Err)
			if got != probe.Want {
				t.Fatalf("ClassifyFailure() = %q, want %q", got, probe.Want)
			}
		})
	}
}

func validateConfig(config Config) error {
	if config.Adapter == nil {
		return errors.New("adapter is required")
	}
	if config.Model.Provider == "" || config.Model.Model == "" {
		return errors.New("model is required")
	}
	if config.Timeout < 0 {
		return errors.New("timeout must not be negative")
	}
	if config.Credential != nil {
		if err := assistant.ValidateCredentialRef(*config.Credential); err != nil {
			return fmt.Errorf("credential: %w", err)
		}
		if config.Credential.Provider != config.Model.Provider {
			return errors.New("credential provider must match model provider")
		}
	}
	seenProbeNames := make(map[string]struct{}, len(config.Probes))
	for _, probe := range config.Probes {
		if probe.Name == "" {
			return errors.New("probe.Name is required")
		}
		if _, duplicate := seenProbeNames[probe.Name]; duplicate {
			return fmt.Errorf("duplicate Probe.Name %q", probe.Name)
		}
		seenProbeNames[probe.Name] = struct{}{}
		for _, capability := range probe.RequiredCapabilities {
			if !capability.Valid() {
				return fmt.Errorf("probe %q has invalid capability %q", probe.Name, capability)
			}
		}
	}
	if config.ProbeQuota {
		if _, ok := config.Adapter.(assistant.QuotaProvider); !ok {
			return errors.New("probeQuota requires assistant.QuotaProvider")
		}
	}
	if len(config.FailureProbes) > 0 {
		if _, ok := config.Adapter.(assistant.FailureClassifier); !ok {
			return errors.New("failureProbes require assistant.FailureClassifier")
		}
	}
	seenFailureNames := make(map[string]struct{}, len(config.FailureProbes))
	for _, probe := range config.FailureProbes {
		if probe.Name == "" {
			return errors.New("failureProbe.Name is required")
		}
		if _, duplicate := seenFailureNames[probe.Name]; duplicate {
			return fmt.Errorf("duplicate FailureProbe.Name %q", probe.Name)
		}
		seenFailureNames[probe.Name] = struct{}{}
		if probe.Err == nil {
			return fmt.Errorf("failureProbe %q requires Err", probe.Name)
		}
		if !probe.Want.Valid() {
			return fmt.Errorf("failureProbe %q has invalid Want %q", probe.Name, probe.Want)
		}
	}
	return nil
}

func operationContext(t *testing.T, configured time.Duration) context.Context {
	t.Helper()
	timeout := configured
	if timeout == 0 {
		timeout = defaultTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)
	return ctx
}

func findModel(models []assistant.ModelDescriptor, ref assistant.ModelRef) (assistant.ModelDescriptor, bool) {
	for _, model := range models {
		if model.Ref == ref {
			return model, true
		}
	}
	return assistant.ModelDescriptor{}, false
}

func hasRole(model assistant.ModelDescriptor, role assistant.ModelRole) bool {
	for _, candidate := range model.Roles {
		if candidate == role {
			return true
		}
	}
	return false
}

func hasCapability(model assistant.ModelDescriptor, capability assistant.Capability) bool {
	for _, candidate := range model.Capabilities {
		if candidate == capability {
			return true
		}
	}
	return false
}
