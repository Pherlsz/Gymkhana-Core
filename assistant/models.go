package assistant

import (
	"sort"
	"unicode/utf8"
)

// ProviderID is a stable provider namespace supplied by a provider adapter.
// Core does not restrict providers to a closed enum.
type ProviderID string

// ModelID is a provider-local model identifier.
type ModelID string

// ModelRef identifies one concrete provider/model pair. It is comparable and is
// therefore used directly as a map key; concatenated string keys are avoided so
// provider/model names containing separators cannot collide.
type ModelRef struct {
	Provider ProviderID `json:"provider"`
	Model    ModelID    `json:"model"`
}

// AccessTier describes current commercial/runtime accessibility. It is adapter
// metadata rather than a permanent property of a model and may change over time.
type AccessTier string

const (
	AccessFree    AccessTier = "free"
	AccessPaid    AccessTier = "paid"
	AccessLocal   AccessTier = "local"
	AccessUnknown AccessTier = "unknown"
)

func (tier AccessTier) Valid() bool {
	switch tier {
	case AccessFree, AccessPaid, AccessLocal, AccessUnknown:
		return true
	default:
		return false
	}
}

// ModelRole describes how a model may participate in an Assistant/RAG stack.
type ModelRole string

const (
	ModelRoleGeneration ModelRole = "generation"
	ModelRoleEmbedding  ModelRole = "embedding"
	ModelRoleReranking  ModelRole = "reranking"
)

func (role ModelRole) Valid() bool {
	switch role {
	case ModelRoleGeneration, ModelRoleEmbedding, ModelRoleReranking:
		return true
	default:
		return false
	}
}

// ModelDescriptor is live metadata returned by provider adapters/catalogs.
// Pricing/model names are not hardcoded into Core because they change independently.
type ModelDescriptor struct {
	Ref             ModelRef     `json:"ref"`
	DisplayName     string       `json:"display_name,omitempty"`
	Access          AccessTier   `json:"access"`
	Roles           []ModelRole  `json:"roles"`
	Capabilities    []Capability `json:"capabilities,omitempty"`
	ContextWindow   int64        `json:"context_window,omitempty"`
	MaxOutputTokens int64        `json:"max_output_tokens,omitempty"`
}

// ModelSelectionMode controls whether selection is exact, explicit fallback, or
// dynamically resolved against adapter-supplied catalog metadata.
type ModelSelectionMode string

const (
	ModelManual          ModelSelectionMode = "manual"
	ModelOrderedFallback ModelSelectionMode = "ordered_fallback"
	ModelDynamic         ModelSelectionMode = "dynamic"
)

func (mode ModelSelectionMode) Valid() bool {
	switch mode {
	case ModelManual, ModelOrderedFallback, ModelDynamic:
		return true
	default:
		return false
	}
}

// ModelPolicy is provider-neutral. Dynamic policies filter current catalog data;
// PreferredAccess only affects deterministic ordering, never availability truth.
type ModelPolicy struct {
	Mode                 ModelSelectionMode `json:"mode"`
	Manual               *ModelRef          `json:"manual,omitempty"`
	Candidates           []ModelRef         `json:"candidates,omitempty"`
	AllowedProviders     []ProviderID       `json:"allowed_providers,omitempty"`
	AllowedAccess        []AccessTier       `json:"allowed_access,omitempty"`
	PreferredAccess      []AccessTier       `json:"preferred_access,omitempty"`
	RequiredCapabilities []Capability       `json:"required_capabilities,omitempty"`
}

// ValidateModelPolicy validates portable routing configuration.
func ValidateModelPolicy(policy ModelPolicy) error {
	if !policy.Mode.Valid() {
		return validationError(CodeInvalidModelPolicy, "model_policy.mode")
	}
	if policy.Manual != nil && !validModelRef(*policy.Manual) {
		return validationError(CodeInvalidModel, "model_policy.manual")
	}
	seenCandidates := make(map[ModelRef]struct{}, len(policy.Candidates))
	for _, ref := range policy.Candidates {
		if !validModelRef(ref) {
			return validationError(CodeInvalidModel, "model_policy.candidates")
		}
		if _, exists := seenCandidates[ref]; exists {
			return validationError(CodeInvalidModelPolicy, "model_policy.candidates")
		}
		seenCandidates[ref] = struct{}{}
	}
	if policy.Mode == ModelManual && policy.Manual == nil {
		return validationError(CodeInvalidModelPolicy, "model_policy.manual")
	}
	if policy.Mode != ModelManual && policy.Manual != nil {
		return validationError(CodeInvalidModelPolicy, "model_policy.manual")
	}
	if policy.Mode == ModelOrderedFallback && len(policy.Candidates) == 0 {
		return validationError(CodeInvalidModelPolicy, "model_policy.candidates")
	}
	if policy.Mode != ModelOrderedFallback && len(policy.Candidates) > 0 {
		return validationError(CodeInvalidModelPolicy, "model_policy.candidates")
	}
	if err := validateProviderIDs(policy.AllowedProviders); err != nil {
		return err
	}
	if err := validateAccessTiers(policy.AllowedAccess); err != nil {
		return err
	}
	if err := validateAccessTiers(policy.PreferredAccess); err != nil {
		return err
	}
	if len(policy.AllowedAccess) > 0 {
		allowed := make(map[AccessTier]struct{}, len(policy.AllowedAccess))
		for _, tier := range policy.AllowedAccess {
			allowed[tier] = struct{}{}
		}
		for _, tier := range policy.PreferredAccess {
			if _, ok := allowed[tier]; !ok {
				return validationError(CodeInvalidModelPolicy, "model_policy.preferred_access")
			}
		}
	}
	if err := ValidateCapabilities(policy.RequiredCapabilities); err != nil {
		return err
	}
	return nil
}

// RequiredCapabilitiesForModules derives generation-model requirements from an
// Assistant's enabled modules.
func RequiredCapabilitiesForModules(modules []ModuleID) []Capability {
	seen := make(map[Capability]struct{})
	result := make([]Capability, 0, len(modules))
	add := func(capability Capability) {
		if _, exists := seen[capability]; exists {
			return
		}
		seen[capability] = struct{}{}
		result = append(result, capability)
	}
	for _, module := range modules {
		switch module {
		case ModuleText:
			add(CapabilityText)
		case ModuleVision:
			add(CapabilityImageInput)
		case ModuleAudioInput:
			add(CapabilityAudioInput)
		case ModuleVideoInput:
			add(CapabilityVideoInput)
		case ModuleFileInput:
			add(CapabilityFileInput)
		case ModuleTools:
			add(CapabilityToolCalling)
		case ModuleStructuredOutput:
			add(CapabilityStructuredOutput)
		}
	}
	return result
}

// ResolveAssistantModels resolves model candidates while automatically applying
// capability requirements implied by the Assistant's enabled modules.
func ResolveAssistantModels(def AssistantDefinition, catalog []ModelDescriptor) ([]ModelRef, error) {
	if err := ValidateAssistantDefinition(def); err != nil {
		return nil, err
	}
	policy := def.ModelPolicy
	seen := make(map[Capability]struct{}, len(policy.RequiredCapabilities))
	for _, capability := range policy.RequiredCapabilities {
		seen[capability] = struct{}{}
	}
	for _, capability := range RequiredCapabilitiesForModules(def.Modules) {
		if _, exists := seen[capability]; exists {
			continue
		}
		policy.RequiredCapabilities = append(policy.RequiredCapabilities, capability)
		seen[capability] = struct{}{}
	}
	return ResolveModelCandidates(policy, catalog)
}

// ResolveModelCandidates returns a deterministic ordered list of eligible
// generation models. Adapters/runtime own health checks and actual invocation.
func ResolveModelCandidates(policy ModelPolicy, catalog []ModelDescriptor) ([]ModelRef, error) {
	if err := ValidateModelPolicy(policy); err != nil {
		return nil, err
	}
	byRef := make(map[ModelRef]ModelDescriptor, len(catalog))
	for _, model := range catalog {
		if err := ValidateModelDescriptor(model); err != nil {
			return nil, err
		}
		if _, exists := byRef[model.Ref]; exists {
			return nil, validationError(CodeInvalidModel, "catalog")
		}
		byRef[model.Ref] = model
	}

	switch policy.Mode {
	case ModelManual:
		model, ok := byRef[*policy.Manual]
		if !ok || !modelEligible(model, policy) {
			return nil, validationError(CodeModelUnavailable, "model_policy.manual")
		}
		return []ModelRef{model.Ref}, nil
	case ModelOrderedFallback:
		resolved := make([]ModelRef, 0, len(policy.Candidates))
		for _, ref := range policy.Candidates {
			if model, ok := byRef[ref]; ok && modelEligible(model, policy) {
				resolved = append(resolved, model.Ref)
			}
		}
		if len(resolved) == 0 {
			return nil, validationError(CodeModelUnavailable, "model_policy.candidates")
		}
		return resolved, nil
	case ModelDynamic:
		models := make([]ModelDescriptor, 0, len(catalog))
		for _, model := range catalog {
			if modelEligible(model, policy) {
				models = append(models, model)
			}
		}
		if len(models) == 0 {
			return nil, validationError(CodeModelUnavailable, "model_policy")
		}
		preference := make(map[AccessTier]int, len(policy.PreferredAccess))
		for index, tier := range policy.PreferredAccess {
			preference[tier] = index
		}
		sort.Slice(models, func(i, j int) bool {
			pi, iPreferred := preference[models[i].Access]
			pj, jPreferred := preference[models[j].Access]
			if iPreferred != jPreferred {
				return iPreferred
			}
			if iPreferred && pi != pj {
				return pi < pj
			}
			return modelRefLess(models[i].Ref, models[j].Ref)
		})
		resolved := make([]ModelRef, len(models))
		for i, model := range models {
			resolved[i] = model.Ref
		}
		return resolved, nil
	default:
		return nil, validationError(CodeInvalidModelPolicy, "model_policy.mode")
	}
}

// ValidateModelDescriptor validates adapter-supplied catalog metadata.
func ValidateModelDescriptor(model ModelDescriptor) error {
	if !validModelRef(model.Ref) || !model.Access.Valid() || !utf8.ValidString(model.DisplayName) || utf8.RuneCountInString(model.DisplayName) > 512 {
		return validationError(CodeInvalidModel, "model")
	}
	if model.ContextWindow < 0 || model.ContextWindow > maxPortableJSONInteger || model.MaxOutputTokens < 0 || model.MaxOutputTokens > maxPortableJSONInteger {
		return validationError(CodeInvalidModel, "model.limits")
	}
	if len(model.Roles) == 0 {
		return validationError(CodeInvalidModel, "model.roles")
	}
	seenRoles := make(map[ModelRole]struct{}, len(model.Roles))
	for _, role := range model.Roles {
		if !role.Valid() {
			return validationError(CodeInvalidModel, "model.roles")
		}
		if _, exists := seenRoles[role]; exists {
			return validationError(CodeInvalidModel, "model.roles")
		}
		seenRoles[role] = struct{}{}
	}
	if err := ValidateCapabilities(model.Capabilities); err != nil {
		return err
	}
	return nil
}

// ResolveModelRole validates that a concrete catalog model exists and advertises
// the requested semantic role, for example embedding or reranking in a RAG stack.
func ResolveModelRole(ref ModelRef, role ModelRole, catalog []ModelDescriptor) (ModelDescriptor, error) {
	if !validModelRef(ref) || !role.Valid() {
		return ModelDescriptor{}, validationError(CodeInvalidModel, "model_role")
	}
	for _, model := range catalog {
		if err := ValidateModelDescriptor(model); err != nil {
			return ModelDescriptor{}, err
		}
		if model.Ref == ref {
			if !containsRole(model.Roles, role) {
				return ModelDescriptor{}, validationError(CodeModelUnavailable, "model_role")
			}
			return model, nil
		}
	}
	return ModelDescriptor{}, validationError(CodeModelUnavailable, "model_role")
}

func modelEligible(model ModelDescriptor, policy ModelPolicy) bool {
	if !containsRole(model.Roles, ModelRoleGeneration) {
		return false
	}
	if len(policy.AllowedProviders) > 0 && !containsProvider(policy.AllowedProviders, model.Ref.Provider) {
		return false
	}
	if len(policy.AllowedAccess) > 0 && !containsAccess(policy.AllowedAccess, model.Access) {
		return false
	}
	for _, required := range policy.RequiredCapabilities {
		if !containsCapability(model.Capabilities, required) {
			return false
		}
	}
	return true
}

func validModelRef(ref ModelRef) bool {
	return validPortableID(string(ref.Provider), 128) && validOpaqueModelID(string(ref.Model), 256)
}

func validOpaqueModelID(value string, max int) bool {
	if value == "" || len(value) > max || !utf8.ValidString(value) {
		return false
	}
	for i := 0; i < len(value); i++ {
		if value[i] < 0x21 || value[i] > 0x7e {
			return false
		}
	}
	return true
}

func modelRefLess(left, right ModelRef) bool {
	if left.Provider != right.Provider {
		return left.Provider < right.Provider
	}
	return left.Model < right.Model
}

func containsRole(values []ModelRole, target ModelRole) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func containsProvider(values []ProviderID, target ProviderID) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func containsAccess(values []AccessTier, target AccessTier) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func containsCapability(values []Capability, target Capability) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func validateProviderIDs(values []ProviderID) error {
	seen := map[ProviderID]struct{}{}
	for _, value := range values {
		if !validPortableID(string(value), 128) {
			return validationError(CodeInvalidModelPolicy, "allowed_providers")
		}
		if _, ok := seen[value]; ok {
			return validationError(CodeInvalidModelPolicy, "allowed_providers")
		}
		seen[value] = struct{}{}
	}
	return nil
}

func validateAccessTiers(values []AccessTier) error {
	seen := map[AccessTier]struct{}{}
	for _, value := range values {
		if !value.Valid() {
			return validationError(CodeInvalidModelPolicy, "access")
		}
		if _, ok := seen[value]; ok {
			return validationError(CodeInvalidModelPolicy, "access")
		}
		seen[value] = struct{}{}
	}
	return nil
}
