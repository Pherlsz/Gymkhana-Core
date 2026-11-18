package assistant

import (
	"slices"
	"sync"
)

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

type usageLedgerKey struct {
	RunID   string
	Attempt int64
}

// UsageLedger is a concurrency-safe, idempotent in-memory usage attribution
// ledger keyed by RunID+Attempt.
type UsageLedger struct {
	mu      sync.RWMutex
	entries []UsageLedgerEntry
	byKey   map[usageLedgerKey]int
}

func NewUsageLedger() *UsageLedger {
	return &UsageLedger{byKey: make(map[usageLedgerKey]int)}
}

// Record validates and records one usage observation. It returns (true, nil)
// when recorded for the first time, (false, nil) when an identical RunID+Attempt
// replay is observed, or CodeInvalidQuota when a conflicting payload reuses the
// same RunID+Attempt key.
func (ledger *UsageLedger) Record(entry UsageLedgerEntry) (bool, error) {
	if ledger == nil {
		return false, validationError(CodeInvalidQuota, "usage_ledger")
	}
	if err := ValidateUsageLedgerEntry(entry); err != nil {
		return false, err
	}
	cloned := cloneUsageLedgerEntry(entry)
	key := usageLedgerKey{RunID: cloned.RunID, Attempt: cloned.Attempt}

	ledger.mu.Lock()
	defer ledger.mu.Unlock()
	if idx, exists := ledger.byKey[key]; exists {
		if !usageLedgerEntriesEqual(ledger.entries[idx], cloned) {
			return false, validationError(CodeInvalidQuota, "usage_ledger.attempt")
		}
		return false, nil
	}
	if ledger.byKey == nil {
		ledger.byKey = make(map[usageLedgerKey]int)
	}
	ledger.byKey[key] = len(ledger.entries)
	ledger.entries = append(ledger.entries, cloned)
	return true, nil
}

// Entries returns a defensive copy of all recorded entries in insertion order.
func (ledger *UsageLedger) Entries() []UsageLedgerEntry {
	if ledger == nil {
		return nil
	}
	ledger.mu.RLock()
	defer ledger.mu.RUnlock()
	result := make([]UsageLedgerEntry, len(ledger.entries))
	for i, entry := range ledger.entries {
		result[i] = cloneUsageLedgerEntry(entry)
	}
	return result
}

// TotalUsageForAssistant sums recorded usage for one AssistantID with portable
// safe-integer overflow checks.
func (ledger *UsageLedger) TotalUsageForAssistant(assistantID string) (Usage, error) {
	if !validPortableID(assistantID, 128) {
		return Usage{}, validationError(CodeInvalidAssistant, "assistant_id")
	}
	if ledger == nil {
		return Usage{}, nil
	}
	ledger.mu.RLock()
	defer ledger.mu.RUnlock()
	var total Usage
	for _, entry := range ledger.entries {
		if entry.AssistantID != assistantID {
			continue
		}
		var err error
		if total, err = addUsage(total, entry.Usage); err != nil {
			return Usage{}, err
		}
	}
	return total, nil
}

func cloneUsageLedgerEntry(entry UsageLedgerEntry) UsageLedgerEntry {
	if entry.Credential != nil {
		cred := *entry.Credential
		entry.Credential = &cred
	}
	return entry
}

func usageLedgerEntriesEqual(left, right UsageLedgerEntry) bool {
	if left.RunID != right.RunID || left.Attempt != right.Attempt || left.AssistantID != right.AssistantID || left.Model != right.Model || left.Usage != right.Usage || left.ObservedAtUnix != right.ObservedAtUnix {
		return false
	}
	if (left.Credential == nil) != (right.Credential == nil) {
		return false
	}
	if left.Credential != nil && *left.Credential != *right.Credential {
		return false
	}
	return true
}

func cloneQuotaState(state QuotaState) QuotaState {
	if state.Credential != nil {
		cred := *state.Credential
		state.Credential = &cred
	}
	if state.Model != nil {
		model := *state.Model
		state.Model = &model
	}
	state.Limits = slices.Clone(state.Limits)
	return state
}
