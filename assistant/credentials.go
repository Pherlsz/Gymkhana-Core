package assistant

// CredentialMode describes how a provider adapter obtains authorization.
type CredentialMode string

const (
	CredentialManaged CredentialMode = "managed"
	CredentialBYOK    CredentialMode = "byok"
)

func (mode CredentialMode) Valid() bool {
	switch mode {
	case CredentialManaged, CredentialBYOK:
		return true
	default:
		return false
	}
}

// CredentialPolicy declares which credential modes an Assistant accepts.
// Secret material is never part of the Assistant definition.
type CredentialPolicy struct {
	AllowedModes []CredentialMode `json:"allowed_modes"`
}

// CredentialRef is an opaque runtime handle. Reference must point to an
// application-owned secure store or ephemeral credential binding and must not
// contain raw API-key/secret material.
type CredentialRef struct {
	Provider  ProviderID     `json:"provider"`
	Mode      CredentialMode `json:"mode"`
	Reference string         `json:"reference"`
}

// ValidateCredentialPolicy validates allowed authorization modes.
func ValidateCredentialPolicy(policy CredentialPolicy) error {
	if len(policy.AllowedModes) == 0 {
		return validationError(CodeEmpty, "credentials.allowed_modes")
	}
	seen := make(map[CredentialMode]struct{}, len(policy.AllowedModes))
	for _, mode := range policy.AllowedModes {
		if !mode.Valid() {
			return validationError(CodeInvalidCredential, "credentials.allowed_modes")
		}
		if _, exists := seen[mode]; exists {
			return validationError(CodeInvalidCredential, "credentials.allowed_modes")
		}
		seen[mode] = struct{}{}
	}
	return nil
}

// ValidateCredentialRef validates an opaque runtime credential binding.
func ValidateCredentialRef(ref CredentialRef) error {
	if !validPortableID(string(ref.Provider), 128) || !ref.Mode.Valid() || !validPortableID(ref.Reference, 256) {
		return validationError(CodeInvalidCredential, "credential")
	}
	return nil
}

// Allows reports whether the policy accepts the binding mode.
func (policy CredentialPolicy) Allows(mode CredentialMode) bool {
	for _, allowed := range policy.AllowedModes {
		if allowed == mode {
			return true
		}
	}
	return false
}
