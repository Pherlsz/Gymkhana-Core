# Portable structured output

This document is normative for Core Spec `0.3`.

Provider schema support differs materially. Core therefore defines `portable_json_schema/v1`, a conservative profile intended to preserve the same validation intent across provider adapters and future language implementations.

## JSON object boundary

Tool arguments and portable schema objects are untrusted input. Core rejects:

- invalid UTF-8;
- duplicate object keys at any depth;
- unpaired UTF-16 surrogate escapes;
- trailing JSON values;
- excessive depth or node count;
- objects larger than the portable byte bound.

Duplicate-key rejection is required because common runtimes disagree on whether first-value, last-value, error, or duplicate-preserving behavior applies.

## `portable_json_schema/v1`

Supported base types:

- `object`;
- `array`;
- `string`;
- `integer`;
- `number`;
- `boolean`;
- `null`.

A nullable value may use exactly one base type plus `null`.

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

Unknown keywords are invalid rather than silently ignored.

Typed object schemas are strict:

- `additionalProperties` must be `false`;
- every declared property must appear exactly once in `required`;
- property count and nesting are bounded.

`{}` is the explicit compatibility form for an unconstrained JSON object. It should be used only when a strict schema is unavailable; adapters may not claim strict structured-output guarantees for this form.

Array sizes, schema depth, property counts, and portable numeric bounds are capped by the implementation/conformance contract to prevent pathological schemas.

## Tool schemas

`ToolCall.arguments` is always a JSON object. Therefore a `ToolDefinition.input_schema` must be either:

- `{}` for an unconstrained object; or
- a `portable_json_schema/v1` schema whose root type is exactly `object`.

A root `string`, `array`, nullable object, or other non-object tool schema is invalid.

## Response schemas

`GenerationRequest.response_schema`, when present, must satisfy `portable_json_schema/v1`. Unlike tool schemas, the response root may use any supported portable type.

Adapters may translate the portable schema into provider-native structured-output configuration. They must reject a schema they cannot preserve rather than silently weakening it.

## Provider extensions

A consumer that needs a provider-specific JSON Schema keyword may keep that configuration in an adapter-owned envelope. Provider extensions are outside Core conformance and must not be serialized as if they were portable Assistant semantics.
