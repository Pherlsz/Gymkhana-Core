# OCR / extraction foundation

Status: implemented by the Core Spec `0.4` OCR/extraction workstream. The normative contract is now [`../spec/ocr/README.md`](../spec/ocr/README.md), with serialized shapes in [`../schemas/ocr.schema.json`](../schemas/ocr.schema.json) and language-neutral behavior in `conformance/v0.4/ocr.json`.

This document records design rationale and deferred work. It is not a second normative source of truth.

## Implemented boundary

The Foundation models:

```text
authorized document/media source
  -> evidence references
  -> raw observations
  -> semantic field candidates
  -> optional schema-guided structured data
  -> validation/review state
```

It deliberately supports both:

1. `schema_guided` extraction when the caller knows the target fields; and
2. `discovery` when the source is unknown/unstructured and field identification is itself part of the task.

Generic JSON/schema semantics remain in `portablejson`. OCR does not import Assistant to represent extraction semantics.

## Data identification

`data_identification/v1` is now a built-in OCR semantic skill, but only for discovery operations. It preserves raw observations/evidence/ambiguity, refuses to invent missing values or assume jurisdiction, and permits normalization only with an explicit canonicalizer.

It is intentionally not a provider prompt template. Adapters may translate the semantics into their provider-specific request format.

## AI-backed extraction

An adapter may use Assistant infrastructure, multimodal models, or `token_economy/v1` when useful. Such composition belongs to the host/adapter layer. OCR itself remains valid for classic OCR engines, deterministic parsers, speech recognition, and non-AI extraction.

Document content remains untrusted data. Extracted text cannot grant instruction authority, authorize tools, or bypass host ACL decisions.

## Deferred items

The Foundation intentionally does not freeze:

- provider adapters or provider SDK DTOs;
- upload/storage/database/queue contracts;
- provider-specific confidence calibration;
- document-type classifiers/catalogs;
- table/reading-order/layout graph semantics beyond bounded evidence locators;
- handwriting/quality taxonomies;
- provider batch/async job protocols;
- review UI or persistence;
- billing/cost accounting;
- autonomous schema synthesis;
- automatic promotion of discovered semantics into Core normalization rules.

Those should be added only after multiple real consumers/providers demonstrate stable reusable semantics.

## Next integration step

After the Core PR is fully validated and released, consumers can implement provider adapters that translate source material and provider responses into these contracts. Consumer integration must preserve source authorization, exact target schemas, evidence references, and review boundaries.
