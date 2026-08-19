package assistant

// CredentialMode describes how a provider adapter obtains authorization.
type CredentialMode string

const (
	CredentialNone    CredentialMode = "none"
	CredentialManaged CredentialMode = "managed"
	CredentialBYOK    CredentialMode = "byok"
)

func (mode CredentialMode) Valid() bool {
	switch mode {
	case CredentialNone, CredentialManaged, CredentialBYOK:
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

// CredentialRef is an opaque runtime credential binding.
//
// ID is a stable non-secret logical identity used for usage attribution and
// deduplication. Reference points to an application-owned secret store or
// ephemeral binding and must never contain raw API-key/secret material.
// QuotaScope is an optional opaque application-owned label identifying the
// provider quota/account/project scope shared by credentials. Two credentials
// with the same non-empty QuotaScope may consume the same provider limits.
// CredentialNone is represented by the absence of a CredentialRef at runtime.
type CredentialRef struct {
	ID         string         `json:"id"`
	Provider   ProviderID     `json:"provider"`
	Mode       CredentialMode `json:"mode"`
	Reference  string         `json:"reference"`
	QuotaScope string         `json:"quota_scope,omitempty"`
}

// CredentialIdentity is safe to persist in usage/trace metadata because it
// deliberately excludes the secret-store Reference handle.
type CredentialIdentity struct {
	ID         string         `json:"id"`
	Provider   ProviderID     `json:"provider"`
	Mode       CredentialMode `json:"mode"`
	QuotaScope string         `json:"quota_scope,omitempty"`
}

// Identity returns the persistable identity of a runtime credential binding.
func (ref CredentialRef) Identity() CredentialIdentity {
	return CredentialIdentity{
		ID:         ref.ID,
		Provider:   ref.Provider,
		Mode:       ref.Mode,
		QuotaScope: ref.QuotaScope,
	}
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
	if ref.Mode == CredentialNone {
		return validationError(CodeInvalidCredential, "credential.mode")
	}
	if !validPortableID(ref.ID, 128) || !validPortableID(string(ref.Provider), 128) || !ref.Mode.Valid() || !validOpaqueModelID(ref.Reference, 256) {
		return validationError(CodeInvalidCredential, "credential")
	}
	if ref.QuotaScope != "" && !validOpaqueModelID(ref.QuotaScope, 256) {
		return validationError(CodeInvalidCredential, "credential.quota_scope")
	}
	return nil
}

// ValidateCredentialIdentity validates persistable credential metadata.
func ValidateCredentialIdentity(identity CredentialIdentity) error {
	if identity.Mode == CredentialNone {
		return validationError(CodeInvalidCredential, "credential_identity.mode")
	}
	if !validPortableID(identity.ID, 128) || !validPortableID(string(identity.Provider), 128) || !identity.Mode.Valid() {
		return validationError(CodeInvalidCredential, "credential_identity")
	}
	if identity.QuotaScope != "" && !validOpaqueModelID(identity.QuotaScope, 256) {
		return validationError(CodeInvalidCredential, "credential_identity.quota_scope")
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
