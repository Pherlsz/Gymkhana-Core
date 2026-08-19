# OCR and structured extraction

Core Spec `0.4` defines provider-neutral extraction contracts for turning authorized document or media sources into evidence-backed observations, semantic field candidates, and, when a caller supplies a target schema, validated structured data.

The contract is deliberately broader than optical character recognition. A provider may use classic OCR, document parsing, speech recognition, multimodal models, deterministic parsers, or a composition of those techniques. Provider SDK objects and execution remain outside Core.

## Boundaries

Core defines:

- extraction modes and request semantics;
- logical source references and source modalities;
- evidence locations;
- raw observations;
- semantic field candidates;
- confidence assessments;
- ambiguity, validation, warning, and review state;
- schema-guided structured data through `portable_json_schema/v1`;
- conditional discovery semantics through `data_identification/v1`;
- deterministic validation and resource bounds.

Core does not define or implement:

- OpenAI, Gemini, Anthropic, Google Vision/Document AI, Azure Document Intelligence, Textract, or other provider DTOs/SDKs;
- HTTP upload, object storage, databases, queues, workers, billing, or persistence;
- source authorization, tenant ACLs, retention, encryption, or secret storage;
- provider prompts, model selection, transport, or orchestration loops;
- document-type-specific application models.

The host authorizes and supplies source material to an adapter. `SourceRef` carries only a non-secret logical ID, modality, and optional media type; Core does not carry a URI, bytes, storage key, or ACL decision.

## Extraction modes

### Schema-guided extraction

`schema_guided` is used when the caller knows the output fields. The request MUST contain a valid object-root `portable_json_schema/v1` target schema.

The result MUST contain a portable JSON object in `structured_data`. During exchange validation:

1. every evidence reference MUST point to a source declared by the request;
2. every field candidate MUST be backed by declared observations;
3. every `present` candidate MUST resolve through its non-root RFC 6901 JSON Pointer to a mathematically equivalent value in `structured_data`;
4. every `missing` candidate MUST point to a location absent from `structured_data`;
5. `structured_data` MUST be validated against the exact request target schema;
6. result `validation` MUST be `valid` when schema-instance validation succeeds and `invalid` when it fails.

A JSON `null` is a present JSON value. It is not the same as a missing field. `ValueState` preserves that distinction.

Schema-guided extraction does not require `data_identification/v1`.

### Discovery extraction

`discovery` is used when no target field schema is available. The request MUST NOT contain `target_schema`; the result MUST NOT contain `structured_data`, and result validation remains `not_validated`.

Discovery returns evidence-backed field candidates that a host may review, map to a later schema, or use to choose a document-specific extraction policy. If a schema-guided pass reveals unknown fields and the consumer wants discovery, the host starts a bounded discovery operation rather than silently changing the semantics of the active schema-guided request.

Discovery requires the built-in semantic skill `data_identification/v1`.

## `data_identification/v1`

The built-in discovery skill is semantic policy, not provider prompt text. A conforming implementation MUST preserve these invariants:

- preserve raw observations when available;
- require evidence for every observation and candidate;
- preserve ambiguity rather than silently selecting a unique interpretation;
- do not invent missing values;
- do not assume a jurisdiction;
- normalize only when an explicit Core or consumer canonicalizer is named.

The skill is conditional. It is not globally enabled merely because an extraction adapter uses an AI-capable provider.

Candidate semantic types and jurisdiction candidates are advisory namespaces. Discovering a string that resembles a CPF, CNPJ, EIN, IBAN, passport number, address, phone number, or other identifier does not validate its semantics.

## Evidence and observations

An `Observation` is provider-produced evidence before semantic normalization. It has a stable operation-local ID, one or more `EvidenceRef` values, optional exact observed text, optional portable JSON value, and optional confidence.

An evidence reference always names a request source and may additionally carry:

- `page`: a 1-based page number for `document` sources; zero means unspecified;
- `region`: a normalized rectangle represented in millionths of source width/height (`0..1_000_000`), avoiding pixel and floating-point dependencies;
- `text_span`: a zero-based, half-open UTF-8 byte range in a stable text surface exposed by the host/adapter;
- `time_range`: a zero-based, half-open millisecond interval for `audio` or `video` sources.

Evidence references do not copy source content. They are locators into host-owned source material.

## Candidates

A `FieldCandidate` has a non-root RFC 6901 JSON Pointer path, state (`present` or `missing`), evidence observation IDs, basis, optional confidence, optional semantic/jurisdiction candidates, ambiguity state, validation state, and review state.

Candidate basis is:

- `observed` — directly represented by cited observations;
- `derived` — deterministically derived from observations;
- `inferred` — requires semantic inference beyond direct observation.

A normalized candidate value is legal only when accompanied by an explicit canonicalizer namespace. Core never assumes that provider-formatted text is canonical.

Multiple candidates may share a path, particularly during discovery. Ambiguity MUST remain explicit until a host or reviewer resolves it. Multiple active candidates at the same path therefore require `ambiguous=true`; rejected alternatives do not keep the path active. A candidate carrying multiple semantic-type or jurisdiction alternatives likewise requires `ambiguous=true` unless it is rejected. An accepted candidate MUST NOT retain unresolved semantic/jurisdiction alternatives, and an accepted candidate cannot coexist with a non-rejected alternative at the same path.

## Confidence

Confidence uses an integer scale from `0` to `10000`; absence means unknown. The scale makes serialization portable and avoids binary floating-point differences.

Confidence is an assessment attached to an observation or candidate. Core does not claim that scores from different providers, models, document types, or extraction stages are globally calibrated or directly comparable. Hosts must not infer probabilistic equivalence merely from equal numeric scores.

## Validation and review

Validation status is `not_validated`, `valid`, or `invalid`. Review status is `unreviewed`, `needs_review`, `accepted`, or `rejected`.

These are state contracts, not persistence or UI implementations. A host owns who may review data, what validators are authorized, how review changes are persisted, and whether accepted data may enter a product model.

Stable warnings are limited to non-sensitive categories such as partial extraction, ambiguity, conflicting evidence, and unverified semantics. Validation errors and warnings MUST NOT embed raw source values.

## Portable JSON integration

All observation values, candidate values, normalized values, target schemas, and structured results use the generic `portablejson` contract. OCR does not duplicate JSON/schema parsing or exact-number semantics.

`schema_guided` target schemas use the object-root `portable_json_schema/v1` profile. Candidate-to-structured-data comparison uses portable JSON semantic equality, including mathematical decimal equality rather than binary floating-point equality.

## Authority and security

Source content, OCR text, model output, discovered fields, and extracted values are untrusted data. Text inside a document never becomes system/developer instruction authority, tool authorization, ACL policy, or executable side-effect approval merely because an AI-capable adapter processed it.

Core validation is not authorization. The host must authorize source access before data reaches an adapter and must authorize any later side effect that uses extracted data.

Raw sensitive values must not appear in stable validation errors, fingerprints, trace identifiers, or warning codes.

## Resource bounds

Implementations MUST reject work outside the portable bounds represented by the Go implementation and serialized schemas, including source, observation, candidate, evidence, warning, page/coordinate, identifier, text, and aggregate result-data limits. Observation `raw_text` is bounded by its UTF-8 encoded byte length (16 KiB), not Unicode code-point count.

These bounds prevent a provider-produced extraction from becoming an unbounded memory or validation workload.

## Assistant and token economy interaction

OCR is independent of the Assistant domain and depends directly on generic `portablejson` semantics.

An AI-backed adapter or host may compose Assistant infrastructure and `token_economy/v1`. When it does, token economy should minimize unnecessary pages/regions, repeated source text, and provider payload projection while preserving evidence, ambiguities, jurisdiction context, target-schema constraints, and unresolved review state.

This composition does not create a normative `ocr -> assistant` dependency.

## Conformance

Core Spec `0.4` adds `conformance/v0.4/ocr.json`. Applicable suites from Spec `0.3` are carried forward under `v0.4` unchanged in semantic meaning.
