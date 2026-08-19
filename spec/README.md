# Core Specification

Gymkhana Core is a language-neutral, product-neutral, provider-neutral, and jurisdiction-neutral foundation for reusable contracts, deterministic algorithms, validation, normalization, AI primitives, and other infrastructure-independent capabilities shared across applications and services.

The specification in this directory defines behavior independently from any implementation language. Go is currently the first implementation, but Go source code is not the architectural source of truth.

The active specification version is stored in `VERSION`; compatibility rules are defined in [`VERSIONING.md`](VERSIONING.md).

## Principles

1. **Language-neutral** — contracts/behavior must be portable to Go, TypeScript/Node.js, Java, .NET/C#, Python, Rust, PHP, and future implementations.
2. **Product-neutral** — no application, including Gymkhana Database, owns Core semantics or roadmap.
3. **Provider-neutral** — portable AI contracts may define providers/models/capabilities/routing semantically, but provider SDK/HTTP objects, raw secrets, volatile pricing, and deployment remain outside the language-neutral boundary.
4. **Jurisdiction-neutral** — Brazil-specific rules are one jurisdictional module among others; international standards and other jurisdictions are first-class.
5. **Deterministic where possible** — equivalent semantic input must produce equivalent canonical output across conforming implementations.
6. **Infrastructure-independent** — persistence, HTTP, queues, cloud services, UI frameworks, secret stores, and deployment concerns remain consumer-owned.
7. **Bounded** — untrusted collections, recursive structures, retry/fallback, and expensive algorithms expose explicit limits/defaults rather than implicit unbounded work.
8. **Authority-aware** — AI-produced/retrieved/tool/document content remains data unless a host deliberately grants higher authority; learning cannot silently self-promote into system/developer instructions.

## Specification layout

Current specification areas:

- `text/` — language-agnostic text normalization semantics;
- `temporal/` — civil date and year-month semantics;
- `contact/` — email and explicitly contextualized telephone semantics;
- `identity/` — generic typed identifier concepts;
- `postal/` — explicitly jurisdiction-aware postal identifiers;
- `fingerprint/` — exact SHA-256 digest semantics and versioned domain-separated framing;
- `json/` — strict portable JSON plus `portable_json_schema/v1` schema/instance semantics reusable across Core domains;
- `assistant/` — provider-neutral modular/multimodal Assistant messages, tools, model/credential routing, RAG, memory, learning, usage/quota, and observability semantics; Assistant consumes the generic JSON contract for structured data;
- `jurisdictions/` — country/jurisdiction-specific semantics, beginning with Brazil because those capabilities already exist in the Go implementation;
- `standards/` — international/cross-jurisdiction standards when a reusable contract is justified;
- `errors/` — stable, non-localized error semantics.

Assistant Spec `0.3` companion documents cover configuration/model/BYOK routing, Assistant-specific structured-output restrictions, tool safety, RAG, task-scoped Skill Build, and content-free quota/usage/tracing. Generic `portable_json_schema/v1` semantics live under `json/`. Provider-neutral streaming event semantics, concrete provider adapters, Context Planner/cache hints/evals, and normative OCR extraction remain later workstreams.

Future top-level areas may include `ocr/`, `matching/`, and `solver/`.

Jurisdiction-specific behavior must be namespaced explicitly, for example `identity.br.cpf` or `postal.br.cep`. A consumer must not be forced to assume a country, language, locale, currency, or timezone from another one of those concepts.

## Normative versus implementation material

`spec/` describes semantics. `conformance/` proves observable behavior. `schemas/` defines portable serialized shapes where machine-readable contracts are useful.

Language implementations may expose idiomatic APIs; they do not have to reproduce Go package/function naming. The cross-language contract is semantic behavior and stable serialization, not syntax.

## Conformance

Machine-readable behavior vectors live under `conformance/`. Every language implementation that claims support for a specification version must pass all applicable vectors.

Changing an existing vector's semantics requires an explicit specification-version decision. Language-local tests, fuzzing, benchmarks, and security checks may add coverage but may not contradict shared vectors.
