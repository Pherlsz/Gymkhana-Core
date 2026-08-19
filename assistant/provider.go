package assistant

import (
	"context"
	"encoding/json"
	"reflect"
	"sync"
	"unicode/utf8"
)

// Well-known provider IDs are conveniences, not a closed provider enum.
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
// adapter. Credential contains only an opaque application-owned reference.
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
// providers/gateways/local runtimes.
type ProviderAdapter interface {
	Descriptor() ProviderDescriptor
	ListModels(ctx context.Context, credential *CredentialRef) ([]ModelDescriptor, error)
	Generate(ctx context.Context, request GenerationRequest) (GenerationResponse, error)
}

// QuotaProvider is an optional adapter capability. Provider quotas are queried
// for a credential/quota scope and optional model; Assistant attribution is a
// separate local UsageLedger concern because provider quotas are often applied
// to projects/accounts rather than individual API keys or Assistant profiles.
type QuotaProvider interface {
	Quota(ctx context.Context, credential CredentialRef, model *ModelRef) (QuotaState, error)
}

// AdapterRegistry stores provider adapters for one runtime.
type AdapterRegistry struct {
	mu       sync.RWMutex
	adapters map[ProviderID]ProviderAdapter
}

func NewAdapterRegistry() *AdapterRegistry {
	return &AdapterRegistry{adapters: make(map[ProviderID]ProviderAdapter)}
}

func (registry *AdapterRegistry) Register(adapter ProviderAdapter) error {
	if isNilProviderAdapter(adapter) {
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

func (registry *AdapterRegistry) Get(provider ProviderID) (ProviderAdapter, bool) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	adapter, ok := registry.adapters[provider]
	return adapter, ok
}

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

// ListModels calls one adapter through the validated Core boundary and rejects
// catalog entries that claim a different provider or duplicate model identity.
func (registry *AdapterRegistry) ListModels(ctx context.Context, provider ProviderID, credential *CredentialRef) ([]ModelDescriptor, error) {
	adapter, ok := registry.Get(provider)
	if !ok {
		return nil, validationError(CodeInvalidProvider, "provider")
	}
	descriptor := adapter.Descriptor()
	if err := ValidateProviderDescriptor(descriptor); err != nil || descriptor.ID != provider {
		return nil, validationError(CodeInvalidProvider, "provider")
	}
	if credential != nil {
		if err := ValidateCredentialRef(*credential); err != nil {
			return nil, err
		}
		if credential.Provider != provider || !providerAllowsCredential(descriptor, credential.Mode) {
			return nil, validationError(CodeInvalidCredential, "credential")
		}
	}
	models, err := adapter.ListModels(ctx, credential)
	if err != nil {
		return nil, err
	}
	seen := make(map[ModelRef]struct{}, len(models))
	for _, model := range models {
		if err := ValidateModelDescriptor(model); err != nil {
			return nil, err
		}
		if model.Ref.Provider != provider {
			return nil, validationError(CodeInvalidProvider, "catalog.provider")
		}
		if _, duplicate := seen[model.Ref]; duplicate {
			return nil, validationError(CodeInvalidModel, "catalog")
		}
		seen[model.Ref] = struct{}{}
	}
	return models, nil
}

// Generate invokes the selected provider through a validated boundary. Direct
// adapter calls remain possible for low-level integrations, but consumers that
// want Core invariants should prefer this method.
func (registry *AdapterRegistry) Generate(ctx context.Context, request GenerationRequest) (GenerationResponse, error) {
	if err := ValidateGenerationRequest(request); err != nil {
		return GenerationResponse{}, err
	}
	adapter, ok := registry.Get(request.Model.Provider)
	if !ok {
		return GenerationResponse{}, validationError(CodeInvalidProvider, "provider")
	}
	descriptor := adapter.Descriptor()
	if err := ValidateProviderDescriptor(descriptor); err != nil || descriptor.ID != request.Model.Provider {
		return GenerationResponse{}, validationError(CodeInvalidProvider, "provider")
	}
	if request.Credential != nil {
		if !providerAllowsCredential(descriptor, request.Credential.Mode) {
			return GenerationResponse{}, validationError(CodeInvalidCredential, "credential.mode")
		}
	} else if !providerAllowsCredential(descriptor, CredentialManaged) && !providerAllowsCredential(descriptor, CredentialNone) {
		return GenerationResponse{}, validationError(CodeInvalidCredential, "credential")
	}
	response, err := adapter.Generate(ctx, request)
	if err != nil {
		return GenerationResponse{}, err
	}
	if err := ValidateGenerationExchange(request, response); err != nil {
		return GenerationResponse{}, err
	}
	return response, nil
}

func ValidateProviderDescriptor(descriptor ProviderDescriptor) error {
	if !validPortableID(string(descriptor.ID), 128) || !utf8.ValidString(descriptor.DisplayName) || utf8.RuneCountInString(descriptor.DisplayName) > 256 {
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

func ValidateGenerationRequest(request GenerationRequest) error {
	if !validModelRef(request.Model) {
		return validationError(CodeInvalidProvider, "generation_request.model")
	}
	if err := ValidateConversation(request.Messages); err != nil {
		return err
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
	if len(request.ResponseSchema) > 0 {
		if err := ValidatePortableJSONSchema(request.ResponseSchema); err != nil {
			return validationError(CodeInvalidSchema, "response_schema")
		}
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
	return ValidateUsage(response.Usage)
}

// ValidateGenerationExchange validates invariants that require both sides of the
// provider boundary, such as provider identity and returned tool allowlists.
func ValidateGenerationExchange(request GenerationRequest, response GenerationResponse) error {
	if err := ValidateGenerationRequest(request); err != nil {
		return err
	}
	if err := ValidateGenerationResponse(response); err != nil {
		return err
	}
	if response.Model.Provider != request.Model.Provider {
		return validationError(CodeInvalidProvider, "generation_response.model.provider")
	}
	allowedTools := make(map[string]struct{}, len(request.Tools))
	for _, tool := range request.Tools {
		allowedTools[tool.Name] = struct{}{}
	}
	toolCalls := 0
	seenCallIDs := make(map[string]struct{})
	for _, part := range response.Message.Content {
		if part.Type != PartToolCall {
			continue
		}
		toolCalls++
		if _, allowed := allowedTools[part.ToolCall.Name]; !allowed {
			return validationError(CodeInvalidToolCall, "generation_response.tool_call.name")
		}
		if _, duplicate := seenCallIDs[part.ToolCall.ID]; duplicate {
			return validationError(CodeInvalidToolCall, "generation_response.tool_call.id")
		}
		seenCallIDs[part.ToolCall.ID] = struct{}{}
	}
	if (response.FinishReason == FinishToolCalls) != (toolCalls > 0) {
		return validationError(CodeInvalidFinishReason, "generation_response.finish_reason")
	}
	return nil
}

func providerAllowsCredential(descriptor ProviderDescriptor, mode CredentialMode) bool {
	for _, allowed := range descriptor.CredentialModes {
		if allowed == mode {
			return true
		}
	}
	return false
}

func isNilProviderAdapter(adapter ProviderAdapter) bool {
	if adapter == nil {
		return true
	}
	value := reflect.ValueOf(adapter)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func sortProviderIDs(values []ProviderID) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}
