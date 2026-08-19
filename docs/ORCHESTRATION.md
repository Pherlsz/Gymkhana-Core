# Gymkhana Core — Architecture and Orchestration

> **Foundation architecture:** 1  
> **Last synchronized:** 2026-08-19  
> **Repository responsibility:** language-neutral reusable contracts and infrastructure-independent algorithms  
> **Specification:** `spec/`  
> **Behavioral source of truth:** `spec/` + versioned `conformance/` vectors

## 1. Mission

Gymkhana Core is a private foundation for reusable semantics and algorithms shared across applications, services, runtimes, and jurisdictions.

The repository is not the core layer of Gymkhana Database. Gymkhana Database is one consumer among current and future services.

Go is the first implementation. Go source is not the architectural source of truth.

Core is designed so conforming implementations can exist in:

- Go;
- TypeScript/Node.js;
- Java;
- .NET/C#;
- Python;
- Rust;
- PHP;
- future languages when justified.

## 2. Architectural invariants

Core is:

1. **Language-neutral** — semantics are specified independently from implementation syntax and runtime conventions.
2. **Product-neutral** — no consumer application owns Core contracts or roadmap.
3. **Provider-neutral** — provider SDKs, concrete model APIs, credentials, pricing, and cloud-specific behavior stay in adapters owned by consumers.
4. **Jurisdiction-neutral** — no country is the implicit default; country-specific behavior is namespaced and isolated.
5. **Infrastructure-independent** — no persistence, HTTP, queue, UI, storage, or deployment dependency is required to execute Core logic.
6. **Deterministic where possible** — the same semantic input produces the same canonical result across conforming implementations.
7. **Explicitly bounded** — algorithms with nontrivial cost expose deterministic budgets/limits.
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
├── assistant/     current Go implementation
└── future language/package implementations as justified
```

The current Go module path remains stable while Go is the only implementation:

```text
github.com/Pherlsz/Gymkhana-Core
```

Moving Go under an implementation subdirectory is deferred until there is enough multi-language value to justify a breaking import-path migration.

## 5. Specification version versus implementation version

`spec/VERSION` versions the language-neutral semantics.

Repository `VERSION` currently versions the Go implementation/release.

These versions are intentionally different concepts. Future language packages may have independent implementation SemVer while declaring which Core specification version they support.

Example future model:

```text
Core Spec 2.1
├── Go          0.9.2  -> spec 2.1
├── TypeScript  0.7.0  -> spec 2.1
├── Java        0.4.1  -> spec 2.0
└── Python      0.3.0  -> spec 2.1
```

Do not force synchronized package releases once multiple language ecosystems exist.

## 6. Conformance model

Shared vectors under `conformance/` define observable behavior independently from implementation language.

Rules:

- every implementation claiming a specification version passes every applicable vector for that version;
- language-local tests may add coverage but may not contradict shared vectors;
- errors are compared by stable semantic code, not localized text;
- conformance data contains no product persistence or infrastructure assumptions;
- changing existing observable behavior requires a specification version decision and updated vectors;
- algorithms may have additional property/fuzz/benchmark requirements where fixed vectors are insufficient.

The current Go implementation executes the shared suites as part of `go test ./...` and through `make conformance`.

## 7. Global versus jurisdiction-specific behavior

Global primitives include concepts such as:

- Unicode/text canonicalization;
- civil temporal values;
- generic identifier semantics;
- canonical encoding/fingerprints;
- provider-neutral Assistant contracts;
- OCR evidence contracts;
- matching primitives;
- generic bounded solver algorithms.

Jurisdictional behavior uses explicit namespaces, for example:

```text
identity.br.cpf
identity.br.cnpj
postal.br.cep
```

Future modules may cover `us`, `pt`, `gb`, `ca`, and others without making any one jurisdiction the default.

Country, jurisdiction, language, locale, currency, and timezone are distinct concepts and are never silently treated as synonyms.

## 8. International standards

Cross-jurisdiction standards belong under `spec/standards/` when they form a stable reusable contract. Potential examples include:

- ISO 3166;
- ISO 4217;
- BCP 47;
- E.164;
- IBAN;
- BIC/SWIFT.

Do not add large standards datasets speculatively. A standard enters Core when behavior, maintenance/versioning, and consumer value are understood.

## 9. Data and heuristic boundaries

Core algorithms must not become a dumping ground for one application's dirty-data vocabulary.

Distinguish:

- normative reusable rules;
- reusable versioned datasets;
- consumer-owned alias catalogs and migration cleanup heuristics.

For example, a generic alias matcher may belong in Core while mappings such as an application-specific shorthand for one city should normally remain consumer data unless a broader stable contract is demonstrated.

## 10. Current Foundation areas

### Text

Language-neutral display/search text semantics and Unicode-aware canonicalization.

### Temporal

Timezone-free `CivilDate` and `YearMonth` semantics.

### Identity

Generic typed identifier concepts with explicit jurisdiction keys. The current Go implementation already includes Brazilian CPF/CNPJ and document behavior; the specification treats Brazil as one jurisdiction module.

### Fingerprint

Exact SHA-256 identity plus versioned domain-separated framing for deterministic multipart fingerprints. Fingerprinting does not perform implicit application canonicalization.

### Assistant

Provider-neutral messages, ordered content parts, external media references, tool definitions/calls/results, finish reasons, usage counters, capability names, portable schemas, and stable validation errors. Provider SDKs and transport remain outside Core.

Streaming is currently a capability only; a shared stream-event/delta contract is deliberately deferred until multiple provider adapters have exercised the semantics.

### Errors

Stable non-localized error codes without sensitive original-value leakage.

## 11. Planned capability order

The Foundation roadmap is currently:

1. Foundation specification/conformance infrastructure — **implemented**;
2. global/jurisdiction separation for existing normalization — **established architecturally; incremental cleanup continues**;
3. canonical `fingerprint` package/specification — **implemented in Core Spec 0.2 / Go 0.4.0**;
4. provider-neutral `assistant` contracts — **implemented by Core Spec 0.3 / Go 0.5.0 workstream**;
5. provider-neutral OCR/multimodal extraction contracts — **next planned capability**;
6. generic matching/evidence primitives;
7. generic solver/constraint/composition algorithms;
8. first non-Go implementation, preferably TypeScript as the first portability stress test;
9. Java, .NET/C#, Python, Rust, and PHP ports as real consumers justify them;
10. package ecosystem publication and implementation/spec compatibility matrix.

This ordering is guidance, not a coupling to any product milestone.

## 12. Assistant boundary

The Assistant specification owns neutral concepts such as:

- messages and semantic roles;
- typed content parts;
- external image/file references;
- tool definitions, calls, and results;
- capabilities;
- finish reasons;
- usage structures;
- stable validation errors.

It does not own:

- OpenAI, Gemini, Anthropic, or other SDK types;
- API keys/secrets;
- concrete model selection;
- provider pricing;
- application prompts;
- persistence/session authorization;
- provider HTTP/retry behavior;
- autonomous unbounded agent loops.

Streaming event semantics remain deferred. Provider adapters may expose native streaming while Core defines only the `streaming` capability until a portable event contract is proven.

## 13. OCR/multimodal boundary

A future OCR/multimodal specification may own provider-neutral suggestions, evidence, bounded regions, review states, and structured extraction results.

Provider invocation, file storage, document rendering, permissions, persistence, and mutation/application remain consumer-owned.

## 14. Matching boundary

A future matching package may own deterministic evidence representation, comparisons, contradiction/support semantics, and ranking/assessment primitives.

Product entities such as Profile, customer, document row, merge queue, or database transaction are not Core concepts.

## 15. Solver boundary

A future solver package may own candidates, slots/roles, generic constraints, deterministic enumeration/pruning, budgets, partial/incomplete outcomes, and evidence.

It must not know Gymkhana-specific persistence, QueryPlan SQL compilation, Profiles, bills, permissions, or workers.

## 16. Implementation design rules

For every public capability:

- prefer standard-library-first implementations;
- use pure functions and immutable values where practical;
- expose interfaces only at real substitution boundaries;
- define stable structured errors;
- document canonical versus display representations;
- preserve leading zeroes where identifiers require them;
- fuzz malformed/untrusted inputs where appropriate;
- benchmark nontrivial algorithms;
- define explicit complexity/budget behavior;
- avoid infrastructure types in public contracts;
- add/update conformance vectors for portable observable semantics.

## 17. What does not belong in Core

- PostgreSQL repositories, pgx, sqlc, migrations, indexes, transactions;
- HTTP handlers/status codes, cookies, sessions, CSRF, resource ownership;
- permissions/roles and application audit storage;
- River, R2, signed URLs, Cloud Run, Vercel, deployment topology;
- React, routes, browser state, product forms, Data Grid components;
- provider SDK types, credentials, concrete prompts/model choices/pricing;
- complete application domain/persistence models;
- arbitrary SQL/code/expression execution;
- consumer-specific migration scripts or one-off cleanup catalogs.

## 18. Multi-language implementation policy

Do not write seven ports preemptively.

A new language implementation starts when there is a real consumer or portability need. It must:

- declare the supported Core spec version;
- consume the shared conformance vectors;
- use idiomatic language APIs rather than mechanically mirroring Go syntax;
- publish through the normal ecosystem package manager when distribution is needed;
- keep implementation-specific dependencies out of the language-neutral spec.

Contract/model generation from schemas is encouraged where appropriate. Nontrivial algorithms should normally be implemented idiomatically per language and proven by conformance/property tests rather than generated into unreadable source.

## 19. Go implementation quality gates

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

Core tests do not spin up PostgreSQL, HTTP servers, provider APIs, or browser environments.

## 20. Pull request requirements

Core PRs describe:

- objective and reusable scope;
- originating consumer requirement when applicable;
- specification impact;
- public API impact per implementation;
- determinism and budgets;
- conformance/test/fuzz/benchmark coverage;
- security/privacy implications;
- consumer migration notes for breaking behavior;
- release/version impact.

A consumer link is context, not architectural authority.

## 21. Definition of Done for a portable Core capability

A capability is complete when:

- responsibility is coherent and reusable;
- normative semantics are documented;
- observable portable behavior has conformance vectors when applicable;
- current implementation tests pass;
- fuzz/property tests exist for untrusted/structural inputs when appropriate;
- nontrivial algorithms have benchmarks and explicit budgets;
- errors are stable and non-localized;
- no infrastructure/provider/product leakage exists;
- compatibility/version impact is documented;
- a consumer can use it without importing another product repository.
