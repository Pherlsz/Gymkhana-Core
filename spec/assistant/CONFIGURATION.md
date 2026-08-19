# Assistant configuration and model routing

This document is normative for Core Spec `0.3`.

## Assistant definitions

An `AssistantDefinition` is a reusable profile for one application function. Applications may create any number of profiles, for example a research assistant, coding assistant, support assistant, OCR assistant, or domain specialist.

A profile owns portable configuration only:

- stable assistant ID/name/description;
- application-owned `system`/`developer` instruction blocks;
- default and optional modules;
- built-in Core skills;
- allowed tool names;
- model-selection policy;
- credential policy;
- optional retrieval/RAG policy;
- bounded execution budgets.

Profiles do not own provider SDK objects, database rows, HTTP sessions, or raw API secrets.

## Modular functionality

Default modules are listed in `modules`. Modules available but disabled by default are listed in `optional_modules`.

Portable modules in Spec `0.3` are:

- `text`;
- `vision`;
- `audio_input`;
- `video_input`;
- `file_input`;
- `tools`;
- `retrieval`;
- `memory`;
- `structured_output`.

A run may disable a default module or explicitly enable an optional module. It may not enable a module that the Assistant definition did not declare.

This separates capability authorization/configuration from model capability. Enabling `vision`, for example, does not guarantee that every registered model supports images; model routing must filter for the corresponding model capability.

The runtime derives these minimum generation-model capabilities automatically:

| Module | Required model capability |
| --- | --- |
| `text` | `text` |
| `vision` | `image_input` |
| `audio_input` | `audio_input` |
| `video_input` | `video_input` |
| `file_input` | `file_input` |
| `tools` | `tool_calling` |
| `structured_output` | `structured_output` |

`retrieval` and `memory` may be implemented outside the generation model and therefore do not imply a model capability by themselves.

## Provider adapters

Concrete providers are represented by adapters. The language-neutral contract does not depend on a provider SDK or HTTP shape.

The current Go implementation exposes a `ProviderAdapter` substitution boundary that normalizes:

- provider metadata;
- live model-catalog discovery;
- non-streaming generation requests/responses.

Provider adapters may internally use vendor SDKs, HTTP, local inference runtimes, gateways, or other mechanisms. Provider-specific types must not cross the Core boundary.

Core defines convenience provider IDs for common adapter targets (`openai`, `anthropic`, `google`, `openrouter`, `groq`, `ollama`, `mistral`, `xai`) but this is not a closed enum. Any valid provider ID may be registered.

## Model catalogs

Model names, availability, free tiers, pricing, context limits, and capabilities change independently from Core releases. Core therefore does not hardcode a permanent model list or price table.

Adapters return current `ModelDescriptor` values containing:

- provider/model reference;
- current access tier (`free`, `paid`, `local`, `unknown`);
- model roles (`generation`, `embedding`, `reranking`);
- portable capabilities;
- reported context/output limits when known.

Access tier is runtime catalog metadata, not a timeless property of a model. A model that is free today may be paid tomorrow without requiring a Core specification change.

## Model selection

Three portable modes exist.

### Manual

`manual` selects exactly one provider/model pair. This is used for explicit user selection and per-run model overrides.

The selected model still has to satisfy Assistant module/capability and policy constraints.

### Ordered fallback

`ordered_fallback` carries an explicit ordered candidate list. The runtime skips unavailable/ineligible candidates and retains caller order for the remaining models.

This is appropriate when the application has deliberately benchmarked a preferred sequence.

### Dynamic

`dynamic` resolves against the current adapter-supplied catalog.

The policy can filter by:

- allowed providers;
- allowed access tiers;
- required capabilities.

`preferred_access` supplies deterministic tier ordering, for example `free` before `paid`. It does not fabricate quality/pricing information and does not override capability requirements.

Within the same preference class, ordering is deterministic by provider/model reference. Applications that need quality/latency ranking should supply an explicit ordered fallback or an application-owned router based on measured telemetry.

## Free, paid, and local models

A profile may be free-only, paid-only, local-only, or any combination by configuring `allowed_access`.

A common cost-aware configuration is:

```json
{
  "mode": "dynamic",
  "allowed_access": ["free", "local", "paid"],
  "preferred_access": ["free", "local", "paid"]
}
```

Paid fallback is never silently enabled when the policy excludes `paid`.

## BYOK

Core supports `managed` and `byok` credential modes.

Raw API keys never belong in an `AssistantDefinition`, `CredentialRef`, log field, validation error, fingerprint, or model catalog.

A `CredentialRef` contains only:

- provider ID;
- credential mode;
- opaque application-owned reference/handle.

For BYOK, the host application stores the supplied key in an appropriate secure or ephemeral credential store and gives the adapter an opaque reference. The adapter resolves that reference through application-owned infrastructure configured outside Core.

This allows one Assistant definition to run with application credentials, a user's own key, or a local model without changing the portable profile.

## Per-run overrides

`RunOverrides` permits temporary changes without mutating the persisted Assistant profile:

- enable declared optional modules;
- disable default modules;
- select one concrete model manually;
- bind one credential reference.

A model override does not bypass allowed provider/access/capability constraints. A BYOK credential bound to a manually selected model must target the same provider.

## Execution budgets

Assistants with tools/retrieval/agent-like workflows must remain bounded. `ExecutionBudget` provides portable limits for turns, tool calls, retrieval rounds, and output tokens.

A zero budget field means "use the runtime/application default"; it never means infinite execution.

Core Spec `0.3` does not define an autonomous unbounded agent loop.
