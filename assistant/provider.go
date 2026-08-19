package assistant

import (
	"context"
	"encoding/json"
	"sync"
	"unicode/utf8"
)

// Well-known provider IDs are conveniences, not a closed provider enum.
// Additional adapters may use any valid ProviderID without changing Core.
const (
	ProviderOpenAI     ProviderID = "openai"
	ProviderAnthropic  ProviderID = "anthropic"
	ProviderGoogle     ProviderID = "google"
	ProviderOpenRouter ProviderID = "openrouter"
	ProviderGroq       ProviderID = "groq"
	ProviderOllama     ProviderID = "ollama"
	ProviderMistral    ProviderID = "mistral"
	ProviderXAI        ProviderID = "xai"
)

// ProviderDescriptor describes one adapter without embedding volatile model or
// pricing data into Core.
type ProviderDescriptor struct {
	ID              ProviderID       `json:"id"`
	DisplayName     string           `json:"display_name,omitempty"`
	CredentialModes []CredentialMode `json:"credential_modes"`
}

// GenerationRequest is the provider-neutral generation boundary used by a Go
// adapter. The adapter instance resolves Credential references through an
// application-owned secure credential resolver configured outside Core.
type GenerationRequest struct {
	Model          ModelRef         `json:"model"`
	Messages       []Message        `json:"messages"`
	Tools          []ToolDefinition `json:"tools,omitempty"`
	ResponseSchema json.RawMessage  `json:"response_schema,omitempty"`
	Credential     *CredentialRef   `json:"credential,omitempty"`
}

// GenerationResponse is the normalized non-streaming provider result.
type GenerationResponse struct {
	Model        ModelRef     `json:"model"`
	Message      Message      `json:"message"`
	FinishReason FinishReason `json:"finish_reason"`
	Usage        Usage        `json:"usage,omitempty"`
}

// ProviderAdapter is the current Go substitution boundary for concrete AI
// providers/gateways/local runtimes. Implementations may use SDKs internally;
// those SDK types must not cross this interface.
type ProviderAdapter interface {
	Descriptor() ProviderDescriptor
	ListModels(ctx context.Context, credential *CredentialRef) ([]ModelDescriptor, error)
	Generate(ctx context.Context, request GenerationRequest) (GenerationResponse, error)
}

// AdapterRegistry stores provider adapters for one runtime. It is safe for
// concurrent lookup/registration and deliberately contains no global singleton.
type AdapterRegistry struct {
	mu       sync.RWMutex
	adapters map[ProviderID]ProviderAdapter
}

// NewAdapterRegistry creates an empty runtime-local provider registry.
func NewAdapterRegistry() *AdapterRegistry {
	return &AdapterRegistry{adapters: make(map[ProviderID]ProviderAdapter)}
}

// Register adds one provider adapter and rejects duplicate provider IDs.
func (registry *AdapterRegistry) Register(adapter ProviderAdapter) error {
	if adapter == nil {
		return validationError(CodeInvalidProvider, "provider")
	}
	descriptor := adapter.Descriptor()
	if err := ValidateProviderDescriptor(descriptor); err != nil {
		return err
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	if _, exists := registry.adapters[descriptor.ID]; exists {
		return validationError(CodeDuplicateProvider, "provider")
	}
	registry.adapters[descriptor.ID] = adapter
	return nil
}

// Get resolves an adapter by provider ID.
func (registry *AdapterRegistry) Get(provider ProviderID) (ProviderAdapter, bool) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	adapter, ok := registry.adapters[provider]
	return adapter, ok
}

// Providers returns deterministic registered provider IDs.
func (registry *AdapterRegistry) Providers() []ProviderID {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	result := make([]ProviderID, 0, len(registry.adapters))
	for provider := range registry.adapters {
		result = append(result, provider)
	}
	sortProviderIDs(result)
	return result
}

// ValidateProviderDescriptor validates stable adapter metadata.
func ValidateProviderDescriptor(descriptor ProviderDescriptor) error {
	if !validPortableID(string(descriptor.ID), 128) || !utf8.ValidString(descriptor.DisplayName) {
		return validationError(CodeInvalidProvider, "provider")
	}
	if len(descriptor.CredentialModes) == 0 {
		return validationError(CodeEmpty, "provider.credential_modes")
	}
	seen := make(map[CredentialMode]struct{}, len(descriptor.CredentialModes))
	for _, mode := range descriptor.CredentialModes {
		if !mode.Valid() {
			return validationError(CodeInvalidProvider, "provider.credential_modes")
		}
		if _, exists := seen[mode]; exists {
			return validationError(CodeInvalidProvider, "provider.credential_modes")
		}
		seen[mode] = struct{}{}
	}
	return nil
}

// ValidateGenerationRequest validates normalized adapter input without invoking
// a provider or resolving credential secrets.
func ValidateGenerationRequest(request GenerationRequest) error {
	if !validModelRef(request.Model) || len(request.Messages) == 0 {
		return validationError(CodeInvalidProvider, "generation_request")
	}
	for _, message := range request.Messages {
		if err := ValidateMessage(message); err != nil {
			return err
		}
	}
	seenTools := make(map[string]struct{}, len(request.Tools))
	for _, tool := range request.Tools {
		if err := ValidateToolDefinition(tool); err != nil {
			return err
		}
		if _, exists := seenTools[tool.Name]; exists {
			return validationError(CodeInvalidProvider, "generation_request.tools")
		}
		seenTools[tool.Name] = struct{}{}
	}
	if len(request.ResponseSchema) > 0 && !validJSONObject(request.ResponseSchema) {
		return validationError(CodeInvalidJSON, "response_schema")
	}
	if request.Credential != nil {
		if err := ValidateCredentialRef(*request.Credential); err != nil {
			return err
		}
		if request.Credential.Provider != request.Model.Provider {
			return validationError(CodeInvalidCredential, "credential.provider")
		}
	}
	return nil
}

// ValidateGenerationResponse validates normalized adapter output. Provider
// generation output must be represented as an assistant-role message; tool
// results remain separate tool-role input messages.
func ValidateGenerationResponse(response GenerationResponse) error {
	if !validModelRef(response.Model) {
		return validationError(CodeInvalidModel, "generation_response.model")
	}
	if response.Message.Role != RoleAssistant {
		return validationError(CodeInvalidProvider, "generation_response.message.role")
	}
	if err := ValidateMessage(response.Message); err != nil {
		return err
	}
	if err := ValidateFinishReason(response.FinishReason); err != nil {
		return err
	}
	if err := ValidateUsage(response.Usage); err != nil {
		return err
	}
	return nil
}

func sortProviderIDs(values []ProviderID) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}
