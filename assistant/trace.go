package assistant

// AttemptTrace records one provider/model attempt without prompt, tool payload,
// media URI, provider error text, or secret-store reference data.
type AttemptTrace struct {
	Index          int64               `json:"index"`
	Model          ModelRef            `json:"model"`
	Credential     *CredentialIdentity `json:"credential,omitempty"`
	Failure        FailureClass        `json:"failure,omitempty"`
	Usage          Usage               `json:"usage,omitempty"`
	DurationMillis int64               `json:"duration_millis,omitempty"`
}

// RunTrace is a content-free portable execution summary suitable for metrics,
// evaluation, routing analysis, and learning evidence references.
type RunTrace struct {
	RunID             string         `json:"run_id"`
	AssistantID       string         `json:"assistant_id"`
	AssistantRevision int64          `json:"assistant_revision,omitempty"`
	EffectiveModules  []ModuleID     `json:"effective_modules"`
	Attempts          []AttemptTrace `json:"attempts,omitempty"`
	RetrievalRounds   int64          `json:"retrieval_rounds,omitempty"`
	ToolCalls         int64          `json:"tool_calls,omitempty"`
	StartedAtUnix     int64          `json:"started_at_unix,omitempty"`
	DurationMillis    int64          `json:"duration_millis,omitempty"`
	FinishReason      FinishReason   `json:"finish_reason,omitempty"`
}

func ValidateRunTrace(trace RunTrace) error {
	if !validOpaqueModelID(trace.RunID, 256) || !validPortableID(trace.AssistantID, 128) {
		return validationError(CodeInvalidTrace, "trace")
	}
	values := [...]int64{trace.AssistantRevision, trace.RetrievalRounds, trace.ToolCalls, trace.StartedAtUnix, trace.DurationMillis}
	for _, value := range values {
		if value < 0 || value > maxPortableJSONInteger {
			return validationError(CodeInvalidTrace, "trace.counter")
		}
	}
	if len(trace.EffectiveModules) == 0 {
		return validationError(CodeInvalidTrace, "trace.effective_modules")
	}
	seenModules := make(map[ModuleID]struct{}, len(trace.EffectiveModules))
	for _, module := range trace.EffectiveModules {
		if !module.Valid() {
			return validationError(CodeInvalidTrace, "trace.effective_modules")
		}
		if _, duplicate := seenModules[module]; duplicate {
			return validationError(CodeInvalidTrace, "trace.effective_modules")
		}
		seenModules[module] = struct{}{}
	}
	for i, attempt := range trace.Attempts {
		if attempt.Index != int64(i) || !validModelRef(attempt.Model) || attempt.DurationMillis < 0 || attempt.DurationMillis > maxPortableJSONInteger {
			return validationError(CodeInvalidTrace, "trace.attempt")
		}
		if attempt.Credential != nil {
			if err := ValidateCredentialIdentity(*attempt.Credential); err != nil {
				return err
			}
			if attempt.Credential.Provider != attempt.Model.Provider {
				return validationError(CodeInvalidTrace, "trace.attempt.credential")
			}
		}
		if attempt.Failure != "" && !attempt.Failure.Valid() {
			return validationError(CodeInvalidTrace, "trace.attempt.failure")
		}
		if err := ValidateUsage(attempt.Usage); err != nil {
			return err
		}
	}
	if trace.FinishReason != "" {
		if err := ValidateFinishReason(trace.FinishReason); err != nil {
			return err
		}
	}
	return nil
}

// TotalUsage sums attempt usage with portable safe-integer overflow checks.
func (trace RunTrace) TotalUsage() (Usage, error) {
	if err := ValidateRunTrace(trace); err != nil {
		return Usage{}, err
	}
	var total Usage
	for _, attempt := range trace.Attempts {
		var err error
		total.InputTokens, err = safeUsageAdd(total.InputTokens, attempt.Usage.InputTokens)
		if err != nil {
			return Usage{}, err
		}
		total.OutputTokens, err = safeUsageAdd(total.OutputTokens, attempt.Usage.OutputTokens)
		if err != nil {
			return Usage{}, err
		}
		total.CachedInputTokens, err = safeUsageAdd(total.CachedInputTokens, attempt.Usage.CachedInputTokens)
		if err != nil {
			return Usage{}, err
		}
		total.ReasoningTokens, err = safeUsageAdd(total.ReasoningTokens, attempt.Usage.ReasoningTokens)
		if err != nil {
			return Usage{}, err
		}
	}
	return total, nil
}

func safeUsageAdd(left, right int64) (int64, error) {
	if left < 0 || right < 0 || left > maxPortableJSONInteger-right {
		return 0, validationError(CodeInvalidUsage, "usage.total")
	}
	return left + right, nil
}
