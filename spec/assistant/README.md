# Assistant semantics

Core Assistant defines provider-neutral conversation and tool contracts that can be shared by applications without importing a model-provider SDK into domain code.

The initial contract is intentionally transport-neutral. It does not define HTTP, authentication, retries, model IDs, pricing, provider options, persistence, orchestration, or prompt-management policy.

## Roles

Portable message roles are:

- `system` — global behavioral or contextual instruction;
- `developer` — application/developer instruction distinct from end-user input;
- `user` — end-user supplied content;
- `assistant` — model/assistant produced content;
- `tool` — tool execution result content.

Adapters may need to merge or translate roles when a provider does not expose the same native role set. They must not silently reinterpret `user` content as a higher-authority instruction.

## Messages and content parts

A `Message` contains a valid role and one or more ordered content parts. Part order is semantically significant.

Portable part types are:

- `text` — non-empty UTF-8 text;
- `image` — an external media reference intended as image input;
- `file` — an external media reference intended as file/document input;
- `tool_call` — an assistant request to invoke a named tool with structured arguments;
- `tool_result` — the result corresponding to a prior tool call.

Each part is a discriminated union: exactly the payload matching its `type` is present. A part carrying fields from another variant is invalid.

### Media references

Media is represented by an opaque non-empty `uri`, with optional `media_type` and `name` metadata. Core does not fetch the URI and does not constrain URI schemes in Spec `0.3`; adapters and consuming applications own accessibility, authorization, size, lifetime, and supported-scheme policy.

Inline binary encoding, uploads, provider file IDs, and media transformation remain adapter/application concerns in this revision.

## Tools

A `ToolDefinition` contains:

- `name` — portable tool identifier;
- optional `description`;
- `input_schema` — a non-null JSON object describing accepted arguments.

The portable tool-name subset is ASCII and matches:

```text
[A-Za-z_][A-Za-z0-9_-]{0,63}
```

This conservative subset is intended to travel across providers without provider-specific name rewriting.

Core requires `input_schema` to be a JSON object but does not enforce a particular JSON Schema dialect in Spec `0.3`, because provider schema support differs. Consumers may impose a stricter dialect.

A `ToolCall` contains an opaque non-empty Unicode `id` of at most 256 Unicode scalar values, a valid tool `name`, and `arguments` as a JSON object. Argument property ordering is not semantic.

A `ToolResult` references the original call through `call_id`, contains one or more result content parts, and may set `is_error`. Tool-result content may contain `text`, `image`, or `file`; nested `tool_call` and `tool_result` parts are invalid.

## Finish reasons

Portable finish reasons are:

- `stop` — normal completion;
- `length` — provider/model output limit reached;
- `tool_calls` — generation paused/completed to request one or more tools;
- `content_filter` — generation stopped by a safety/content filter;
- `error` — generation terminated because of an execution/provider error;
- `other` — a provider reason with no more precise portable mapping.

Provider-specific raw reasons may be retained by adapters outside the Core contract when diagnostics require them.

## Usage

`Usage` exposes non-negative counters for:

- `input_tokens`;
- `output_tokens`;
- `cached_input_tokens`;
- `reasoning_tokens`.

Token accounting and tokenization are provider/model-defined. Counts from different models are not assumed to be directly comparable. Zero means not reported or zero consumed; Core does not infer missing counts.

## Capabilities

Portable capability identifiers are:

- `text`;
- `image_input`;
- `file_input`;
- `tool_calling`;
- `structured_output`;
- `streaming`.

A capability set contains no duplicates. Capability discovery itself is adapter/application-owned; Core only defines portable names.

## Stable errors

Assistant validation uses stable, non-localized codes:

- `empty`;
- `invalid_role`;
- `invalid_content`;
- `invalid_content_type`;
- `invalid_media`;
- `invalid_tool_name`;
- `invalid_json`;
- `invalid_tool_call`;
- `invalid_tool_result`;
- `invalid_finish_reason`;
- `invalid_usage`;
- `invalid_capability`;
- `duplicate_capability`.

Errors must not embed prompt text, tool arguments, tool results, URIs, or other sensitive original values.

## Serialization

Portable serialized contracts use the JSON field names defined under `schemas/assistant.schema.json`. Provider adapters may expose idiomatic language APIs but must preserve the same semantics at serialization boundaries.

Unknown provider-specific fields do not belong in the Core serialized shape. Applications that need provider metadata should keep it in adapter-owned envelopes rather than weakening the portable contract.

## Streaming

Spec `0.3` defines `streaming` as a capability but deliberately does not standardize a `StreamEvent` shape yet. Providers differ materially in text deltas, tool-argument fragments, usage timing, error events, and event ordering. A later specification revision should define streaming only after the contract has been exercised against multiple provider adapters.

## Security and privacy

Assistant contracts can carry sensitive prompts, files, tool arguments, and tool results. Core performs no redaction, encryption, persistence, logging, or retention. Consuming applications own those policies.

Tool definitions and tool calls are data, not authorization. A consumer must independently authorize every side effect before executing a tool call.
