# Assistant observability, usage, and quotas

This document is normative for Core Spec `0.4`.

Core separates provider quota/rate-limit state, usage observed for one model invocation, and application attribution of that usage to a run, Assistant, and optional credential identity.

## Credential identity

`CredentialRef.reference` is a secret-store handle and is never persisted in portable usage/trace data. `CredentialIdentity` contains logical credential ID, provider ID, credential mode, and optional quota-scope label.

## Quota scope

Provider limits are not assumed to be per API key. `QuotaState` identifies provider, optional credential identity, optional model, whether quota is known, and zero or more normalized quota windows.

Portable quota dimensions are `requests`, `input_tokens`, `output_tokens`, and `total_tokens`. Missing data stays unknown; Core does not fabricate quota from local token usage.

## Ordered credential fallback

Multiple credentials for the same provider form an ordered fallback chain. The default credential fallback policy advances only after normalized `quota` or `rate_limit` failures.

When the current credential and a later credential have the same non-empty `quota_scope`, the default policy skips the later credential after quota/rate-limit exhaustion. An empty quota scope means the relationship is unknown.

## Usage ledger

`UsageLedgerEntry` attributes one provider-reported `Usage` to `run_id`, zero-based attempt index, Assistant ID, optional credential identity, model, and optional observation timestamp.

`run_id + attempt` is the portable idempotency key. A persistence layer may enforce uniqueness on that pair, but persistence itself is not a Core responsibility.

## Run trace

`RunTrace` is intentionally content-free. It may record run and Assistant identity/revision, effective modules, ordered attempts, selected model and optional credential identity per attempt, normalized failure class, provider-reported usage, durations, retrieval/tool counters, and final finish reason.

It must not contain prompts or model response text, tool arguments/results, retrieved evidence content, media URIs, provider error bodies, credential secret-store references, or raw keys.

This makes traces suitable for routing analysis, cost/usage dashboards, regression evaluation, and operational diagnostics without making telemetry a second copy of sensitive conversation content.

## Failure classification

Adapters may map provider-specific errors to stable classes such as `rate_limit`, `quota`, `timeout`, `network`, `context_limit`, `unsupported_capability`, `auth`, `safety`, or `invalid_request`.

Fallback policy must never reclassify authentication, safety, invalid-request, or caller cancellation failures as transient merely to keep trying other providers.
