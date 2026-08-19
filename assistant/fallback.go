package assistant

// FailureClass is a provider-neutral execution failure category used by routing,
// retry, and credential fallback policy. Provider adapters map transport/vendor
// failures to these semantic classes without leaking provider payloads into Core.
type FailureClass string

const (
	FailureAuth          FailureClass = "auth"
	FailureRateLimit     FailureClass = "rate_limit"
	FailureQuota         FailureClass = "quota"
	FailureTimeout       FailureClass = "timeout"
	FailureUnavailable   FailureClass = "unavailable"
	FailureNetwork       FailureClass = "network"
	FailureContextLimit  FailureClass = "context_limit"
	FailureUnsupported   FailureClass = "unsupported_capability"
	FailureSafety        FailureClass = "safety"
	FailureInvalid       FailureClass = "invalid_request"
	FailureCancelled     FailureClass = "cancelled"
	FailureUnknown       FailureClass = "unknown"
)

func (class FailureClass) Valid() bool {
	switch class {
	case FailureAuth, FailureRateLimit, FailureQuota, FailureTimeout, FailureUnavailable, FailureNetwork, FailureContextLimit, FailureUnsupported, FailureSafety, FailureInvalid, FailureCancelled, FailureUnknown:
		return true
	default:
		return false
	}
}

// CredentialFallbackPolicy controls whether an ordered credential chain may
// advance after a normalized provider failure.
type CredentialFallbackPolicy struct {
	On                 []FailureClass `json:"on"`
	SkipSameQuotaScope bool           `json:"skip_same_quota_scope"`
}

// DefaultCredentialFallbackPolicy advances only for quota/rate-limit exhaustion
// and skips credentials known to share the same provider quota scope. This is
// important for providers whose limits apply to a project/account rather than
// independently to every API key.
func DefaultCredentialFallbackPolicy() CredentialFallbackPolicy {
	return CredentialFallbackPolicy{
		On:                 []FailureClass{FailureQuota, FailureRateLimit},
		SkipSameQuotaScope: true,
	}
}

func ValidateCredentialFallbackPolicy(policy CredentialFallbackPolicy) error {
	return validateFailureSet(policy.On, "credential_fallback.on")
}

func (policy CredentialFallbackPolicy) Allows(class FailureClass) bool {
	for _, allowed := range policy.On {
		if allowed == class {
			return true
		}
	}
	return false
}

// RoutingFallbackPolicy bounds automatic model/provider failover. It does not
// authorize bypassing safety/auth/invalid-request failures.
type RoutingFallbackPolicy struct {
	ModelOn     []FailureClass `json:"model_on,omitempty"`
	ProviderOn  []FailureClass `json:"provider_on,omitempty"`
	MaxAttempts int64          `json:"max_attempts,omitempty"`
}

// DefaultRoutingFallbackPolicy allows failover only for operational failures.
func DefaultRoutingFallbackPolicy() RoutingFallbackPolicy {
	return RoutingFallbackPolicy{
		ModelOn: []FailureClass{
			FailureRateLimit,
			FailureQuota,
			FailureTimeout,
			FailureUnavailable,
			FailureNetwork,
			FailureContextLimit,
			FailureUnsupported,
		},
		ProviderOn: []FailureClass{
			FailureRateLimit,
			FailureQuota,
			FailureTimeout,
			FailureUnavailable,
			FailureNetwork,
		},
		MaxAttempts: 4,
	}
}

func ValidateRoutingFallbackPolicy(policy RoutingFallbackPolicy) error {
	if err := validateFailureSet(policy.ModelOn, "routing_fallback.model_on"); err != nil {
		return err
	}
	if err := validateFailureSet(policy.ProviderOn, "routing_fallback.provider_on"); err != nil {
		return err
	}
	if policy.MaxAttempts < 0 || policy.MaxAttempts > maxPortableJSONInteger {
		return validationError(CodeInvalidModelPolicy, "routing_fallback.max_attempts")
	}
	return nil
}

func (policy RoutingFallbackPolicy) AllowsModel(class FailureClass) bool {
	return failureSetContains(policy.ModelOn, class)
}

func (policy RoutingFallbackPolicy) AllowsProvider(class FailureClass) bool {
	return failureSetContains(policy.ProviderOn, class)
}

func validateFailureSet(values []FailureClass, field string) error {
	seen := make(map[FailureClass]struct{}, len(values))
	for _, class := range values {
		if !class.Valid() {
			return validationError(CodeInvalidProvider, field)
		}
		if _, exists := seen[class]; exists {
			return validationError(CodeInvalidProvider, field)
		}
		seen[class] = struct{}{}
	}
	return nil
}

func failureSetContains(values []FailureClass, target FailureClass) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
