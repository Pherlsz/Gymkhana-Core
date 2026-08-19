# Assistant tool safety

This document is normative for Core Spec `0.3`.

Tool definitions and tool calls are data. They do not grant permission to execute a side effect.

## Tool policy

`ToolPolicy` classifies one allowed tool independently from its implementation:

- effect: `unknown`, `read_only`, `write`, or `destructive`;
- confirmation: `never`, `side_effect`, or `always`;
- whether concurrent execution is safe;
- whether repeated execution is idempotent;
- optional per-tool call limit;
- optional timeout;
- optional maximum result projection size.

Zero numeric limits mean use a bounded host/runtime default. Zero never means unlimited execution.

## Conservative defaults

A tool without an explicit policy resolves to:

```text
effect = unknown
confirmation = always
```

Additional invariants:

- `write` tools may not use `confirmation=never`;
- `destructive` tools require `confirmation=always`;
- `unknown` tools require `confirmation=always`.

Application authorization is still mandatory even if a policy says confirmation is not required. Tool policy describes execution risk/interaction; it is not an ACL or permission system.

## Tool allowlist

An Assistant may call only tools declared in its `tools` allowlist. A provider adapter response containing a tool call outside the request allowlist is invalid.

Tool call IDs must be unique within a conversation request. Every tool result must resolve one outstanding prior tool call exactly once. A new generation request must not be sent with unresolved client-side tool calls.

## Result projection

Tool results can be large and may contain sensitive data. `token_economy/v1` encourages forwarding only the fields required by the current task. `max_result_bytes` provides an application/runtime bound for that projection.

Compaction must not break tool-call/result linkage, hide an execution error, discard evidence required for grounding, or remove unresolved constraints.

## MCP and remote tool catalogs

MCP or another tool-discovery protocol may be implemented as an adapter into the same Core tool catalog. External tool metadata and annotations remain untrusted until the host trusts the server and validates its policy.

Connecting a tool server must not silently expand an Assistant's allowlist or weaken `ToolPolicy`. The host owns user consent, authorization, authentication, data-sharing UI, and network policy.
