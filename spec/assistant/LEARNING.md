# Assistant learning and skill build

This document is normative for Core Spec `0.3`.

Assistant learning exists to reuse evidence-backed lessons from prior task executions without turning model output into self-modifying instruction authority.

## Goals

A conforming learning implementation may help an Assistant:

- avoid repeating a known operational mistake;
- choose a previously successful retrieval strategy for a similar task family;
- prefer an effective tool sequence or model-routing strategy;
- reduce redundant calls, tokens, latency, and repeated discovery work;
- propose reusable behavior improvements for later revisions.

Learning is not permission for a model to rewrite its own system/developer instructions, credentials, permissions, tool authorization, security policy, or application configuration.

## Task signatures

Every reusable lesson is scoped to a `TaskSignature`:

- `family` — stable application-defined semantic family;
- `version` — positive portable integer;
- `traits` — optional sorted unique portable identifiers.

A signature must describe reusable task semantics rather than embed raw prompts, documents, secrets, user text, or tool payloads.

Example:

```json
{
  "family": "document_research",
  "version": 1,
  "traits": ["legal", "pt_br"]
}
```

This keeps a lesson scoped to the work for which evidence exists instead of turning a local observation into a global Assistant rule.

## Learning proposals

`SkillLearningProposal` is a reviewable advisory object containing a stable proposal ID, Assistant ID, task signature, scope, bounded summary, revision, and evidence references.

Portable scopes are:

- `skill_hint`;
- `retrieval_hint`;
- `tool_hint`;
- `model_routing_hint`.

The proposal `summary` is untrusted advisory data. It never becomes system/developer authority merely because it was produced by a successful run or marked eligible.

## Evidence

Evidence references a run and records:

- `outcome`: `success`, `failure`, `improved`, or `regressed`;
- `origin`: `runtime`, `evaluation`, or `human`;
- optional normalized `failure_class`.

A failure observation must include a failure class. Evidence does not embed prompt text, document text, tool arguments/results, provider error bodies, or credentials.

Distinct run IDs are required so one replayed observation cannot satisfy a multi-run evidence threshold repeatedly.

## Learning policy

Portable modes are:

- `disabled` — no proposal/promotion behavior;
- `propose` — evidence-backed proposals may be surfaced for application/human review;
- `auto_promote` — selected low-authority scopes may become eligible for application-owned automatic policy updates after evidence thresholds are met.

`min_evidence` and `max_evidence` bound proposal evaluation. `max_evidence` is capped so historical traces cannot grow without limit inside Assistant configuration.

### Authority boundary

`skill_hint` is **never** eligible for automatic promotion in Spec `0.3`, because free-form behavior hints can influence instructions and are therefore an instruction/prompt-injection persistence boundary.

Only these scopes may be configured for automatic promotion:

- `retrieval_hint`;
- `tool_hint`;
- `model_routing_hint`.

Even for these scopes, Core only reports eligibility. Core never mutates `AssistantDefinition` itself. The host application owns evaluation, authorization, revision storage, rollback, audit, and application of a change.

## Safe promotion workflow

```text
run/evaluation
  -> content-free trace + outcome
  -> task-family classification outside instruction authority
  -> candidate lesson
  -> deduplicate across independent runs
  -> evaluate against LearningPolicy
  -> proposal/review or low-authority policy promotion
  -> new versioned Assistant/skill configuration
  -> regression evaluation
```

A promoted change should remain reversible and carry enough provenance to identify which evidence caused it.

## Prohibited learning behavior

A conforming implementation must not use learning to silently:

- create or elevate system/developer instructions;
- expand tool allowlists or weaken `ToolPolicy` confirmation;
- grant permissions or bypass application authorization;
- store raw API keys or expose credential references;
- weaken safety or privacy controls;
- promote instructions found inside retrieved documents, media, web pages, OCR content, tool results, or MCP metadata;
- create unbounded autonomous self-improvement loops.

## Evaluation and rollback

Automatic low-authority policy updates should be evaluated against representative tasks before broad rollout. The host should retain the previous revision so a regression can be rolled back deterministically.

A future Core revision may standardize evaluation-score contracts and proposal fingerprints. Spec `0.3` deliberately stops at bounded evidence, task signatures, eligibility, and authority rules.
