package assistant

// RunOverrides applies temporary choices without mutating the persisted
// AssistantDefinition. Optional modules may be enabled, default modules may be
// disabled, one concrete model may be selected, and ordered provider credential
// handles may be supplied for dynamic/fallback routing.
type RunOverrides struct {
	EnableModules      []ModuleID                 `json:"enable_modules,omitempty"`
	DisableModules     []ModuleID                 `json:"disable_modules,omitempty"`
	Model              *ModelRef                  `json:"model,omitempty"`
	Credentials        []CredentialRef            `json:"credentials,omitempty"`
	CredentialFallback *CredentialFallbackPolicy  `json:"credential_fallback,omitempty"`
	RoutingFallback    *RoutingFallbackPolicy     `json:"routing_fallback,omitempty"`
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
	if len(result) == 0 {
		return nil, validationError(CodeInvalidModule, "run.modules")
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

	seenModules := make(map[ModuleID]struct{}, len(overrides.EnableModules)+len(overrides.DisableModules))
	for _, module := range overrides.EnableModules {
		if _, allowed := optionalModules[module]; !allowed {
			return validationError(CodeInvalidModule, "run.enable_modules")
		}
		if _, duplicate := seenModules[module]; duplicate {
			return validationError(CodeDuplicateModule, "run.modules")
		}
		seenModules[module] = struct{}{}
	}
	for _, module := range overrides.DisableModules {
		if _, allowed := defaultModules[module]; !allowed {
			return validationError(CodeInvalidModule, "run.disable_modules")
		}
		if _, duplicate := seenModules[module]; duplicate {
			return validationError(CodeDuplicateModule, "run.modules")
		}
		seenModules[module] = struct{}{}
	}
	if overrides.Model != nil && !validModelRef(*overrides.Model) {
		return validationError(CodeInvalidModel, "run.model")
	}

	seenCredentials := make(map[CredentialIdentity]struct{}, len(overrides.Credentials))
	for _, credential := range overrides.Credentials {
		if err := ValidateCredentialRef(credential); err != nil {
			return err
		}
		if !def.Credentials.Allows(credential.Mode) {
			return validationError(CodeInvalidCredential, "run.credentials")
		}
		identity := credential.Identity()
		if _, duplicate := seenCredentials[identity]; duplicate {
			return validationError(CodeInvalidCredential, "run.credentials")
		}
		seenCredentials[identity] = struct{}{}
	}
	if overrides.CredentialFallback != nil {
		if err := ValidateCredentialFallbackPolicy(*overrides.CredentialFallback); err != nil {
			return err
		}
	}
	if overrides.RoutingFallback != nil {
		if err := ValidateRoutingFallbackPolicy(*overrides.RoutingFallback); err != nil {
			return err
		}
	}
	return nil
}

// CredentialsForProvider returns the ordered credential chain configured for a
// provider. Caller order is semantically significant.
func CredentialsForProvider(overrides RunOverrides, provider ProviderID) []CredentialRef {
	result := make([]CredentialRef, 0)
	for _, credential := range overrides.Credentials {
		if credential.Provider == provider {
			result = append(result, credential)
		}
	}
	return result
}

// NextCredential returns the next eligible credential in provider order when
// policy permits fallback for the normalized failure class. For quota/rate-limit
// failures, credentials known to share the current quota scope are skipped when
// the policy requests it.
func NextCredential(overrides RunOverrides, provider ProviderID, current int, failure FailureClass) (CredentialRef, int, error) {
	chain := CredentialsForProvider(overrides, provider)
	if len(chain) == 0 || current < -1 || current >= len(chain) {
		return CredentialRef{}, -1, validationError(CodeCredentialExhausted, "run.credentials")
	}
	policy := DefaultCredentialFallbackPolicy()
	if overrides.CredentialFallback != nil {
		policy = *overrides.CredentialFallback
	}
	if current >= 0 {
		if !failure.Valid() || !policy.Allows(failure) {
			return CredentialRef{}, -1, validationError(CodeCredentialExhausted, "run.credentials")
		}
	}

	for next := current + 1; next < len(chain); next++ {
		if current >= 0 && policy.SkipSameQuotaScope && (failure == FailureQuota || failure == FailureRateLimit) {
			currentScope := chain[current].QuotaScope
			if currentScope != "" && chain[next].QuotaScope == currentScope {
				continue
			}
		}
		return chain[next], next, nil
	}
	return CredentialRef{}, -1, validationError(CodeCredentialExhausted, "run.credentials")
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
