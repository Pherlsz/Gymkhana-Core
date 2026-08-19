# Assistant configuration and model routing

This document is normative for Core Spec `0.4`.

## Assistant definitions

`AssistantDefinition` is one reusable profile for one application function. Applications may keep multiple definitions such as research, coding, support, OCR, or domain-specific Assistants.

Portable configuration may include:

- stable ID/name/description;
- system/developer instruction blocks;
- default and optional modules;
- built-in skills;
- tool allowlist and `ToolPolicy` values;
- generation-model selection policy;
- credential policy;
- model/provider fallback policy;
- optional RAG policy;
- bounded execution budgets.

Provider SDK objects, HTTP sessions, raw API secrets, persistence rows, product memory, learning/promotion state, and product authorization are not part of the definition.

## Modules

Default modules are listed in `modules`. Disabled-by-default capabilities available to a run are listed in `optional_modules`.

Portable modules in Spec `0.4` are:

- `text`;
- `vision`;
- `audio_input`;
- `video_input`;
- `file_input`;
- `image_output`;
- `audio_output`;
- `video_output`;
- `file_output`;
- `tools`;
- `retrieval`;
- `structured_output`.

Default and optional module sets are disjoint. A run may enable only declared optional modules and disable only declared defaults. The effective set may not be empty.

Generation-model requirements are derived from active modules:

| Module | Required capability |
| --- | --- |
| `text` | `text` |
| `vision` | `image_input` |
| `audio_input` | `audio_input` |
| `video_input` | `video_input` |
| `file_input` | `file_input` |
| `image_output` | `image_output` |
| `audio_output` | `audio_output` |
| `video_output` | `video_output` |
| `file_output` | `file_output` |
| `tools` | `tool_calling` |
| `structured_output` | `structured_output` |

`retrieval` may execute outside the generation model and therefore does not directly imply a model capability.

Declaring the retrieval module requires the corresponding policy even when the module is optional, so enabling it at runtime cannot activate undefined semantics.

Product-owned persistent memory is deliberately not an Assistant module. Hosts may supply selected prior context through messages or retrieval.

## Provider adapters

Concrete providers are represented by adapters. The language-neutral contract does not depend on vendor SDK/HTTP types.

The current Go boundary normalizes:

- provider descriptor;
- live model-catalog discovery;
- non-streaming generation request/response;
- optional failure classification;
- optional quota observation.

Provider-specific models, headers, request IDs, raw errors, retry-after fields, credentials, and pricing metadata remain adapter/runtime concerns unless mapped into an explicit portable contract.

Convenience provider IDs include `openai`, `anthropic`, `google`, `openrouter`, `groq`, `ollama`, `mistral`, and `xai`. The namespace is extensible.

## Model catalogs

Core does not freeze provider model names, prices, free tiers, or availability. Adapters provide current `ModelDescriptor` values containing:

- `(provider, model)` identity;
- display name;
- current access tier: `free`, `paid`, `local`, `unknown`;
- semantic roles: `generation`, `embedding`, `reranking`;
- portable capabilities;
- optional context/output limits.

Model identity is the tuple `(provider, model)`, not a concatenated string. This prevents collisions when either component itself contains separators.

Catalog and candidate sizes are bounded so dynamic routing remains predictable.

## Model selection

Portable modes:

### `manual`

Select exactly one provider/model pair. Manual selection still has to satisfy provider/access/capability restrictions and the active module set.

### `ordered_fallback`

Carry an explicit ordered list of candidates. This is appropriate when the application has benchmark/evaluation evidence for the desired order.

### `dynamic`

Filter the current catalog by allowed providers, access tiers, and required capabilities. `preferred_access` defines deterministic access-tier priority, for example free/local before paid. Within an equal preference class, ordering is deterministic by provider then model.

`preferred_access` must be a subset of `allowed_access` when an allowed set is supplied.

Core deliberately does not guess relative model quality from price or provider name.

## Routing fallback

Selecting candidate order and deciding whether to advance after a failure are separate operations.

`RoutingFallbackPolicy` declares:

- failure classes that allow another model in the same provider;
- failure classes that allow switching provider;
- maximum provider invocation attempts.

The default policy permits bounded fallback for operational conditions such as quota/rate-limit, timeout, unavailability, network failure, context-limit, or unsupported capability as appropriate.

Authentication, safety/refusal policy, invalid-request, and caller-cancellation failures cannot be configured as automatic failover triggers in Spec `0.4`. Switching providers to bypass these failures would change authorization/safety semantics rather than provide operational resilience.

`NextModelCandidate`/equivalent consumes an already-resolved candidate list and distinguishes same-provider model fallback from cross-provider fallback.

## Access tiers

An Assistant may restrict generation to free, paid, local, or mixed catalogs.

Example cost preference:

```json
{
  "mode": "dynamic",
  "allowed_access": ["free", "local", "paid"],
  "preferred_access": ["free", "local", "paid"]
}
```

This is a cost/access ordering only. It is not a quality ranking. Paid fallback cannot occur when `paid` is excluded.

## Credentials and BYOK

Portable credential modes:

- `none` — no runtime credential reference;
- `managed` — host/provider-managed authorization;
- `byok` — host resolves a user/application-owned credential.

A runtime `CredentialRef` carries:

- stable non-secret logical `id`;
- provider;
- mode (`managed` or `byok`);
- namespaced opaque secret-store `reference` such as `session:key_01` or `vault:path`;
- optional `quota_scope`.

`none` is represented by absence of a `CredentialRef`, not by constructing a fake empty credential.

The namespaced-handle rule is a defense against accidentally placing a bare API key in portable configuration. It is not a secret detector; the host remains responsible for never passing raw key material.

`CredentialIdentity` removes `reference` and is the only credential representation intended for portable usage/trace persistence.

## Multiple credentials and quota-aware fallback

A run may provide multiple credentials for the same provider. Their order is significant.

Logical credential IDs and secret-store references must be unique within a provider chain so the runtime does not retry the same binding under another alias.

The default `CredentialFallbackPolicy` advances only for normalized `quota` or `rate_limit` failures. It does not advance after auth, safety, invalid-request, or unrelated failures.

`quota_scope` identifies credentials known by the host/adapter to consume the same provider quota scope. After quota/rate-limit exhaustion, the default policy skips later credentials with the same non-empty scope. This prevents pointless retries when several API keys share one project/account limit.

An empty quota scope means the relationship is unknown; Core does not infer sharing.

See [`OBSERVABILITY.md`](OBSERVABILITY.md) for quota and usage attribution semantics.

## Per-run overrides

`RunOverrides` may temporarily:

- enable declared optional modules;
- disable default modules;
- select one concrete model;
- supply ordered credentials;
- override credential-fallback policy;
- override routing-fallback policy.

Overrides do not mutate the persisted definition and do not bypass model capabilities, credential modes, tool policies, or product authorization.

## Tool and RAG policy

Tools require a declared tool module and are governed by [`TOOLS.md`](TOOLS.md).

Retrieval requires a declared retrieval module and a `RetrievalPolicy`; see [`RAG.md`](RAG.md).

Long-lived memory, task signatures, learning proposals, evaluation evidence, and promotion workflows belong to the host/application rather than the portable Assistant configuration.

## Execution budgets

`ExecutionBudget` bounds turns, tool calls, retrieval rounds, and requested output tokens.

Zero means use a bounded host/runtime default. Zero never means infinite execution.

Spec `0.4` does not define an autonomous unbounded agent loop.
