# Machine-readable schemas

Schemas define portable serialized contracts that can be consumed by multiple language implementations and generators. They complement, but do not replace, the normative prose in `spec/` and observable behavior in `conformance/`.

Current schemas:

- `conformance.schema.json` — shape of shared conformance suites;
- `assistant.schema.json` — provider-neutral Assistant messages, multimodal content, tools, finish reasons, usage, capabilities, and built-in skills;
- `assistant-config.schema.json` — reusable Assistant definitions, modules, model catalogs/routing, credential/fallback policy, tool/RAG policy, and execution budgets;
- `assistant-runtime.schema.json` — per-run overrides plus normalized provider-adapter request/response metadata;
- `assistant-observability.schema.json` — provider quota observations, local usage attribution, and content-free run traces;
- `ocr.schema.json` — provider-neutral OCR/extraction requests, source/evidence references, observations, field candidates, review/validation state, and warnings.

`portable_json_schema/v1` is a semantic subset/profile defined under `spec/json/` and tested through the active `portablejson.json` conformance suite. It is not represented by pretending one ordinary JSON Schema can replace strict lexical, exact-number, resource-bound, and instance-validation requirements.

Likewise, `ocr.schema.json` describes serialized shape but does not replace semantic checks such as source-reference integrity, evidence modality, aggregate data bounds, candidate-to-structured-data equality, ambiguity invariants, discovery correctness, or validation against the exact caller target schema.

Provider SDK objects, raw credentials, HTTP payloads, persistence, source URIs/storage keys, ACL state, provider-specific extension bags, product memory, and learning/promotion state remain outside portable Core shapes.

Assistant defines the versioned `token_economy/v1` semantic skill. OCR discovery rules are direct extraction invariants, not a separate Core skill or provider prompt contract.

Model lists and free/paid availability are intentionally not frozen in schemas. Provider adapters supply current model catalogs at runtime.
