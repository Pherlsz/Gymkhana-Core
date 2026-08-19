# Conformance suites

This directory contains language-neutral behavior vectors for the Core specification.

A language implementation that claims support for a specification version must pass every applicable vector for that version. The vectors define observable semantics; implementation details and idiomatic API shapes may differ by language.

## Layout

```text
conformance/
├── README.md
├── conformance_test.go   # current Go conformance runner
└── v0.1/
    ├── normalize.json
    └── temporal.json
```

Future implementations should consume the same JSON vectors rather than copying test cases into language-specific fixtures.

## Rules

- vectors contain no product-specific persistence or infrastructure assumptions;
- errors are asserted by stable semantic code, never localized message text;
- sensitive original values must not appear in errors;
- jurisdiction-specific operations use an explicit jurisdiction in their operation key;
- equivalent cases must keep the same meaning across all supported languages;
- changing an existing vector's semantics requires a specification version change.

The JSON shape is described by `schemas/conformance.schema.json`.
