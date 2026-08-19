# Assistant structured output

This document is normative for Core Spec `0.3`.

Assistant consumes the generic [`portable_json_schema/v1`](../json/README.md) contract. The JSON/schema profile is owned by Core's generic portable JSON domain rather than by Assistant itself.

This document defines only the Assistant-specific restrictions and provider boundary.

## Tool schemas

`ToolCall.arguments` is always a JSON object. Therefore a `ToolDefinition.input_schema` must be either:

- `{}` for the generic unconstrained-object compatibility form; or
- a `portable_json_schema/v1` schema whose root type is exactly non-nullable `object`.

A root `string`, `array`, nullable object, or other non-object tool schema is invalid.

Provider-generated arguments must be validated against the exact schema exposed for the matching tool before a host considers executing the call. Schema validation does not authorize execution; `ToolPolicy`, product permissions, and required user confirmation still apply independently.

## Response schemas

`GenerationRequest.response_schema`, when present, must satisfy `portable_json_schema/v1`. Unlike tool schemas, the response root may use any supported portable type.

Adapters may translate the portable schema into provider-native structured-output configuration. They must reject a schema they cannot preserve rather than silently weakening it.

## Provider extensions

A consumer that needs provider-specific schema features may keep that configuration in an adapter-owned envelope. Provider extensions are outside Core conformance and must not be serialized as if they were portable Assistant semantics.

## Boundary summary

```text
portablejson
    strict JSON + schema + instance semantics
          |
          v
assistant
    tool object-root restriction
    response structured-output contract
    provider translation boundary
          |
          v
host
    tool authorization/execution
    persistence/transport/provider SDKs
```
