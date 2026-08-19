package assistant

// RunOverrides applies temporary choices without mutating the persisted
// AssistantDefinition. Optional modules may be enabled, default modules may be
// disabled, and one concrete model/credential may be selected for the run.
type RunOverrides struct {
	EnableModules  []ModuleID     `json:"enable_modules,omitempty"`
	DisableModules []ModuleID     `json:"disable_modules,omitempty"`
	Model          *ModelRef      `json:"model,omitempty"`
	Credential     *CredentialRef `json:"credential,omitempty"`
}

// EffectiveModules returns deterministic effective modules for one run.
func EffectiveModules(def AssistantDefinition, overrides RunOverrides) ([]ModuleID, error) {
	if err := ValidateAssistantDefinition(def); err != nil {
		return nil, err
	}
	if err := ValidateRunOverrides(def, overrides); err != nil {
		return nil, err
	}

	enabled := make(map[ModuleID]bool, len(def.Modules)+len(def.OptionalModules))
	order := make([]ModuleID, 0, len(def.Modules)+len(def.OptionalModules))
	for _, module := range def.Modules {
		enabled[module] = true
		order = append(order, module)
	}
	for _, module := range def.OptionalModules {
		enabled[module] = false
		order = append(order, module)
	}
	for _, module := range overrides.EnableModules {
		enabled[module] = true
	}
	for _, module := range overrides.DisableModules {
		enabled[module] = false
	}

	result := make([]ModuleID, 0, len(order))
	for _, module := range order {
		if enabled[module] {
			result = append(result, module)
		}
	}
	return result, nil
}

// ValidateRunOverrides ensures a run can only toggle modules declared by the
// Assistant profile and can only use credential modes allowed by the profile.
func ValidateRunOverrides(def AssistantDefinition, overrides RunOverrides) error {
	defaultModules := make(map[ModuleID]struct{}, len(def.Modules))
	for _, module := range def.Modules {
		defaultModules[module] = struct{}{}
	}
	optionalModules := make(map[ModuleID]struct{}, len(def.OptionalModules))
	for _, module := range def.OptionalModules {
		optionalModules[module] = struct{}{}
	}

	seen := make(map[ModuleID]struct{}, len(overrides.EnableModules)+len(overrides.DisableModules))
	for _, module := range overrides.EnableModules {
		if _, allowed := optionalModules[module]; !allowed {
			return validationError(CodeInvalidModule, "run.enable_modules")
		}
		if _, duplicate := seen[module]; duplicate {
			return validationError(CodeDuplicateModule, "run.modules")
		}
		seen[module] = struct{}{}
	}
	for _, module := range overrides.DisableModules {
		if _, allowed := defaultModules[module]; !allowed {
			return validationError(CodeInvalidModule, "run.disable_modules")
		}
		if _, duplicate := seen[module]; duplicate {
			return validationError(CodeDuplicateModule, "run.modules")
		}
		seen[module] = struct{}{}
	}
	if overrides.Model != nil && !validModelRef(*overrides.Model) {
		return validationError(CodeInvalidModel, "run.model")
	}
	if overrides.Credential != nil {
		if err := ValidateCredentialRef(*overrides.Credential); err != nil {
			return err
		}
		if !def.Credentials.Allows(overrides.Credential.Mode) {
			return validationError(CodeInvalidCredential, "run.credential")
		}
		if overrides.Model != nil && overrides.Credential.Provider != overrides.Model.Provider {
			return validationError(CodeInvalidCredential, "run.credential.provider")
		}
	}
	return nil
}

// ResolveAssistantModelsForRun applies module and manual model overrides before
// resolving the current adapter-supplied model catalog.
func ResolveAssistantModelsForRun(def AssistantDefinition, overrides RunOverrides, catalog []ModelDescriptor) ([]ModelRef, error) {
	modules, err := EffectiveModules(def, overrides)
	if err != nil {
		return nil, err
	}
	policy := def.ModelPolicy
	if overrides.Model != nil {
		policy.Mode = ModelManual
		policy.Manual = overrides.Model
		policy.Candidates = nil
		policy.PreferredAccess = nil
	}

	seen := make(map[Capability]struct{}, len(policy.RequiredCapabilities))
	for _, capability := range policy.RequiredCapabilities {
		seen[capability] = struct{}{}
	}
	for _, capability := range RequiredCapabilitiesForModules(modules) {
		if _, exists := seen[capability]; exists {
			continue
		}
		policy.RequiredCapabilities = append(policy.RequiredCapabilities, capability)
		seen[capability] = struct{}{}
	}
	return ResolveModelCandidates(policy, catalog)
}
