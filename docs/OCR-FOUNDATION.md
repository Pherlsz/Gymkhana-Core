# OCR / extraction foundation

Status: implemented by the Core Spec `0.4` OCR/extraction workstream. The normative contract is [`../spec/ocr/README.md`](../spec/ocr/README.md), with serialized shapes in [`../schemas/ocr.schema.json`](../schemas/ocr.schema.json) and language-neutral behavior in `conformance/v0.4/ocr.json`.

This document records design rationale and deferred work. It is not a second normative source of truth.

## Implemented boundary

The Foundation models authorized source -> evidence -> raw observations -> semantic field candidates -> optional schema-guided structured data -> validation/review state.

It supports `schema_guided` extraction when the caller knows the target fields and `discovery` when field identification is part of the task. Generic JSON/schema semantics remain in `portablejson`; OCR does not import Assistant.

## Discovery semantics

Discovery correctness is encoded directly in the OCR contract instead of a built-in OCR skill. Implementations preserve raw observations when available, require evidence, keep ambiguity explicit, do not invent missing values, do not assume jurisdiction, and normalize only with an explicit canonicalizer.

Provider adapters may translate these invariants into provider-specific prompts or request structures, but those prompt mechanics are not Core API.

## AI-backed extraction

An adapter may use Assistant infrastructure, multimodal models, or `token_economy/v1` when useful. Such composition belongs to the host/adapter layer. OCR remains valid for classic OCR engines, deterministic parsers, speech recognition, and non-AI extraction.

Document content remains untrusted data and cannot grant instruction authority, authorize tools, or bypass host ACL decisions.

## Deferred items

The Foundation intentionally does not freeze provider adapters/SDK DTOs, upload/storage/database/queue contracts, provider-specific confidence calibration, document-type classifiers/catalogs, detailed layout graphs, handwriting/quality taxonomies, provider batch protocols, review UI/persistence, billing/cost accounting, autonomous schema synthesis, or automatic promotion of discovered semantics into Core normalization rules.

## Next integration step

Consumers can implement provider adapters that translate source material and provider responses into these contracts while preserving authorization, exact target schemas, evidence, ambiguity, and review boundaries.
