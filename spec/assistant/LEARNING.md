# Assistant learning and skill build

This document is normative for Core Spec `0.3`.

## Goal

Assistant executions may reveal reusable lessons: a retrieval strategy that avoids a repeated miss, a tool-selection hint that fixes a recurring error, a model-routing preference backed by measured outcomes, or a behavior hint that reduces unnecessary work.

Core calls these reviewable learning proposals rather than silently mutating an Assistant definition.

## Safety boundary

Learning is proposal-driven. A completed run may produce a `SkillLearningProposal`, but Core does not automatically rewrite system/developer instructions, tool authorization, credential policy, secrets, provider configuration, or persistence.

A proposal contains:

- stable proposal ID;
- target assistant ID;
- version/revision;
- learning scope;
- concise reusable summary;
- one or more evidence references to prior runs/outcomes.

Raw prompts, tool payloads, credentials, private retrieved content, and hidden reasoning are not required by the portable proposal contract.

## Portable scopes

Spec `0.3` defines:

- `skill_hint` — reusable behavioral/process lesson;
- `retrieval_hint` — reusable retrieval/query/reranking lesson;
- `tool_hint` — reusable tool-selection or tool-usage lesson;
- `model_routing_hint` — reusable routing/fallback observation.

Consumers may evaluate, deduplicate, approve, reject, expire, or promote proposals into application-owned configuration. Promotion must create an explicit revision so behavior remains auditable and reversible.

## Intended workflow

```text
run -> outcome/trace -> learning proposal -> evaluation/dedup -> approval/policy -> new assistant/skill revision
```

The same task family can therefore benefit from previous executions without granting an assistant unrestricted self-modification.

## Evidence and quality

A proposal should not be promoted from one accidental success when the change is material. Consumers should use repeated evidence or eval results when available. Conflicting proposals remain explicit rather than being merged silently.

Learning must preserve instruction hierarchy, authorization boundaries, grounding requirements, and unresolved constraints. A token-saving lesson cannot remove required evidence or permissions checks.
