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

Strict lexical JSON validation by itself does not imply that every numeric value is suitable for every Core domain. Schema-instance validation applies the additional portable numeric rules below.

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

Arrays require an `items` schema and may use bounded `minItems`/`maxItems`.

### Numeric semantics

Portable schema numeric comparison is mathematical rather than binary floating-point comparison. Decimal/exponent JSON numbers are compared exactly as rational values for `minimum`, `maximum`, and numeric `enum` membership. Therefore values such as `0.1` and `0.10000000000000001` remain distinct even on runtimes whose native floating-point type would round them to the same representation.

The absolute magnitude of a portable schema/instance number is bounded to `2^53-1`. This is a portability bound, not a statement that every decimal is exactly representable as an IEEE-754 double; conforming implementations must preserve the exact mathematical comparison semantics before converting to a provider/runtime representation.

The portable `integer` type is deliberately stricter than mathematical JSON-Schema integer equivalence: its lexical JSON number must be representable as a base-10 integer without a fractional or exponent form and must fit `[-(2^53-1), 2^53-1]`. Thus `1` is an integer, while `1.0` and `1e0` are portable `number` values rather than portable `integer` values. This removes runtime-specific integer coercion from conformance.

## Instance validation

A schema is useful only if the produced value is checked against it. Conforming implementations therefore expose equivalent semantics for validating an instance against `portable_json_schema/v1`.

Instance validation must preserve strict JSON lexical checks before semantic validation. In particular, a duplicate-key object cannot become valid merely because the host language decoder would overwrite one duplicate.

If a schema declares `enum`, membership remains mandatory for nullable values as well. `null` is accepted only when both the type permits `null` and the enum explicitly contains `null`.

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
