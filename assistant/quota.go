package assistant

// QuotaDimension identifies a normalized provider limit dimension. Providers
// may expose only a subset.
type QuotaDimension string

const (
	QuotaRequests     QuotaDimension = "requests"
	QuotaInputTokens  QuotaDimension = "input_tokens"
	QuotaOutputTokens QuotaDimension = "output_tokens"
	QuotaTotalTokens  QuotaDimension = "total_tokens"
)

func (dimension QuotaDimension) Valid() bool {
	switch dimension {
	case QuotaRequests, QuotaInputTokens, QuotaOutputTokens, QuotaTotalTokens:
		return true
	default:
		return false
	}
}

// QuotaWindow is one normalized provider quota/rate-limit observation. Rolling
// and fixed windows are both represented by optional WindowSeconds/ResetAtUnix.
type QuotaWindow struct {
	Dimension     QuotaDimension `json:"dimension"`
	Limit         int64          `json:"limit,omitempty"`
	Used          int64          `json:"used,omitempty"`
	Remaining     int64          `json:"remaining,omitempty"`
	WindowSeconds int64          `json:"window_seconds,omitempty"`
	ResetAtUnix   int64          `json:"reset_at_unix,omitempty"`
}

// QuotaState is normalized runtime quota metadata for the actual provider quota
// scope. Credential is optional because managed/local adapters may not expose a
// runtime CredentialRef. It does not claim quotas are per Assistant.
type QuotaState struct {
	Provider   ProviderID          `json:"provider"`
	Credential *CredentialIdentity `json:"credential,omitempty"`
	Model      *ModelRef           `json:"model,omitempty"`
	Known      bool                `json:"known"`
	Limits     []QuotaWindow       `json:"limits,omitempty"`
}

// UsageLedgerEntry attributes one provider usage observation to a run,
// Assistant, optional logical credential identity and model without persisting
// the secret-store Reference handle. RunID+Attempt is an idempotency key for
// avoiding double counting on retries/replays.
type UsageLedgerEntry struct {
	RunID          string              `json:"run_id"`
	Attempt        int64               `json:"attempt"`
	AssistantID    string              `json:"assistant_id"`
	Credential     *CredentialIdentity `json:"credential,omitempty"`
	Model          ModelRef            `json:"model"`
	Usage          Usage               `json:"usage"`
	ObservedAtUnix int64               `json:"observed_at_unix,omitempty"`
}

// ValidateQuotaState validates provider-normalized quota metadata.
func ValidateQuotaState(state QuotaState) error {
	if !validPortableID(string(state.Provider), 128) {
		return validationError(CodeInvalidQuota, "quota.provider")
	}
	if state.Credential != nil {
		if err := ValidateCredentialIdentity(*state.Credential); err != nil {
			return err
		}
		if state.Credential.Provider != state.Provider {
			return validationError(CodeInvalidQuota, "quota.credential.provider")
		}
	}
	if state.Model != nil {
		if !validModelRef(*state.Model) || state.Model.Provider != state.Provider {
			return validationError(CodeInvalidQuota, "quota.model")
		}
	}
	if !state.Known && len(state.Limits) > 0 {
		return validationError(CodeInvalidQuota, "quota.known")
	}
	seen := make(map[QuotaDimension]struct{}, len(state.Limits))
	for _, limit := range state.Limits {
		if !limit.Dimension.Valid() {
			return validationError(CodeInvalidQuota, "quota.dimension")
		}
		if _, duplicate := seen[limit.Dimension]; duplicate {
			return validationError(CodeInvalidQuota, "quota.dimension")
		}
		seen[limit.Dimension] = struct{}{}
		values := [...]int64{limit.Limit, limit.Used, limit.Remaining, limit.WindowSeconds, limit.ResetAtUnix}
		for _, value := range values {
			if value < 0 || value > maxPortableJSONInteger {
				return validationError(CodeInvalidQuota, "quota.limit")
			}
		}
		if limit.Limit > 0 && limit.Remaining > limit.Limit {
			return validationError(CodeInvalidQuota, "quota.remaining")
		}
	}
	return nil
}

// ValidateUsageLedgerEntry validates usage attribution metadata.
func ValidateUsageLedgerEntry(entry UsageLedgerEntry) error {
	if !validOpaqueModelID(entry.RunID, 256) || !validPortableID(entry.AssistantID, 128) || !validModelRef(entry.Model) {
		return validationError(CodeInvalidQuota, "usage_ledger")
	}
	if entry.Attempt < 0 || entry.Attempt > maxPortableJSONInteger || entry.ObservedAtUnix < 0 || entry.ObservedAtUnix > maxPortableJSONInteger {
		return validationError(CodeInvalidQuota, "usage_ledger.attempt")
	}
	if entry.Credential != nil {
		if err := ValidateCredentialIdentity(*entry.Credential); err != nil {
			return err
		}
		if entry.Credential.Provider != entry.Model.Provider {
			return validationError(CodeInvalidCredential, "usage_ledger.credential.provider")
		}
	}
	return ValidateUsage(entry.Usage)
}
