# Assistant semantics

Core Assistant defines provider-neutral multimodal conversation, configuration, model-routing, BYOK-reference, retrieval, tool-safety, learning, usage/quota, and observability contracts without importing a model-provider SDK into domain code.

The contract is transport-neutral. It does not define provider HTTP payloads, raw authentication secrets, volatile price tables, provider SDK objects, product persistence, or an autonomous unbounded agent loop.

Normative companion documents:

- [`CONFIGURATION.md`](CONFIGURATION.md) — modular Assistants, providers, model routing, BYOK, fallback, and budgets;
- [`STRUCTURED-OUTPUT.md`](STRUCTURED-OUTPUT.md) — strict portable JSON/schema profile;
- [`TOOLS.md`](TOOLS.md) — tool allowlist, side-effect policy, confirmation, and remote/MCP boundary;
- [`RAG.md`](RAG.md) — retrieval and grounding;
- [`LEARNING.md`](LEARNING.md) — task-scoped reviewable Skill Build;
- [`OBSERVABILITY.md`](OBSERVABILITY.md) — usage, quotas, credential attribution, and content-free traces.

## Roles and conversation authority

Portable roles are `system`, `developer`, `user`, `assistant`, and `tool`.

Adapters may translate roles when providers expose a different native vocabulary, but they must not promote user/tool/document content to higher instruction authority.

Role/content invariants:

- system/developer/user messages may not contain `tool_call` or `tool_result` parts;
- assistant messages may contain tool calls but not tool results;
- tool-role messages contain tool results only;
- tool call IDs are unique within a generation request;
- each tool result resolves one outstanding prior call exactly once;
- a new generation request may not contain unresolved client-side tool calls.

## Multimodal content

Portable ordered content parts are:

- `text`;
- `image`;
- `audio`;
- `video`;
- `file`;
- `tool_call`;
- `tool_result`.

Media uses opaque URI references with optional media type/name. Core does not fetch, authorize, upload, transcode, render, or persist media.

Portable model capabilities include text, image/audio/video/file input, image/audio/video/file output, tool calling, structured output, and streaming discovery.

## Multiple modular Assistants

Applications create separate `AssistantDefinition` values for different functions instead of mutating one global Assistant. A definition may configure instructions, default/optional modules, skills, tools and tool policies, model selection, credentials, routing fallback, RAG, memory, learning, and bounded execution budgets.

Per-run overrides may enable declared optional modules, disable defaults, select a concrete model, bind ordered credentials, and override bounded fallback policy without changing the persisted Assistant definition.

## Models and providers

Concrete AI systems are adapters. Model catalogs are discovered at runtime rather than frozen into Core releases.

Portable selection modes are:

- `manual` — one exact provider/model;
- `ordered_fallback` — explicit caller/benchmark-defined order;
- `dynamic` — deterministic filtering/ranking of current catalog metadata.

Active modules contribute required capabilities automatically. For example, an Assistant with `vision + tools` cannot route to a model lacking `image_input + tool_calling`.

Model identity is the pair `(provider, model)`. It is never represented internally by ambiguous string concatenation.

Convenience provider IDs such as OpenAI, Anthropic, Google, OpenRouter, Groq, Ollama, Mistral, and xAI are not a closed enum.

## Credentials and BYOK

Portable credential modes are:

- `none` — no credential reference, typically local/public runtime;
- `managed` — application/provider-managed authorization;
- `byok` — user/application-owned key resolved through an external secret store.

A runtime `CredentialRef` contains:

- logical non-secret credential ID;
- provider;
- mode;
- namespaced secret-store handle;
- optional quota-scope label.

Raw API keys are not portable Core configuration. The secret-store handle is also excluded from usage ledgers and run traces; those carry `CredentialIdentity` only.

Multiple credentials for a provider form an ordered chain. Default BYOK fallback occurs only after normalized quota/rate-limit exhaustion and skips later credentials known to share the same non-empty quota scope.

Provider quotas are not assumed to be per API key. See [`OBSERVABILITY.md`](OBSERVABILITY.md).

## Failure and routing semantics

Adapters may normalize vendor/transport failures into stable classes including auth, rate-limit, quota, timeout, unavailable, network, context-limit, unsupported-capability, safety, invalid-request, cancelled, and unknown.

Automatic model/provider fallback is bounded and permitted only for classes configured by `RoutingFallbackPolicy`. Authentication, safety, invalid-request, and caller-cancellation failures cannot be configured as automatic failover triggers in Spec `0.3`.

## Structured JSON

Tool arguments, tool schemas, and response schemas use strict portable JSON rules. Duplicate keys, ambiguous Unicode surrogate escapes, excessive nesting, and oversized values are rejected.

`portable_json_schema/v1` is deliberately smaller than full JSON Schema. Provider adapters must reject unsupported translation rather than silently weaken schema semantics.

See [`STRUCTURED-OUTPUT.md`](STRUCTURED-OUTPUT.md).

## Tools

Tool calls are data, not authorization. `ToolPolicy` classifies read/write/destructive risk, required confirmation, idempotency, concurrency safety, call limits, timeout, and result-size projection.

Unknown tools default to requiring confirmation. Destructive tools always require confirmation. Product ACL/authorization remains mandatory regardless of confirmation policy.

See [`TOOLS.md`](TOOLS.md).

## RAG

`RetrievalPolicy` supports lexical/vector/hybrid retrieval, original/rewrite/multi-query transformation, bounded candidate/context sets, optional reranking, grounding rules, citations, and optional embedding/reranking model references.

Authorization/filtering occurs before evidence reaches the model. Retrieved content remains untrusted data and never instruction authority.

Evidence content/metadata, scores, counts, and model roles are validated/bounded before admission to portable RAG flows.

## Memory

The optional `memory` module requires an explicit `MemoryPolicy`. Persistent writes are disabled, explicit, or confirmation-gated; Spec `0.3` intentionally provides no unrestricted automatic-write mode.

Read/write scope can be session, Assistant, user, or tenant. Sensitive data is excluded or confirmation-gated according to host classification. Core owns neither identity nor persistence.

## Built-in token economy

`token_economy/v1` requests:

- concise-but-complete responses;
- avoidance of unnecessary restatement;
- safe reuse of stable prior context;
- compact projection of tool results;
- mandatory preservation of instruction hierarchy;
- mandatory preservation of unresolved constraints/state.

It is a semantic profile rather than provider prompt text. It does not define a universal tokenizer, lossy summarizer, message-deletion algorithm, cache ID, or guaranteed savings percentage.

## Skill Build / learning

Learning is task-scoped and proposal-driven. A `TaskSignature` identifies a reusable task family without storing raw prompts. `SkillLearningProposal` references content-free run/evaluation evidence and may suggest retrieval, tool, routing, or behavior improvements.

Core never rewrites an Assistant by itself. Free-form `skill_hint` proposals cannot auto-promote in Spec `0.3`; only low-authority policy scopes may become eligible for host-owned auto-promotion after evidence thresholds.

## Finish reasons

Portable finish reasons are:

- `stop`;
- `length`;
- `tool_calls`;
- `refusal`;
- `paused`;
- `content_filter`;
- `error`;
- `other`.

`refusal` is distinct from a provider/transport error. `paused` represents a resumable provider turn rather than normal completion.

## Usage, quota, and traces

`Usage` contains provider-reported input/output/cached/reasoning token counters bounded to the JSON safe-integer range.

`UsageLedgerEntry` attributes usage to run + attempt + Assistant + model + optional logical credential identity. `RunTrace` records content-free routing/fallback/tool/retrieval/usage metadata for observability and learning without duplicating prompts or provider error bodies.

## Streaming

`streaming` remains a capability in Spec `0.3`; a provider-neutral stream-event wire shape is deliberately deferred until multiple concrete adapters prove the common event lifecycle. Providers differ in text deltas, partial tool arguments, pause/resume semantics, usage timing, and errors.

## Security and privacy

Core validation errors and observability structures must not embed prompt text, raw tool payloads, media URIs, retrieved content, provider error bodies, raw credentials, or secret-store reference handles.

Consumers own encryption, redaction, persistence, retention, permissions, secret resolution, network policy, and user-facing consent.
