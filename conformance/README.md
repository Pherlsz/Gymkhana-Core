# Conformance suites

This directory contains language-neutral behavior vectors for the Core specification.

A language implementation that claims support for a specification version must pass every applicable vector for that version. The vectors define observable semantics; implementation details and idiomatic API shapes may differ by language.

## Layout

```text
conformance/
├── README.md
├── conformance_test.go   # shared Go runner for existing suites
├── assistant_test.go     # current Go Assistant conformance adapter
├── v0.1/
│   ├── normalize.json
│   └── temporal.json
├── v0.2/
│   ├── fingerprint.json
│   ├── normalize.json
│   └── temporal.json
└── v0.3/
    ├── assistant.json
    ├── fingerprint.json
    ├── normalize.json
    └── temporal.json
```

Each active specification-version directory carries the complete applicable suite set for an implementation claiming that version. Historical vectors remain immutable in meaning.

Future implementations should consume the same JSON vectors rather than copying test cases into language-specific fixtures.

## Vector model

Each case has:

- a stable human-readable `name`;
- an operation key such as `identity.br.cpf.canonicalize`, `fingerprint.sha256.framed_text`, or `assistant.message.validate`;
- JSON `input`, which may be a scalar, array, or structured object;
- exactly one expected JSON `output` or stable `error` code.

Structured inputs let conformance cover typed contracts without encoding multiple semantic arguments into an opaque string. Assistant vectors validate provider-neutral messages, tool definitions, finish reasons, usage, and capability sets.

## Rules

- vectors contain no product-specific persistence or infrastructure assumptions;
- errors are asserted by stable semantic code, never localized message text;
- sensitive original values must not appear in errors;
- jurisdiction-specific operations use an explicit jurisdiction in their operation key;
- provider-neutral Assistant vectors contain no provider SDK/model/transport fields;
- equivalent cases must keep the same meaning across all supported languages;
- implementations may use idiomatic APIs but adapters/runners must map them to the same operation semantics;
- changing an existing vector's semantics requires a specification version change.

The JSON shape is described by `schemas/conformance.schema.json`.
