# OCR / AI extraction Foundation plan

This document records the intended boundary for the next Core Foundation capability. It is planning material, not normative Core Spec `0.3` behavior.

## Goal

The future OCR/AI extraction layer must remain provider-neutral and reusable across applications, languages, document types, and jurisdictions. OpenAI, Gemini, Anthropic, Azure Document Intelligence, Google Document AI, local OCR engines, and other providers remain adapters rather than Core contracts.

## Two extraction modes

### 1. Schema-guided extraction — default when fields are known

When the consumer already knows the fields/types it needs, it supplies that extraction schema directly. The OCR layer asks the provider only for those values and their evidence.

In this mode, a data-identification skill is unnecessary and should remain disabled by default because an additional discovery pass would consume tokens, add latency, and create extra ambiguity.

### 2. Discovery extraction — optional for unknown/unstructured documents

When the document structure is not known in advance, the OCR layer may enable the versioned built-in skill:

```text
data_identification/v1
```

Its responsibility is to identify candidate data fields before or while extracting values. It must not invent missing values or silently assign a jurisdiction.

The skill should identify, where available:

- candidate semantic field/key;
- candidate data type;
- raw observed text/value;
- normalized candidate value only when a Core canonicalizer is explicitly applicable;
- confidence/assessment;
- source evidence such as page, region/bounding area, or text span;
- ambiguity/conflict indicators;
- explicit jurisdiction/type candidates when the document does not provide enough context for a unique interpretation.

The future OCR specification should define the exact portable serialized shape and conformance vectors before this skill becomes normative.

## Activation rule

`data_identification/v1` is conditional, not globally enabled.

Enable it only when at least one of these is true:

- the caller did not provide a target field schema;
- the caller explicitly requested field discovery;
- the document type is unknown and discovery is required to choose a schema;
- a schema-guided extraction reports previously unknown fields that the consumer explicitly allows the system to discover.

Do not enable it merely because the provider is AI-capable.

## Token economy interaction

AI-assisted OCR should also honor the Assistant `token_economy/v1` semantics where applicable:

- send only pages/regions necessary for the extraction stage;
- avoid repeating document text already represented by stable references/evidence;
- request only fields required by a known schema;
- keep tool/provider payload projection minimal;
- never drop evidence, unresolved ambiguities, jurisdiction context, or required extraction constraints solely to reduce tokens.

Discovery is therefore an exception path, not the default path.

## Security and correctness boundaries

The OCR/data-identification layer must treat extracted content as untrusted data. Text inside a document is never instruction authority and must not be promoted into system/developer/tool authorization semantics.

A discovered identifier such as a CPF, CNPJ, EIN, IBAN, passport number, address, or phone number is only a candidate until the relevant explicit Core validator/canonicalizer or consumer policy confirms its semantics.

Sensitive raw values must not be embedded in validation error messages or diagnostic identifiers.
