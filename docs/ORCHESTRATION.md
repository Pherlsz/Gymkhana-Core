# Gymkhana Core — Architecture and Orchestration

> **Foundation architecture:** 1  
> **Last synchronized:** 2026-08-19  
> **Repository responsibility:** language-neutral reusable contracts and infrastructure-independent algorithms  
> **Specification:** `spec/`  
> **Behavioral source of truth:** `spec/` + versioned `conformance/` vectors

## 1. Mission

Gymkhana Core is a private foundation for reusable semantics and algorithms shared across applications, services, runtimes, providers, languages, and jurisdictions.

The repository is not the core layer of Gymkhana Database. Gymkhana Database is one consumer among current and future services.

Go is the first implementation. Go source is not the architectural source of truth.

Core is designed so conforming implementations can exist in Go, TypeScript/Node.js, Java, .NET/C#, Python, Rust, PHP, and future languages when justified.

## 2. Architectural invariants

Core is:

1. **Language-neutral** — semantics are specified independently from implementation syntax and runtime conventions.
2. **Product-neutral** — no consumer application owns Core contracts or roadmap.
3. **Provider-neutral** — portable provider/model/credential-reference semantics are allowed, while provider SDK types, raw secrets, provider HTTP payloads, and volatile pricing/model tables remain outside portable contracts.
4. **Jurisdiction-neutral** — no country is the implicit default; country-specific behavior is namespaced and isolated.
5. **Infrastructure-independent** — no persistence, HTTP, queue, UI, storage, or deployment dependency is required to execute Core logic.
6. **Deterministic where possible** — the same semantic input produces the same canonical result across conforming implementations.
7. **Explicitly bounded** — algorithms or agent-like workflows with nontrivial cost expose deterministic budgets/limits; zero-value budgets mean application defaults, never infinity.
8. **Small at the public boundary** — do not create catch-all `utils`, `helpers`, `common`, `shared`, or `core` packages.

## 3. Sources of truth

Precedence for Core behavior:

1. an explicit current architectural decision recorded in Core;
2. the versioned language-neutral specification under `spec/`;
3. versioned conformance vectors under `conformance/`;
4. machine-readable schemas under `schemas/`;
5. language implementation code and language-local tests;
6. consumer integration behavior.

A consumer issue can motivate a Core change but cannot redefine Core semantics implicitly.

## 4. Repository layers

```text
Gymkhana-Core/
├── spec/          language-neutral normative semantics
├── conformance/   shared observable behavior vectors
├── schemas/       portable serialized contracts
├── normalize/     current Go implementation
├── civiltime/     current Go implementation
├── fingerprint/   current Go implementation
├── portablejson/  current generic structured-data implementation
├── assistant/     current provider-neutral Assistant implementation
├── ocr/           current provider-neutral extraction implementation
└── future language/package implementations as justified
```

The current Go module path remains stable while Go is the only implementation:

```text
github.com/Pherlsz/Gymkhana-Core
```

Moving Go under an implementation subdirectory is deferred until there is enough multi-language value to justify a breaking import-path migration.

## 5. Specification version versus implementation version

`spec/VERSION` versions the language-neutral semantics. Repository `VERSION` currently versions the Go implementation/release.

These versions are intentionally different concepts. Future language packages may have independent implementation SemVer while declaring which Core specification version they support.

Do not force synchronized package releases once multiple language ecosystems exist.

## 6. Conformance model

Shared vectors under `conformance/` define observable behavior independently from implementation language.

Rules:

- every implementation claiming a specification version passes every applicable vector for that version;
- language-local tests may add coverage but may not contradict shared vectors;
- errors are compared by stable semantic code, not localized text;
- conformance data contains no product persistence or infrastructure assumptions;
- model-routing vectors use synthetic model catalogs rather than freezing volatile provider availability/pricing;
- BYOK vectors contain opaque credential references, never secrets;
- changing existing observable behavior requires a specification version decision and updated vectors;
- algorithms may have additional property/fuzz/benchmark requirements where fixed vectors are insufficient.

The current Go implementation executes the shared suites as part of `go test ./...` and through `make conformance`.

## 7. Global versus jurisdiction-specific behavior

Global primitives include Unicode/text canonicalization, civil temporal values, generic identifier semantics, canonical encoding/fingerprints, provider-neutral Assistant contracts, OCR evidence contracts, matching primitives, and generic bounded solver algorithms.

Jurisdictional behavior uses explicit namespaces, for example:

```text
identity.br.cpf
identity.br.cnpj
postal.br.cep
```

Future modules may cover `us`, `pt`, `gb`, `ca`, and others without making any one jurisdiction the default.

Country, jurisdiction, language, locale, currency, and timezone are distinct concepts and are never silently treated as synonyms.

## 8. International standards

Cross-jurisdiction standards belong under `spec/standards/` when they form a stable reusable contract. Potential examples include ISO 3166, ISO 4217, BCP 47, E.164, IBAN, and BIC/SWIFT.

Do not add large standards datasets speculatively. A standard enters Core when behavior, maintenance/versioning, and consumer value are understood.

## 9. Data and heuristic boundaries

Core algorithms must not become a dumping ground for one application's dirty-data vocabulary.

Distinguish normative reusable rules, reusable versioned datasets, and consumer-owned alias catalogs/migration cleanup heuristics.

A generic alias matcher may belong in Core while mappings such as an application-specific shorthand for one city should normally remain consumer data unless a broader stable contract is demonstrated.

## 10. Current Foundation areas

### Text

Language-neutral display/search text semantics and Unicode-aware canonicalization.

### Temporal

Timezone-free `CivilDate` and `YearMonth` semantics.

### Identity

Generic typed identifier concepts with explicit jurisdiction keys. The current Go implementation already includes Brazilian CPF/CNPJ and document behavior; the specification treats Brazil as one jurisdiction module.

### Fingerprint

Exact SHA-256 identity plus versioned domain-separated framing for deterministic multipart fingerprints. Fingerprinting does not perform implicit application canonicalization.

### Portable JSON

Strict portable JSON and the deliberately bounded `portable_json_schema/v1` profile provide reusable structured-data semantics for Assistant, OCR, and future Core domains without making JSON behavior belong to any one AI package.

### Assistant

The Assistant foundation originated in Core Spec `0.3` / Go `0.5.0`. Active Spec `0.4` keeps the reusable provider-neutral boundary:

- semantic roles and ordered text/image/audio/video/file/tool content;
- built-in versioned skills such as `token_economy/v1`;
- multiple reusable `AssistantDefinition` profiles;
- default and optional modules with per-run enable/disable overrides;
- tool allowlists and bounded execution budgets;
- live provider-supplied model catalogs;
- manual, ordered-fallback, and dynamic model routing;
- free/paid/local/unknown access metadata without frozen Core price tables;
- automatic model-capability filtering from active modules;
- managed/BYOK credential policies using opaque per-provider references rather than raw secrets;
- lexical/vector/hybrid RAG, query transformation, candidate/context limits, optional reranking, grounding, citations, and retrieval evidence;
- quota/usage attribution and content-free run traces;
- Go provider-adapter/registry substitution boundaries with extensible provider IDs.

Product-owned persistent memory and learning/promotion workflows are intentionally host concerns in Spec `0.4`. Hosts may project selected prior context through messages or RAG without requiring Core to own identity, persistence, retention, evaluation, or self-modification semantics.

Concrete provider network adapters may use SDKs/HTTP internally and may be delivered independently; their SDK types must never leak into portable Core contracts.

Streaming remains a capability only; a shared stream-event/delta contract is deliberately deferred until multiple concrete provider adapters exercise the semantics.

### OCR / extraction

Core Spec `0.4` / Go `0.6.0` defines provider-neutral schema-guided and discovery extraction across text, image, document, audio, and video sources. The boundary owns logical sources, evidence, observations, field candidates, confidence, ambiguity, validation/review state, and bounded structured results.

Discovery correctness is encoded directly as extraction invariants. Core does not expose a separate OCR discovery skill or provider prompt profile. OCR depends on generic `portablejson`, not Assistant; an AI-backed host may compose Assistant infrastructure without creating a normative package dependency.

Provider invocation, uploads/storage, document rendering, permissions, persistence, review UI, billing, and product-specific document models remain consumer/adapter concerns.

### Errors

Stable non-localized error codes without sensitive original-value leakage.

## 11. Planned capability order

The Foundation roadmap is currently:

1. Foundation specification/conformance infrastructure — **implemented**;
2. global/jurisdiction separation for existing normalization — **established architecturally; incremental cleanup continues**;
3. canonical `fingerprint` package/specification — **implemented in Core Spec 0.2 / Go 0.4.0**;
4. modular provider-neutral `assistant` contracts, model routing, BYOK references, RAG policy, and adapter boundary — **implemented in Core Spec 0.3 / Go 0.5.0; narrowed in Spec 0.4**;
5. provider-neutral OCR/multimodal extraction contracts — **implemented in Core Spec 0.4 / Go 0.6.0; discovery invariants are direct contract semantics**;
6. concrete provider-adapter validation against several materially different provider families/gateways/local runtimes — **next integration stress test; may ship independently from Core semantics**;
7. generic matching/evidence primitives;
8. generic solver/constraint/composition algorithms;
9. first non-Go implementation, preferably TypeScript as the first portability stress test;
10. Java, .NET/C#, Python, Rust, and PHP ports as real consumers justify them;
11. package ecosystem publication and implementation/spec compatibility matrix.

Concrete adapter work may proceed without destabilizing the portable Assistant or OCR contracts.

## 12. Assistant boundary

The Assistant specification may own:

- messages, semantic roles, and multimodal content references;
- tool definitions/calls/results and structured-output contracts;
- reusable Assistant definitions/modules/skills/budgets;
- model/provider references, live catalog metadata, and deterministic routing policy;
- managed/BYOK credential modes and opaque credential references;
- RAG/retrieval/evidence/citation policy;
- provider adapter substitution boundaries;
- capabilities, finish reasons, usage/quota, content-free traces, and stable errors.

It does not own:

- OpenAI, Gemini, Anthropic, or other provider SDK types in portable APIs;
- raw API keys/secrets or secret storage;
- provider HTTP/retry payload formats;
- volatile provider pricing/free-tier/model availability tables;
- consumer-specific prompt policy embedded as Core behavior;
- product persistence/session authorization;
- persistent user/tenant memory or memory storage policy;
- task signatures, learning proposals, evaluation evidence, or promotion workflows;
- autonomous unbounded agent loops.

Provider/model references are not provider coupling: they are opaque portable identifiers resolved by adapters. Concrete model availability and access tier are runtime data.

## 13. RAG boundary

Core owns provider-neutral retrieval policy and evidence semantics, but does not own a vector database, search product, ACL store, ingestion pipeline, or one universal chunking strategy.

Authorization filtering must occur before retrieved evidence enters model context. Retrieved documents are untrusted data and cannot promote themselves into system/developer instruction authority.

Retrieval/query/reranking changes should be evaluated against representative queries for relevance/recall, groundedness, citation correctness, latency, and token/cost usage rather than assumed universally optimal.

## 14. OCR/multimodal extraction boundary

Core Spec `0.4` owns provider-neutral extraction requests/results, candidate field identification, evidence, bounded regions/text spans/time ranges, confidence/ambiguity, review/validation states, and schema-guided structured extraction results.

When a target schema is already known, schema-guided extraction uses that exact schema and does not silently switch to generic discovery. When fields are unknown, discovery must preserve raw observations when available, require evidence, preserve ambiguity, avoid inventing missing values, avoid assuming jurisdiction, and normalize only through an explicit canonicalizer.

Those discovery rules are direct OCR contract invariants, not a built-in Core skill. Provider-specific prompt text and request shaping remain adapter concerns.

Provider invocation, file storage, document rendering, permissions, persistence, billing, review workflow, and mutation/application remain consumer-owned.

## 15. Matching boundary

A future matching package may own deterministic evidence representation, comparisons, contradiction/support semantics, and ranking/assessment primitives.

Product entities such as Profile, customer, document row, merge queue, or database transaction are not Core concepts.

## 16. Solver boundary

A future solver package may own candidates, slots/roles, generic constraints, deterministic enumeration/pruning, budgets, partial/incomplete outcomes, and evidence.

It must not know Gymkhana-specific persistence, QueryPlan SQL compilation, Profiles, bills, permissions, or workers.

## 17. Implementation design rules

For every public capability:

- prefer standard-library-first implementations;
- use pure functions and immutable values where practical;
- expose interfaces only at real substitution boundaries;
- define stable structured errors;
- document canonical versus display representations;
- preserve leading zeroes where identifiers require them;
- fuzz malformed/untrusted structural inputs where appropriate;
- benchmark nontrivial algorithms;
- define explicit complexity/budget behavior;
- avoid infrastructure types in portable public contracts;
- add/update conformance vectors for portable observable semantics.

## 18. What does not belong in Core

- PostgreSQL repositories, pgx, sqlc, migrations, indexes, transactions;
- HTTP handlers/status codes, cookies, sessions, CSRF, resource ownership;
- permissions/roles and application audit storage;
- River, R2, signed URLs, Cloud Run, Vercel, deployment topology;
- React, routes, browser state, product forms, Data Grid components;
- provider SDK types/HTTP payloads/raw credentials/volatile pricing tables in portable contracts;
- product-owned persistent memory, learning/evaluation state, and promotion workflows;
- complete application domain/persistence models;
- arbitrary SQL/code/expression execution;
- autonomous unbounded agent loops;
- consumer-specific migration scripts or one-off cleanup catalogs.

## 19. Multi-language implementation policy

Do not write seven ports preemptively.

A new language implementation starts when there is a real consumer or portability need. It must declare the supported Core spec version, consume shared conformance vectors, use idiomatic APIs, publish through the normal ecosystem package manager when distribution is needed, and keep implementation/provider-specific dependencies out of language-neutral contracts.

Contract/model generation from schemas is encouraged where appropriate. Nontrivial algorithms should normally be implemented idiomatically per language and proven by conformance/property tests rather than generated into unreadable source.

## 20. Go implementation quality gates

Current Go development uses:

```text
make setup
make format
make format-check
make vet
make lint
make test
make conformance
make test-race
make fuzz-smoke
make vuln
make security
make check
```

`make check` is the required local gate before marking a public-contract PR ready for review.

Core tests do not spin up PostgreSQL, provider APIs, or browser environments.

## 21. Pull request requirements

Core PRs describe objective/reusable scope, specification impact, public API impact, determinism/budgets, conformance/test/fuzz coverage, security/privacy implications, migration notes for breaking behavior, and release/version impact.

A consumer link is context, not architectural authority.

## 22. Definition of Done for a portable Core capability

A capability is complete when responsibility is coherent/reusable, normative semantics are documented, portable behavior has conformance vectors when applicable, current implementation tests pass, fuzz/property tests cover untrusted structural inputs where appropriate, nontrivial algorithms have benchmarks/budgets, errors are stable/non-localized, no infrastructure/provider/product leakage violates the boundary, compatibility/version impact is documented, and a consumer can use it without importing another product repository.
