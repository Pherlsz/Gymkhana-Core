# Assistant Foundation hardening review

Status: Core Spec `0.3` / Go `0.5.0` workstream (`feat/assistant-foundation`).

This document records correctness, security, portability, and performance issues discovered while stress-testing the initial Assistant contracts before the first public release. It is decision history rather than a replacement for normative material under `spec/assistant/` and `spec/json/`.

## Correctness bugs found and addressed

### Model identity collisions

String-concatenated model keys could collide when provider/model IDs contained the separator. Example: `(a, b/c)` and `(a/b, c)` both serialize informally as `a/b/c`.

Resolution: `ModelRef` itself is the map key and ordering compares provider then model independently.

### Tool-call lifecycle was locally valid but globally inconsistent

Validating each `Message` independently allowed orphan tool results, duplicate call IDs, user-authored tool calls, and unresolved client-side tool calls to survive until a provider rejected them.

Resolution: role/content rules plus `ValidateConversation` enforce call/result linkage across the whole generation request.

### Tool schema and tool-call argument shape could disagree

A tool schema could describe a non-object root while `ToolCall.arguments` is always an object.

Resolution: Assistant requires the generic portable schema to have an object root or use the explicit `{}` unconstrained-object compatibility form.

### JSON duplicate keys were runtime-dependent

Standard JSON parsers differ on duplicate property names. Depending on runtime, first/last/merged/error behavior can differ and create cross-language security/semantic drift.

Resolution: the generic `portablejson` package rejects duplicate keys at every nesting level before language-native object decoding is treated as semantic input.

### Malformed Unicode could normalize differently between runtimes

Escaped unpaired UTF-16 surrogates can be replaced or rejected differently by language JSON decoders.

Resolution: `portablejson` rejects unpaired surrogate escapes before semantic decoding.

### Decimal comparison could drift through floating-point rounding

Schema numeric bounds and enum values were initially compared through `float64`, which could make distinct JSON decimals compare equal after binary floating-point rounding.

Resolution: portable schema/instance number comparison uses exact rational decimal semantics while keeping the portable magnitude bound.

### Nullable enum could accidentally accept undeclared null

A nullable type was initially able to return successfully for `null` before applying an `enum` restriction.

Resolution: enum membership is evaluated before the nullable early return, so `null` is valid only when the enum itself includes it.

### Provider response could invoke undeclared tools

A syntactically valid tool call from a provider was not sufficient proof that the current request actually exposed that tool.

Resolution: `ValidateGenerationExchange` checks provider-returned tool calls against the request allowlist and validates the finish/tool-call relationship.

### BYOK fallback retried the same logical quota

Multiple API keys may share one project/account/provider quota. Blindly moving key1 -> key2 after quota exhaustion can repeat the same guaranteed failure.

Resolution: credentials carry optional non-secret `quota_scope`; default quota/rate-limit fallback skips later bindings known to share that scope.

### Credential telemetry could leak secret-store handles

Persisting `CredentialRef` in usage/traces would preserve the application-owned secret-store locator even though it is not raw key material.

Resolution: persistent telemetry uses `CredentialIdentity`, which excludes `reference`.

### Duplicate credential aliases could retry the same secret

Different logical IDs could point at the same secret-store handle and appear to form a useful fallback chain.

Resolution: run validation rejects duplicate logical credential IDs and duplicate handles inside a provider chain.

### Quota model assumed one window per dimension

Providers can expose several simultaneous windows such as requests/minute and requests/day.

Resolution: a quota window is unique by `(dimension, window_seconds)`, not dimension alone.

### Memory existed without semantics

The `memory` module was originally configurable without defining read/write/retention authority.

Resolution: enabling or making memory optional now requires `MemoryPolicy`; persistent writes are disabled, explicit, or confirmation-gated. Spec `0.3` has no unrestricted automatic persistent-write mode.

### Learned behavior could become persistent prompt injection

A free-form lesson derived from model/document output could be automatically promoted into future behavior and thereby persist an injected instruction.

Resolution: learning is task-scoped, evidence-backed, advisory, and revisioned. `skill_hint` is never auto-promotable in Spec `0.3`; Core never mutates an `AssistantDefinition` itself.

## Safety boundaries added

### Tool policy

Tools are classified independently from provider function-calling syntax:

- `unknown`;
- `read_only`;
- `write`;
- `destructive`.

Unknown and destructive tools require confirmation. Writes may not use `confirmation=never`. Product authorization remains mandatory regardless of confirmation mode.

Tool policies may also bound calls, execution time, result projection size, idempotency, and concurrent execution safety.

### Failure taxonomy and bounded failover

Provider/transport failures normalize to portable classes before failover policy is applied. Model/provider failover is bounded and cannot be configured to bypass authentication, safety, invalid-request, or caller-cancellation failures.

Credential failover is separate from model/provider failover.

### Content-free traces

`RunTrace` records only operational metadata required for debugging, routing analysis, evaluation, and learning. It deliberately excludes prompts, response text, tool payloads, evidence content, media URIs, raw provider error bodies, raw keys, and secret-store references.

## Portability decisions

### Portable JSON is generic Core, not Assistant

The first Assistant implementation contained strict JSON and `portable_json_schema/v1` directly under `assistant`. The semantics are useful beyond LLMs, including future OCR/extraction, matching, solver, workflow, and other structured-data domains.

Resolution: the implementation and normative ownership moved to the generic `portablejson` package and `spec/json/`. Assistant retains compatibility wrappers for the original `assistant.ValidatePortable*` API and keeps only Assistant-specific integration such as object-root tool schemas and `ValidateToolArguments`.

This preserves code/API value while preventing future Core domains from importing Assistant solely to validate structured JSON.

### `portable_json_schema/v1`

Full JSON Schema is not treated as a provider-neutral capability. Core defines a conservative subset and requires consumers/adapters to reject translations they cannot preserve instead of silently weakening validation.

The profile is strict by design:

- explicit type;
- bounded nesting/properties/arrays;
- no unknown or type-inapplicable keywords;
- strict typed objects use `additionalProperties=false`;
- all declared object properties are required; nullable fields express optional semantic values;
- `{}` is the explicit unconstrained-object form;
- domain-specific consumers may impose stricter roots, such as Assistant tool schemas requiring an object.

### JSON-safe integer counters

Usage, budget, quota, revision, and similar portable counters are bounded to `2^53-1` where exact JSON/TypeScript interoperability matters.

### Provider quota is not Assistant quota

Provider quota observations remain provider/account/project/model scoped according to what an adapter can actually prove. Local `UsageLedgerEntry` separately attributes observed usage to a run, Assistant, logical credential identity, and model.

This lets a UI show per-Assistant and per-key local consumption without falsely claiming the provider granted an independent token allowance to each key.

## Performance controls

Untrusted collection sizes are bounded before expensive traversal/allocation:

- model candidate/catalog sizes;
- Assistant catalog size;
- messages and content parts;
- tool definitions/results;
- RAG candidate/context/evidence/metadata sizes;
- JSON bytes/depth/node count;
- schema depth/property/array limits;
- learning evidence count;
- quota windows.

Dynamic model routing remains deterministic and uses bounded catalog sorting. `assistant/benchmark_test.go` tracks the 10k-model routing path; `portablejson/benchmark_test.go` tracks strict JSON/schema validation. Benchmarks remain local and do not add hosted CI cost.

## Deliberately deferred from `0.5.0`

### Concrete provider adapters

OpenAI, Anthropic, Google, OpenRouter, Groq, Ollama, Mistral, xAI, and other network adapters should be implemented after this semantic boundary is stable. Their release cadence must not force Core specification churn.

### Provider-neutral streaming events

The `streaming` capability exists, but event/delta semantics remain deferred until multiple real adapters exercise text, partial tool arguments, pause/resume, usage timing, refusals, and errors.

### Context planner and caching

`token_economy/v1` defines preservation requirements, not a lossy compaction algorithm. Prompt/context caching also differs in minimum prefix size, implicit/explicit behavior, TTL, retention, and privacy semantics between providers.

A dedicated follow-up should define:

- context priorities;
- reserved output budget;
- stable/cache-friendly context hints;
- history/retrieval/tool-result budgeting;
- compaction provenance;
- privacy/retention policy.

It should not put provider cache IDs into the portable Assistant contract.

### Evaluation contracts

Dynamic routing and Skill Build will eventually need task-scoped eval data rather than price/provider heuristics. A follow-up may standardize portable evaluation observations for schema adherence, groundedness, tool correctness, latency, usage, and application-defined quality.

### MCP adapter

MCP or other tool protocols should adapt into the existing Core tool catalog. They must not bypass `ToolPolicy`, application authorization, consent, or data-sharing boundaries.

## Host boundary

Assistant owns contracts, portable semantics, invariants, and deterministic policy decisions. The consuming host/application owns:

- provider SDK/HTTP execution;
- raw secret storage/resolution;
- user/tenant authorization and tool side effects;
- persistence for memory, usage, traces, and learning revisions;
- RAG indexes, embedding storage, and infrastructure;
- orchestration loops, jobs, UI, billing, and product-specific behavior.

The Go `AdapterRegistry` is an implementation convenience for the current Go runtime, not a requirement that every conforming language expose the same registry shape.

## Validation before Ready for review

Run locally after updating the branch:

```sh
make format
make format-check
make conformance
make test
make test-race
make fuzz-smoke
make check
```

Useful manual performance baselines:

```sh
go test -run '^$' -bench 'BenchmarkResolveModelCandidates10000' -benchmem ./assistant
go test -run '^$' -bench 'BenchmarkValidate(Object64K|Schema)' -benchmem ./portablejson
```

Do not mark the PR Ready until the full local gate passes. Hosted CI/Security remains the final integration gate rather than the development loop.
