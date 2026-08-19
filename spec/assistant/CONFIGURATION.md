# Assistant configuration and model routing

This document is normative for Core Spec `0.4`.

## Assistant definitions

`AssistantDefinition` is one reusable profile for one application function. Applications may keep multiple definitions such as research, coding, support, OCR, or domain-specific Assistants.

Portable configuration may include stable ID/name/description, system/developer instruction blocks, default and optional modules, built-in skills, tool allowlist and policies, generation-model selection, credential and fallback policy, optional RAG policy, and bounded execution budgets.

Provider SDK objects, HTTP sessions, raw API secrets, persistence rows, product memory, learning/promotion state, and product authorization are not part of the definition.

## Modules

Portable modules in Spec `0.4` are `text`, `vision`, `audio_input`, `video_input`, `file_input`, `image_output`, `audio_output`, `video_output`, `file_output`, `tools`, `retrieval`, and `structured_output`.

Default and optional module sets are disjoint. A run may enable only declared optional modules and disable only declared defaults. The effective set may not be empty.

Generation-model requirements are derived from active modules: text -> `text`, vision -> `image_input`, media input/output modules -> matching capability, tools -> `tool_calling`, and structured output -> `structured_output`.

`retrieval` may execute outside the generation model and therefore does not directly imply a model capability. Declaring the retrieval module requires a `RetrievalPolicy`, including when optional, so enabling it at runtime cannot activate undefined semantics.

Product-owned persistent memory is deliberately not an Assistant module. Hosts may supply selected prior context through messages or retrieval.

## Provider adapters

Concrete providers are represented by adapters. The language-neutral contract does not depend on vendor SDK/HTTP types. The current Go boundary normalizes provider descriptor, live model-catalog discovery, non-streaming generation request/response, optional failure classification, and optional quota observation.

Provider-specific models, headers, request IDs, raw errors, retry-after fields, credentials, and pricing metadata remain adapter/runtime concerns unless mapped into an explicit portable contract.

## Model catalogs and selection

Core does not freeze provider model names, prices, free tiers, or availability. Model identity is the tuple `(provider, model)`, not a concatenated string.

Portable modes are `manual`, `ordered_fallback`, and `dynamic`. `preferred_access` defines deterministic access-tier priority and must be a subset of `allowed_access` when an allowed set is supplied. Core does not infer model quality from price or provider name.

## Routing fallback

`RoutingFallbackPolicy` declares failure classes that allow another model in the same provider, failure classes that allow switching provider, and maximum provider invocation attempts.

Authentication, safety/refusal policy, invalid-request, and caller-cancellation failures cannot be configured as automatic failover triggers in Spec `0.4`.

## Credentials and BYOK

Portable credential modes are `none`, `managed`, and `byok`.

A runtime `CredentialRef` carries stable non-secret logical `id`, provider, mode, namespaced opaque secret-store `reference`, and optional `quota_scope`. `none` is represented by absence of a `CredentialRef`.

`CredentialIdentity` removes `reference` and is the only credential representation intended for portable usage/trace persistence. Multiple credentials for the same provider form an ordered fallback chain; the default policy advances only for normalized `quota` or `rate_limit` failures and skips credentials known to share the same non-empty quota scope.

See [`OBSERVABILITY.md`](OBSERVABILITY.md).

## Per-run overrides

`RunOverrides` may temporarily enable declared optional modules, disable default modules, select one concrete model, supply ordered credentials, override credential-fallback policy, and override routing-fallback policy.

Overrides do not mutate the Assistant definition and do not bypass model capabilities, credential modes, tool policies, or product authorization.

## Tool and RAG policy

Tools require a declared tool module and are governed by [`TOOLS.md`](TOOLS.md). Retrieval requires a declared retrieval module and a `RetrievalPolicy`; see [`RAG.md`](RAG.md).

Long-lived memory, task signatures, learning proposals, evaluation evidence, and promotion workflows belong to the host/application rather than the portable Assistant configuration.

## Execution budgets

`ExecutionBudget` bounds turns, tool calls, retrieval rounds, and requested output tokens. Zero means use a bounded host/runtime default; zero never means infinite execution. Spec `0.4` does not define an autonomous unbounded agent loop.
