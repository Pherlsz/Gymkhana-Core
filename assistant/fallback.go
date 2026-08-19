package assistant

// FailureClass is a provider-neutral execution failure category used by routing
// and credential fallback policy.
type FailureClass string

const (
	FailureAuth        FailureClass = "auth"
	FailureRateLimit   FailureClass = "rate_limit"
	FailureQuota       FailureClass = "quota"
	FailureTimeout     FailureClass = "timeout"
	FailureUnavailable FailureClass = "unavailable"
	FailureSafety      FailureClass = "safety"
	FailureInvalid     FailureClass = "invalid_request"
	FailureUnknown     FailureClass = "unknown"
)

func (class FailureClass) Valid() bool {
	switch class {
	case FailureAuth, FailureRateLimit, FailureQuota, FailureTimeout, FailureUnavailable, FailureSafety, FailureInvalid, FailureUnknown:
		return true
	default:
		return false
	}
}

// CredentialFallbackPolicy controls whether an ordered credential chain may
// advance after a normalized provider failure.
type CredentialFallbackPolicy struct {
	On []FailureClass `json:"on"`
}

// DefaultCredentialFallbackPolicy advances only for quota/rate-limit style
// exhaustion. Authentication, safety and invalid-request failures are not
// bypassed automatically.
func DefaultCredentialFallbackPolicy() CredentialFallbackPolicy {
	return CredentialFallbackPolicy{On: []FailureClass{FailureQuota, FailureRateLimit}}
}

func ValidateCredentialFallbackPolicy(policy CredentialFallbackPolicy) error {
	seen := make(map[FailureClass]struct{}, len(policy.On))
	for _, class := range policy.On {
		if !class.Valid() {
			return validationError(CodeInvalidCredential, "credential_fallback.on")
		}
		if _, ok := seen[class]; ok {
			return validationError(CodeInvalidCredential, "credential_fallback.on")
		}
		seen[class] = struct{}{}
	}
	return nil
}

func (policy CredentialFallbackPolicy) Allows(class FailureClass) bool {
	for _, allowed := range policy.On {
		if allowed == class {
			return true
		}
	}
	return false
}
