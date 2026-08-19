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

`Usage` exposes counters for:

- `input_tokens`;
- `output_tokens`;
- `cached_input_tokens`;
- `reasoning_tokens`.

Every counter is an integer in the inclusive range `0..9007199254740991` (`2^53-1`), the largest integer that remains exact in JSON-backed JavaScript/TypeScript runtimes. Token accounting and tokenization are provider/model-defined. Counts from different models are not assumed to be directly comparable. Zero means not reported or zero consumed; Core does not infer missing counts.

## Capabilities

Portable capability identifiers are:

- `text`;
- `image_input`;
- `file_input`;
- `tool_calling`;
- `structured_output`;
- `streaming`.

A capability set contains no duplicates. Capability discovery itself is adapter/application-owned; Core only defines portable names.

## Built-in skills

Assistant skills are versioned provider-neutral behavior profiles. They are not provider prompts, authorization rules, hidden chain-of-thought instructions, or model-specific configuration objects. A provider adapter may realize a skill through native settings, context construction, concise instructions, or a combination of those mechanisms, but the observable intent must remain portable.

Spec `0.3` defines one built-in skill:

### `token_economy/v1`

The canonical token-economy profile enables these behaviors:

- `prefer_concise_responses` — prefer the shortest complete answer that satisfies the request;
- `avoid_restatement` — do not repeat user input, already-established decisions, or prior output unless repetition is needed for correctness;
- `reuse_prior_context` — reference stable prior context instead of duplicating it when the consumer/provider can preserve the reference safely;
- `compact_tool_results` — when the consumer controls tool-result projection, include only fields needed for the current task rather than forwarding irrelevant payload;
- `preserve_instruction_hierarchy` — token reduction must never weaken or remove higher-authority instructions;
- `preserve_unresolved_constraints` — compaction must retain unresolved requirements, decisions, identifiers, error conditions, and other state still needed to complete the task.

The last two properties are mandatory for every valid `token_economy/v1` profile. At least one of the four optimization properties must be enabled.

The skill does **not** define a universal tokenizer, exact input-token estimator, lossy summarizer, automatic message deletion policy, or guaranteed token-savings percentage. Tokenization differs by provider/model, and deleting/summarizing context without application knowledge can change semantics. Consumers that perform context compaction remain responsible for preserving instruction authority, tool-call/result linkage, required evidence, and unresolved constraints.

`BuiltinTokenEconomySkill`/equivalent language APIs return the canonical profile with every defined behavior enabled. `BuiltinSkills`/equivalent language APIs include this skill by default. Consumers may explicitly select a narrower skill set, but a modified profile must still satisfy the safety-preservation invariants above.

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
- `duplicate_capability`;
- `invalid_skill`;
- `duplicate_skill`.

Errors must not embed prompt text, tool arguments, tool results, URIs, or other sensitive original values.

## Serialization

Portable serialized contracts use the JSON field names defined under `schemas/assistant.schema.json`. Provider adapters may expose idiomatic language APIs but must preserve the same semantics at serialization boundaries.

Unknown provider-specific fields do not belong in the Core serialized shape. Applications that need provider metadata should keep it in adapter-owned envelopes rather than weakening the portable contract.

## Streaming

Spec `0.3` defines `streaming` as a capability but deliberately does not standardize a `StreamEvent` shape yet. Providers differ materially in text deltas, tool-argument fragments, usage timing, error events, and event ordering. A later specification revision should define streaming only after the contract has been exercised against multiple provider adapters.

## Security and privacy

Assistant contracts can carry sensitive prompts, files, tool arguments, and tool results. Core performs no redaction, encryption, persistence, logging, or retention. Consuming applications own those policies.

Tool definitions and tool calls are data, not authorization. A consumer must independently authorize every side effect before executing a tool call.
