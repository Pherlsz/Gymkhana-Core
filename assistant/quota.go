package assistant

// QuotaDimension identifies a normalized provider limit dimension. Providers
// may expose only a subset. The set is intentionally small and extensible by a
// future specification revision rather than passing provider-specific strings.
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
// Values are observations, not promises: providers may reserve capacity or
// update counters asynchronously.
type QuotaWindow struct {
	Dimension     QuotaDimension `json:"dimension"`
	Limit         int64          `json:"limit,omitempty"`
	Used          int64          `json:"used,omitempty"`
	Remaining     int64          `json:"remaining,omitempty"`
	WindowSeconds int64          `json:"window_seconds,omitempty"`
	ResetAtUnix   int64          `json:"reset_at_unix,omitempty"`
}

// QuotaState is normalized runtime quota metadata for the actual provider
// quota scope. It deliberately does not claim quotas are per API key or per
// Assistant: UsageLedgerEntry handles local Assistant/key attribution.
type QuotaState struct {
	Credential CredentialIdentity `json:"credential"`
	Model      *ModelRef          `json:"model,omitempty"`
	Known      bool               `json:"known"`
	Limits     []QuotaWindow      `json:"limits,omitempty"`
}

// UsageLedgerEntry attributes one provider usage observation to a run,
// Assistant, logical credential identity and model without persisting the
// secret-store Reference handle. RunID+Attempt gives consumers an idempotency
// key for avoiding double counting on retries/replays.
type UsageLedgerEntry struct {
	RunID          string             `json:"run_id"`
	Attempt        int64              `json:"attempt"`
	AssistantID    string             `json:"assistant_id"`
	Credential     CredentialIdentity `json:"credential"`
	Model          ModelRef           `json:"model"`
	Usage          Usage              `json:"usage"`
	ObservedAtUnix int64              `json:"observed_at_unix,omitempty"`
}

// ValidateQuotaState validates provider-normalized quota metadata.
func ValidateQuotaState(state QuotaState) error {
	if err := ValidateCredentialIdentity(state.Credential); err != nil {
		return err
	}
	if state.Model != nil {
		if !validModelRef(*state.Model) || state.Model.Provider != state.Credential.Provider {
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
	if err := ValidateCredentialIdentity(entry.Credential); err != nil {
		return err
	}
	if entry.Credential.Provider != entry.Model.Provider {
		return validationError(CodeInvalidCredential, "usage_ledger.credential.provider")
	}
	return ValidateUsage(entry.Usage)
}
