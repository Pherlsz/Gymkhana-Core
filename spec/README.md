# Core Specification

Gymkhana Core is a language-neutral, product-neutral, provider-neutral, and jurisdiction-neutral foundation for reusable contracts, deterministic algorithms, validation, normalization, AI primitives, and other infrastructure-independent capabilities shared across applications and services.

The specification in this directory defines behavior independently from any implementation language. Go is currently the first implementation, but Go source code is not the architectural source of truth.

## Principles

1. **Language-neutral** — contracts and behavior must be portable to Go, TypeScript/Node.js, Java, .NET/C#, Python, Rust, PHP, and future implementations.
2. **Product-neutral** — no application, including Gymkhana Database, owns or defines Core semantics.
3. **Provider-neutral** — provider SDKs, model choices, credentials, pricing, and transport-specific behavior remain outside Core.
4. **Jurisdiction-neutral** — Brazil-specific rules are one jurisdictional module among others; international standards and other jurisdictions are first-class.
5. **Deterministic where possible** — equivalent semantic input must produce equivalent canonical output across conforming implementations.
6. **Infrastructure-independent** — persistence, HTTP, queues, cloud services, UI frameworks, and deployment concerns remain consumer-owned.

## Specification layout

The initial specification areas are:

- `text/` — language-agnostic text normalization semantics;
- `temporal/` — civil date and year-month semantics;
- `identity/` — generic identifier concepts plus jurisdictional identifiers;
- `errors/` — stable, non-localized error semantics;
- future areas may include `fingerprint/`, `assistant/`, `ocr/`, `matching/`, `solver/`, and international standards.

Jurisdiction-specific behavior must be namespaced explicitly, for example `identity/br/cpf`. A consumer must not be forced to assume a country, language, locale, currency, or timezone from another one of those concepts.

## Conformance

Machine-readable behavior vectors live under `conformance/`. Every language implementation that claims support for a specification version must pass the applicable vectors.

The specification defines semantics. Each implementation may expose idiomatic APIs for its language as long as observable behavior remains conformant.
