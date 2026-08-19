# OCR and structured extraction

Core Spec `0.4` defines provider-neutral extraction contracts for turning authorized document or media sources into evidence-backed observations, semantic field candidates, and, when a caller supplies a target schema, validated structured data.

The contract is deliberately broader than optical character recognition. A provider may use classic OCR, document parsing, speech recognition, multimodal models, deterministic parsers, or a composition of those techniques. Provider SDK objects and execution remain outside Core.

## Boundaries

Core defines extraction modes, logical source references, evidence locations, raw observations, semantic field candidates, confidence, ambiguity/validation/review state, schema-guided structured data, direct discovery invariants, deterministic validation, and resource bounds.

Core does not define provider DTOs/SDKs, upload/storage/databases/queues/workers/billing/persistence, source authorization/ACLs, provider prompts/model selection/transport/orchestration, or document-type-specific application models.

## Extraction modes

`schema_guided` is used when the caller knows the output fields. The request MUST contain a valid object-root `portable_json_schema/v1` target schema. Present candidates resolve through non-root RFC 6901 pointers to mathematically equivalent `structured_data`; missing candidates point to absent locations; structured data is validated against the exact target schema.

`discovery` is used when no target field schema is available. The request MUST NOT contain `target_schema`; the result MUST NOT contain `structured_data`, and result validation remains `not_validated`.

Discovery behavior is part of the extraction contract itself. Core does not expose a separate OCR skill or provider prompt profile for it.

## Discovery invariants

A conforming discovery implementation MUST preserve raw observations when available, require evidence for every observation and candidate, preserve ambiguity, not invent missing values, not assume jurisdiction, and normalize only when an explicit Core or consumer canonicalizer is named.

Candidate semantic types and jurisdiction candidates are advisory namespaces; recognition does not itself validate semantics.

## Evidence and observations

An `Observation` is provider-produced evidence before semantic normalization. Evidence may carry 1-based document page, normalized millionth-scale region, half-open UTF-8 byte text span, or half-open millisecond audio/video range. Evidence references locate host-owned material rather than copying it.

## Candidates

A `FieldCandidate` has a non-root RFC 6901 path, present/missing state, observation IDs, observed/derived/inferred basis, optional confidence, semantic/jurisdiction alternatives, ambiguity, validation, and review state. JSON `null` is a present value, not missing.

Normalized values require an explicit canonicalizer. Multiple active same-path candidates and multi-valued semantic/jurisdiction interpretations keep ambiguity explicit until resolved according to the review invariants.

## Confidence, validation, and review

Confidence uses integer `0..10000`; absence means unknown, and scores are not globally calibrated across providers/models/document types.

Validation is `not_validated`, `valid`, or `invalid`. Review is `unreviewed`, `needs_review`, `accepted`, or `rejected`. These are state contracts, not persistence/UI implementations.

## Portable JSON integration

Observation/candidate/normalized values, target schemas, and structured results use generic `portablejson`. OCR does not duplicate JSON/schema parsing or exact-number semantics.

## Authority and security

Source content, OCR text, model output, discovered fields, and extracted values remain untrusted data. They do not become instruction authority, tool authorization, ACL policy, or side-effect approval. Raw sensitive values must not appear in stable errors or warning codes.

## Resource bounds

Implementations reject work outside the portable source/observation/candidate/evidence/warning/page/coordinate/identifier/text/aggregate-result bounds. Observation raw text is bounded by UTF-8 bytes.

## Assistant interaction

OCR is independent of Assistant and depends directly on `portablejson`. An AI-backed host may compose Assistant infrastructure and `token_economy/v1`; this creates no normative `ocr -> assistant` dependency.

## Conformance

Core Spec `0.4` adds `conformance/v0.4/ocr.json`. Applicable Spec `0.3` semantics are carried forward where still part of the `0.4` public boundary; `0.4` intentionally removes Assistant memory/learning contracts and the separate OCR discovery-skill abstraction.
