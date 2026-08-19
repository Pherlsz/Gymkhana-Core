# Portable JSON and structured schema

This document is normative for Core Spec `0.3`.

Portable JSON is a generic Core primitive. It is not owned by Assistant: Assistant tool arguments and structured responses are consumers, and future OCR/extraction, matching, solver, workflow, or other Core domains may use the same contract without importing Assistant semantics.

## Strict JSON boundary

Portable validation accepts UTF-8 JSON and rejects:

- invalid UTF-8;
- duplicate object keys at any depth;
- unpaired UTF-16 surrogate escapes;
- trailing JSON values;
- payloads above the portable byte bound;
- excessive recursive depth or node count.

Object-specific boundaries additionally require the root JSON value to be an object.

Duplicate keys are invalid because runtimes and libraries differ on first-value, last-value, error, or duplicate-preserving behavior. A conforming implementation must reject the ambiguity rather than normalize it silently.

## `portable_json_schema/v1`

The portable schema profile intentionally supports a conservative subset instead of claiming full JSON Schema compatibility.

Supported base types:

- `object`;
- `array`;
- `string`;
- `integer`;
- `number`;
- `boolean`;
- `null`.

A nullable value may contain exactly one non-null base type plus `null`.

Supported keywords:

- `type`;
- `properties`;
- `required`;
- `additionalProperties`;
- `items`;
- `enum`;
- `description`;
- `title`;
- `minimum`;
- `maximum`;
- `minItems`;
- `maxItems`.

Unknown keywords and keywords that do not apply to the declared base type are invalid rather than ignored.

Typed object schemas are strict:

- `additionalProperties` is explicitly `false`;
- every declared property appears exactly once in `required`;
- property count and nesting are bounded.

`{}` is the explicit compatibility form for an unconstrained JSON **object**. It is not an unconstrained arbitrary JSON value.

Arrays require an `items` schema and may use bounded `minItems`/`maxItems`. Numeric bounds and integer values must stay inside the portable safe range where the profile requires exact cross-runtime integer behavior.

## Instance validation

A schema is useful only if the produced value is checked against it. Conforming implementations therefore expose equivalent semantics for validating an instance against `portable_json_schema/v1`.

Instance validation must preserve strict JSON lexical checks before semantic validation. In particular, a duplicate-key object cannot become valid merely because the host language decoder would overwrite one duplicate.

## Domain-specific restrictions

A consuming domain may impose a stricter root or smaller resource limits without changing the generic schema semantics.

For example, Assistant `ToolCall.arguments` is always an object. Assistant therefore requires `ToolDefinition.input_schema` to be either `{}` or a non-nullable root `object` schema, while a structured generation response may use another portable root type.

## Out of scope

Core Spec `0.3` does not define:

- full JSON Schema Draft 2020-12 semantics;
- provider-specific structured-output extensions;
- JSON canonicalization for fingerprinting;
- a universal object serializer;
- persistence or transport of structured values.

Provider or product extensions remain outside the portable contract. A consumer/adaptor must reject a translation it cannot preserve rather than silently weakening the requested schema.
