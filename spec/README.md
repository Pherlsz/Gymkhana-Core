# Core Specification

Gymkhana Core is a language-neutral, product-neutral, provider-neutral, and jurisdiction-neutral foundation for reusable contracts, deterministic algorithms, validation, normalization, AI primitives, and other infrastructure-independent capabilities shared across applications and services.

The specification in this directory defines behavior independently from any implementation language. Go is currently the first implementation, but Go source code is not the architectural source of truth.

The active specification version is stored in `VERSION`; its compatibility rules are defined in [`VERSIONING.md`](VERSIONING.md).

## Principles

1. **Language-neutral** — contracts and behavior must be portable to Go, TypeScript/Node.js, Java, .NET/C#, Python, Rust, PHP, and future implementations.
2. **Product-neutral** — no application, including Gymkhana Database, owns or defines Core semantics.
3. **Provider-neutral** — provider SDKs, model choices, credentials, pricing, and transport-specific behavior remain outside Core.
4. **Jurisdiction-neutral** — Brazil-specific rules are one jurisdictional module among others; international standards and other jurisdictions are first-class.
5. **Deterministic where possible** — equivalent semantic input must produce equivalent canonical output across conforming implementations.
6. **Infrastructure-independent** — persistence, HTTP, queues, cloud services, UI frameworks, and deployment concerns remain consumer-owned.

## Specification layout

The current specification areas are:

- `text/` — language-agnostic text normalization semantics;
- `temporal/` — civil date and year-month semantics;
- `contact/` — email and explicitly contextualized telephone semantics;
- `identity/` — generic typed identifier concepts;
- `postal/` — explicitly jurisdiction-aware postal identifiers;
- `fingerprint/` — exact SHA-256 digest semantics and versioned domain-separated framing;
- `assistant/` — provider-neutral messages, multimodal references, tools, finish reasons, usage, and capability semantics;
- `jurisdictions/` — country/jurisdiction-specific semantics, beginning with Brazil because those capabilities already exist in the Go implementation;
- `standards/` — international/cross-jurisdiction standards when a reusable contract is justified;
- `errors/` — stable, non-localized error semantics.

Future areas may include `ocr/`, `matching/`, and `solver/`. Assistant streaming event semantics are intentionally deferred until multiple provider adapters have exercised the contract.

Jurisdiction-specific behavior must be namespaced explicitly, for example `identity.br.cpf` or `postal.br.cep`. A consumer must not be forced to assume a country, language, locale, currency, or timezone from another one of those concepts.

## Normative versus implementation material

`spec/` describes semantics. `conformance/` proves observable behavior. `schemas/` defines portable serialized shapes where machine-readable contracts are useful.

Language implementations may expose idiomatic APIs; they do not have to reproduce Go package/function naming. The cross-language contract is semantic behavior and stable serialization, not syntax.

## Conformance

Machine-readable behavior vectors live under `conformance/`. Every language implementation that claims support for a specification version must pass the applicable vectors.

Changing an existing vector's semantics requires an explicit specification-version decision. Language-local tests may add coverage but may not contradict the shared vectors.
