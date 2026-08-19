# Changelog

All notable changes to this project will be documented in this file.

The project follows Semantic Versioning. Breaking changes remain explicit before `v1.0.0`.

## [0.6.0] - 2026-08-19

### Added

- Provider-neutral `ocr` Go package for bounded schema-guided and discovery extraction across document, image, text, audio, and video sources, with evidence references, raw observations, semantic field candidates, fixed-scale confidence, review/validation state, and stable non-sensitive errors.
- Conditional built-in `data_identification/v1` discovery skill that preserves evidence and ambiguity, distinguishes missing values from JSON `null`, avoids assumed jurisdiction, and permits normalization only through an explicit canonicalizer.
- Core Spec `0.4`, `spec/ocr/`, `schemas/ocr.schema.json`, and `conformance/v0.4/ocr.json`, with the complete Spec `0.3` conformance suite set carried forward unchanged in semantic meaning.
- Candidate-to-structured-data consistency checks using RFC 6901 JSON Pointers and exact mathematical numeric equality, plus bounded aggregate result-data validation.
- OCR request/result fuzz targets integrated into `make fuzz-smoke`.

### Changed

- OCR/extraction is now a first-class Core domain that depends directly on generic `portablejson`; Assistant infrastructure remains optional host/adapter composition rather than a required OCR dependency.
- `docs/OCR-FOUNDATION.md` now records rationale and deferred adapter/infrastructure work while the normative contract lives in `spec/ocr/`.

## [0.5.0] - 2026-08-19

### Added

- Provider-neutral `assistant` Go package with portable message roles, ordered text/image/audio/video/file content parts, tool definitions/calls/results, finish reasons, usage counters, and capability identifiers.
- Versioned Assistant skill contracts plus built-in `token_economy/v1`, covering concise responses, restatement avoidance, prior-context reuse, tool-result compaction, and mandatory preservation of instruction hierarchy and unresolved constraints.
- Reusable `AssistantDefinition`/`AssistantCatalog` profiles with default and optional modules, application instruction blocks, tools, tool safety policies, skills, bounded memory/learning policies, bounded execution budgets, and per-run module/model/credential overrides.
- Provider-neutral model catalogs and deterministic manual, ordered-fallback, and dynamic routing across current `free`, `paid`, `local`, and `unknown` access tiers, including automatic capability filtering from active Assistant modules and bounded model/provider failover.
- Managed/BYOK credential policy, ordered credential fallback, non-secret credential identities, provider quota observations, and local usage attribution without placing raw provider secrets in portable Core contracts.
- Runtime-local Go `ProviderAdapter`/`AdapterRegistry` boundaries with convenience IDs for common provider/gateway/local-runtime targets while keeping provider IDs extensible.
- Portable RAG contracts for lexical/vector/hybrid retrieval, original/rewrite/multi-query search, bounded candidate/context selection, optional reranking, grounding, citations, embedding/reranking model references, and untrusted retrieval evidence.
- Task-scoped Skill Build contracts with `TaskSignature`, evidence-backed learning proposals, bounded promotion policy, and protection against silently promoting free-form learned text into higher instruction authority.
- Content-free Assistant run traces for routing/fallback/tool/retrieval/usage diagnostics without persisting prompt, response, tool payload, evidence content, media URI, raw key, or secret-store reference data.
- Generic `portablejson` Go package and normative `spec/json/` contract for strict duplicate-safe JSON, `portable_json_schema/v1`, exact bounded decimal semantics, object-schema restrictions, and schema-instance validation reusable outside Assistant.
- Assistant structured-output/tool validation consumes the generic `portablejson` layer while retaining the initial `assistant.ValidatePortable*` APIs as compatibility wrappers.
- Core Spec `0.3` semantics for generic portable JSON plus Assistant configuration/model/BYOK routing, tool safety, RAG, memory, learning, quota/usage, observability, and structured-output integration.
- `schemas/assistant.schema.json`, `schemas/assistant-config.schema.json`, `schemas/assistant-runtime.schema.json`, and Assistant observability/configuration shapes for language-neutral serialized boundaries.
- Complete `conformance/v0.3` suite set with generic portable JSON/schema vectors, Assistant core/multimodal/configuration/hardening/structured vectors, and carried-forward normalize, temporal, and fingerprint suites.
- `docs/OCR-FOUNDATION.md` defining the next provider-neutral OCR/extraction boundary and conditional `data_identification/v1` discovery skill plan.
- Bounded Assistant and portable JSON/schema fuzzing integrated into `make fuzz-smoke`, plus local-only routing/JSON/schema performance benchmarks.

### Changed

- Core specification advances from `0.2` to `0.3` for generic portable JSON and the new Assistant contract family.
- Generic structured JSON/schema semantics are no longer conceptually owned by Assistant; future OCR/extraction, matching, solver, and other Core domains can reuse `portablejson` directly.
- Provider neutrality now distinguishes portable provider/model routing and opaque credential-reference contracts from excluded provider SDK types, raw credentials, volatile price/model availability tables, HTTP transport, persistence, and product orchestration.
- Streaming is represented as a portable capability but its event/delta wire contract is intentionally deferred until multiple provider adapters have exercised the semantics.
- Token-economy behavior is defined as a portable semantic profile rather than provider prompt text; consumers/adapters remain responsible for provider tokenization, context caching, and any lossy context compaction.

## [0.4.0] - 2026-08-18

### Added

- `fingerprint` Go package with exact SHA-256 byte/string fingerprints, canonical lowercase hexadecimal digests, strict digest parsing, and stable validation errors.
- Core FingerPrint frame v1 (`CFP\x01`) with explicit namespace, part count, and big-endian part lengths for portable domain separation and unambiguous ordered multipart fingerprints.
- Language-neutral fingerprint specification and shared conformance vectors under Core Spec `0.2`, including exact digest, framing, namespace separation, boundary preservation, and parsing behavior.
- Bounded fingerprint parse fuzzing integrated into `make fuzz-smoke`.

### Fixed

- Preserve `FormatHouseNumber` idempotency when punctuation would otherwise canonicalize a standalone number label or empty marker into a value such as `N`, `NUM`, or `SN` that the next pass removes.

### Changed

- Core specification advances from `0.1` to `0.2` for the new fingerprint operation family; `conformance/v0.2` carries the complete applicable normalize, temporal, and fingerprint suite set.
- Canonical JSON is explicitly deferred until a precise cross-language structured canonicalization contract is selected; fingerprinting never performs implicit Unicode, locale, JSON, or application normalization.

## [0.3.1] - 2026-08-18

### Fixed

- Preserve `FormatBlock` and `FormatApartment` idempotency when a canonical payload itself matches a recognized slot label, such as `bl. BL` or `apt. AP`.

### Changed

- Updated `golang.org/x/text` from `v0.40.0` to `v0.41.0`.
- Updated pinned GitHub Actions to `actions/checkout v7.0.1` and `actions/setup-go v7.0.0`.

## [0.3.0] - 2026-08-18

### Added

- Language-neutral Foundation structure with versioned `spec/`, shared `conformance/` vectors, and portable `schemas/`; Go remains the first implementation rather than the architectural source of truth.
- Initial Core specification `0.1` for global text/temporal semantics, generic typed identifiers, stable error codes, explicit jurisdiction modules, and an international-standards extension point.
- Shared normalization and temporal conformance suites plus a Go conformance runner; `make conformance` executes the current implementation against the language-neutral vectors.
- `FormatStreet`, `FormatHouseNumber`, `FormatBlock`, `FormatApartment`, `FormatCEP`, `FormatUF`, `FormatCity`, `FormatNeighborhood`, `FormatCountry`, and `FormatNationality` for address and origin slots.
- `FormatPersonName`, `FormatGender`, `FormatMaritalStatus`, `FormatBloodType`, and `FormatMembershipType` for profile civil fields.
- `FormatTeam`, `FormatSector`, `FormatClub`, `FormatHealthPlan`, `FormatCard`, `FormatCollection`, `FormatAnimal`, `FormatVehiclePlate`, `FormatVehicleModel`, `FormatVehicleColor`, and `FormatVehicleYear` for gincana catalogs. Empty, nenhum, and s/n stay empty. Formatters are idempotent where a canonical form exists.
- `CanonicalDocument` and `FormatDocument` for catalog document kinds, with check digits where they exist and UF folded into the persisted identifier (`114458/RS`) so state is not a separate column.
- Document kind `pis` (PIS/PASEP/NIT/NIS): 11 digits with the Caixa modulo-11 weights `3,2,9,8,7,6,5,4,3,2`. It is a separate type from CTPS and is not inferred from an unlabeled number.
- `IdentifyDocument` infers a catalog kind from an unlabeled number only when a single strong validator matches, so weak shapes such as RG or OAB are not guessed.
- `CanonicalCNPJ` and document kind `cnpj` accept the Receita Federal alphanumeric CNPJ (12 letters or digits plus two numeric check digits) while remaining compatible with existing numeric CNPJs.

### Changed

- Repository governance is now explicitly language-neutral, product-neutral, provider-neutral, and jurisdiction-neutral. Gymkhana Database is a consumer rather than the source of truth for Core semantics or roadmap.
- Specification versions are separated conceptually from language implementation/package versions so future TypeScript, Java, .NET/C#, Python, Rust, and PHP ports can evolve independently while declaring conformance to the same Core spec.

## [0.2.1] - 2026-07-13

### Fixed

- Updated `golang.org/x/text` to `v0.40.0` for current Unicode normalization and malformed-input fixes.
- Removed nonexistent custom labels from Dependabot configuration so scheduled updates open without warnings.

## [0.2.0] - 2026-07-13

### Added

- `normalize` package for Unicode-aware text, CPF, email, Brazilian phone, digit, and alphanumeric canonicalization.
- `civiltime` package with immutable `CivilDate` and `YearMonth` values, strict parsing, canonical serialization, comparison, leap-year rules, and month arithmetic.
- Table-driven unit tests and bounded fuzz targets for normalization and civil-time round trips.
- In-PR semantic version preparation with automatic GitHub Releases after merge.

### Changed

- `make fuzz-smoke` executes every declared fuzz target independently and is part of `make check`.
- Release versions and changelog entries are carried by the functional pull request instead of a generated release pull request.

## [0.1.0] - 2026-07-13

### Added

- Milestone 0 repository bootstrap.
- Pinned Go quality and security tooling.
- CI, security, and release workflow foundations.
- Contribution, security, and repository-boundary documentation.
