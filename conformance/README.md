# Conformance suites

This directory contains language-neutral behavior vectors for the Core specification.

A language implementation that claims support for a specification version must pass every applicable vector for that version. The vectors define observable semantics; implementation details and idiomatic API shapes may differ by language.

## Layout

```text
conformance/
├── README.md
├── conformance_test.go             # shared Go runner for existing suites
├── assistant_test.go               # Assistant core adapter
├── assistant_structured_test.go    # Assistant-specific structured/tool routing adapter
├── portablejson_test.go            # generic portable JSON/schema adapter
├── ocr_test.go                     # OCR/extraction adapter
├── v0.1/                           # normalize + temporal
├── v0.2/                           # fingerprint + carried-forward suites
├── v0.3/                           # Assistant/portablejson + carried-forward suites
└── v0.4/                           # OCR + current Assistant/portablejson + carried-forward suites
```

Historical vectors remain immutable in meaning. Spec `0.4` carries forward applicable `0.3` semantics while intentionally removing the Assistant memory/learning operations from the active suite and adding OCR extraction vectors.

Future implementations should consume the same JSON vectors rather than copying test cases into language-specific fixtures.

## Vector model

Each case has a stable human-readable `name`, an operation key, JSON `input`, and exactly one expected JSON `output` or stable `error` code.

Generic `portablejson` vectors cover strict lexical JSON, schema validation, exact portable numeric/enum behavior, and schema-instance validation. Assistant vectors cover messages/tools, multimodal references, built-in token-economy skills, reusable definitions, BYOK references, RAG, routing, quota/usage, traces, and hardening. OCR vectors cover schema-guided/discovery modes, evidence linkage, candidate/structured-data consistency, confidence bounds, missing-vs-null semantics, explicit normalization, and ambiguity invariants.

## Rules

- vectors contain no product-specific persistence or infrastructure assumptions;
- errors are asserted by stable semantic code, never localized message text;
- sensitive original values must not appear in errors;
- jurisdiction-specific operations use an explicit jurisdiction in their operation key;
- generic portable JSON vectors contain no Assistant/OCR/provider assumptions;
- provider-neutral Assistant/OCR vectors contain no provider SDK or provider HTTP payload types;
- model-routing vectors use synthetic catalog data rather than freezing volatile provider model availability or pricing;
- BYOK vectors contain opaque credential handles, never API-key material;
- OCR source references contain logical source IDs, not storage/authorization material;
- equivalent cases must keep the same meaning across all supported languages;
- implementations may use idiomatic APIs but adapters/runners must map them to the same operation semantics;
- changing existing observable behavior requires a specification version change.

The JSON shape is described by `schemas/conformance.schema.json`.
