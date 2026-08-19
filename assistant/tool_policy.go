package assistant

// ToolEffect classifies the side-effect risk of invoking a tool. Unknown is the
// safe default for tools that have not been explicitly classified.
type ToolEffect string

const (
	ToolEffectUnknown     ToolEffect = "unknown"
	ToolEffectReadOnly    ToolEffect = "read_only"
	ToolEffectWrite       ToolEffect = "write"
	ToolEffectDestructive ToolEffect = "destructive"
)

func (effect ToolEffect) Valid() bool {
	switch effect {
	case ToolEffectUnknown, ToolEffectReadOnly, ToolEffectWrite, ToolEffectDestructive:
		return true
	default:
		return false
	}
}

// ToolConfirmation controls whether application/user authorization is required
// before a tool may execute. It does not replace product permission checks.
type ToolConfirmation string

const (
	ToolConfirmNever      ToolConfirmation = "never"
	ToolConfirmSideEffect ToolConfirmation = "side_effect"
	ToolConfirmAlways     ToolConfirmation = "always"
)

func (mode ToolConfirmation) Valid() bool {
	switch mode {
	case ToolConfirmNever, ToolConfirmSideEffect, ToolConfirmAlways:
		return true
	default:
		return false
	}
}

// ToolPolicy adds execution semantics to one tool name without coupling Core to
// the implementation or authorization system. MaxCalls=0 means use the broader
// run/application limit rather than unlimited execution.
type ToolPolicy struct {
	Name         string           `json:"name"`
	Effect       ToolEffect       `json:"effect"`
	Confirmation ToolConfirmation `json:"confirmation"`
	ParallelSafe bool             `json:"parallel_safe,omitempty"`
	Idempotent   bool             `json:"idempotent,omitempty"`
	MaxCalls     int64            `json:"max_calls,omitempty"`
}

// DefaultToolPolicy is deliberately conservative for an unclassified tool.
func DefaultToolPolicy(name string) ToolPolicy {
	return ToolPolicy{
		Name:         name,
		Effect:       ToolEffectUnknown,
		Confirmation: ToolConfirmAlways,
	}
}

func ValidateToolPolicy(policy ToolPolicy) error {
	if err := ValidateToolName(policy.Name); err != nil {
		return err
	}
	if !policy.Effect.Valid() || !policy.Confirmation.Valid() || policy.MaxCalls < 0 || policy.MaxCalls > maxPortableJSONInteger {
		return validationError(CodeInvalidToolPolicy, "tool_policy")
	}
	switch policy.Effect {
	case ToolEffectUnknown:
		if policy.Confirmation != ToolConfirmAlways {
			return validationError(CodeInvalidToolPolicy, "tool_policy.confirmation")
		}
	case ToolEffectWrite:
		if policy.Confirmation == ToolConfirmNever {
			return validationError(CodeInvalidToolPolicy, "tool_policy.confirmation")
		}
	case ToolEffectDestructive:
		if policy.Confirmation != ToolConfirmAlways {
			return validationError(CodeInvalidToolPolicy, "tool_policy.confirmation")
		}
	}
	return nil
}

// ValidateToolPolicies validates explicit policies against the Assistant's
// declared tool allowlist. Missing policies are legal and resolve to the safe
// DefaultToolPolicy.
func ValidateToolPolicies(allowedTools []string, policies []ToolPolicy) error {
	allowed := make(map[string]struct{}, len(allowedTools))
	for _, name := range allowedTools {
		allowed[name] = struct{}{}
	}
	seen := make(map[string]struct{}, len(policies))
	for _, policy := range policies {
		if err := ValidateToolPolicy(policy); err != nil {
			return err
		}
		if _, ok := allowed[policy.Name]; !ok {
			return validationError(CodeInvalidToolPolicy, "tool_policy.name")
		}
		if _, duplicate := seen[policy.Name]; duplicate {
			return validationError(CodeInvalidToolPolicy, "tool_policy.name")
		}
		seen[policy.Name] = struct{}{}
	}
	return nil
}

// ResolveToolPolicy returns an explicit policy or the conservative default.
func ResolveToolPolicy(name string, policies []ToolPolicy) (ToolPolicy, error) {
	if err := ValidateToolName(name); err != nil {
		return ToolPolicy{}, err
	}
	for _, policy := range policies {
		if policy.Name == name {
			if err := ValidateToolPolicy(policy); err != nil {
				return ToolPolicy{}, err
			}
			return policy, nil
		}
	}
	return DefaultToolPolicy(name), nil
}
