# Assistant observability, usage, and quotas

This document is normative for Core Spec `0.3`.

Core separates three concepts that are often incorrectly conflated:

1. provider quota/rate-limit state;
2. usage observed for one model invocation;
3. application attribution of that usage to a run, Assistant, and optional credential identity.

## Credential identity

`CredentialRef.reference` is a secret-store handle and is never persisted in portable usage/trace data.

`CredentialIdentity` contains only:

- logical credential ID;
- provider ID;
- credential mode;
- optional quota-scope label.

The logical ID lets an application report usage per registered BYOK key without storing the key or its secret-store locator in telemetry.

## Quota scope

Provider limits are not assumed to be per API key. Depending on the provider they may apply to a project, account, organization, workspace, model, time window, or another scope.

`QuotaState` therefore identifies:

- provider;
- optional credential identity;
- optional model;
- whether quota is known;
- zero or more normalized quota windows.

Portable quota dimensions are:

- `requests`;
- `input_tokens`;
- `output_tokens`;
- `total_tokens`.

A quota window may expose a limit, used amount, remaining amount, window duration, and reset timestamp when the provider makes those values available.

Missing data stays unknown. Core does not fabricate a quota from local token usage.

## Ordered credential fallback

Multiple credentials for the same provider form an ordered fallback chain.

The default credential fallback policy advances only after normalized `quota` or `rate_limit` failures. It does not fall back on authentication, safety, or invalid-request failures.

When the current credential and a later credential have the same non-empty `quota_scope`, the default policy skips the later credential after quota/rate-limit exhaustion because retrying another key known to share the same provider limit would waste latency and calls.

An empty quota scope means the relationship is unknown; Core does not guess that two keys share limits.

## Usage ledger

`UsageLedgerEntry` attributes one provider-reported `Usage` to:

- `run_id`;
- zero-based attempt index;
- Assistant ID;
- optional credential identity;
- model;
- observation timestamp.

`run_id + attempt` is the portable idempotency key. Persistence layers should enforce uniqueness on that pair so retries, replayed events, or at-least-once delivery do not double-count usage.

Provider-managed/local runs may omit credential identity while still being attributed to an Assistant and model.

## Run trace

`RunTrace` is intentionally content-free. It may record:

- run and Assistant identity/revision;
- effective modules;
- ordered attempts;
- selected model and optional credential identity per attempt;
- normalized failure class;
- provider-reported usage;
- durations;
- retrieval/tool counters;
- final finish reason.

It must not contain:

- prompts or model response text;
- tool arguments/results;
- retrieved evidence content;
- media URIs;
- provider error bodies;
- credential secret-store references or raw keys.

This makes traces suitable for routing analysis, cost/usage dashboards, regression evaluation, and learning evidence without making telemetry a second copy of sensitive conversation content.

## Failure classification

Adapters may implement `FailureClassifier`/equivalent to map provider-specific errors to stable classes such as `rate_limit`, `quota`, `timeout`, `network`, `context_limit`, `unsupported_capability`, `auth`, `safety`, or `invalid_request`.

Context cancellation/deadline should be recognized before provider-specific mapping where the implementation runtime supports those concepts.

Fallback policy must never reclassify authentication, safety, invalid-request, or caller cancellation failures as transient merely to keep trying other providers.
