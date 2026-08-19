# Machine-readable schemas

Schemas define portable serialized contracts that can be consumed by multiple language implementations and generators.

They complement, but do not replace, the normative prose in `spec/` and the observable behavior in `conformance/`.

Use schemas for data contracts where generation is appropriate. Do not attempt to generate nontrivial algorithms solely to avoid writing idiomatic implementations in each language.

Current schemas:

- `conformance.schema.json` — shape of shared conformance suites;
- `assistant.schema.json` — provider-neutral Assistant messages, multimodal content, tools, finish reasons, usage, capabilities, and built-in skills;
- `assistant-config.schema.json` — reusable Assistant definitions, modules, model catalogs/routing, credential/fallback policy, tool/RAG/memory/learning policy, and execution budgets;
- `assistant-runtime.schema.json` — per-run overrides plus normalized provider-adapter request/response metadata;
- `assistant-observability.schema.json` — provider quota observations, local usage attribution, content-free run traces, task signatures, and learning proposals;
- `ocr.schema.json` — provider-neutral OCR/extraction requests, source/evidence references, observations, field candidates, review/validation state, warnings, and `data_identification/v1`.

`portable_json_schema/v1` is itself a semantic subset/profile defined under `spec/json/` and tested through the active `portablejson.json` conformance suite. It is not represented by pretending that one ordinary JSON Schema document can replace its strict lexical, exact-number, resource-bound, and instance-validation requirements.

Likewise, `ocr.schema.json` describes serialized shape but does not replace semantic checks such as source-reference integrity, evidence modality, aggregate data bounds, candidate-to-structured-data equality, or validation against the exact caller target schema.

Provider SDK objects, raw credentials, HTTP payloads, persistence, source URIs/storage keys, ACL state, and provider-specific extension bags remain outside portable Core shapes.

Built-in semantic skills are versioned profiles. Assistant defines `token_economy/v1`; OCR discovery defines conditional `data_identification/v1`. Neither schema prescribes provider prompt text.

Model lists and free/paid availability are intentionally not frozen in schemas. Provider adapters supply current model catalogs at runtime.

Future schemas may cover stable error envelopes, streaming events after multi-provider validation, matching/solver contracts, and other genuinely cross-language serialized values.
