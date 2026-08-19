package assistant

// QuotaState is normalized runtime quota metadata for one credential binding.
// Providers may expose incomplete data; unknown values remain zero with Known=false.
type QuotaState struct {
	Credential   CredentialRef `json:"credential"`
	AssistantID  string        `json:"assistant_id,omitempty"`
	Known        bool          `json:"known"`
	TokenLimit   int64         `json:"token_limit,omitempty"`
	TokensUsed   int64         `json:"tokens_used,omitempty"`
	TokensRemain int64         `json:"tokens_remaining,omitempty"`
	ResetAtUnix  int64         `json:"reset_at_unix,omitempty"`
}

// UsageLedgerEntry attributes normalized provider usage to one assistant and
// credential handle without exposing secret material.
type UsageLedgerEntry struct {
	AssistantID string        `json:"assistant_id"`
	Credential  CredentialRef `json:"credential"`
	Model       ModelRef      `json:"model"`
	Usage       Usage         `json:"usage"`
}

// ValidateQuotaState validates provider-normalized quota metadata.
func ValidateQuotaState(state QuotaState) error {
	if err := ValidateCredentialRef(state.Credential); err != nil {
		return err
	}
	if state.AssistantID != "" && !validPortableID(state.AssistantID, 128) {
		return validationError(CodeInvalidQuota, "quota.assistant_id")
	}
	values := [...]int64{state.TokenLimit, state.TokensUsed, state.TokensRemain}
	for _, value := range values {
		if value < 0 || value > maxPortableJSONInteger {
			return validationError(CodeInvalidQuota, "quota.tokens")
		}
	}
	if state.ResetAtUnix < 0 {
		return validationError(CodeInvalidQuota, "quota.reset_at_unix")
	}
	if state.Known && state.TokenLimit > 0 && state.TokensUsed > state.TokenLimit {
		return validationError(CodeInvalidQuota, "quota.tokens_used")
	}
	return nil
}

// ValidateUsageLedgerEntry validates usage attribution metadata.
func ValidateUsageLedgerEntry(entry UsageLedgerEntry) error {
	if !validPortableID(entry.AssistantID, 128) || !validModelRef(entry.Model) {
		return validationError(CodeInvalidQuota, "usage_ledger")
	}
	if err := ValidateCredentialRef(entry.Credential); err != nil {
		return err
	}
	if entry.Credential.Provider != entry.Model.Provider {
		return validationError(CodeInvalidCredential, "usage_ledger.credential.provider")
	}
	return ValidateUsage(entry.Usage)
}
