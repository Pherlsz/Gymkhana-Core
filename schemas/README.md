# Machine-readable schemas

Schemas define portable serialized contracts that can be consumed by multiple language implementations and generators.

They complement, but do not replace, the normative prose in `spec/` and the observable behavior in `conformance/`.

Use schemas for data contracts where generation is appropriate. Do not attempt to generate nontrivial algorithms solely to avoid writing idiomatic implementations in each language.

Current schemas:

- `conformance.schema.json` — shape of shared conformance suites;
- `assistant.schema.json` — provider-neutral Assistant messages, multimodal content, tools, finish reasons, usage, capabilities, and built-in skills;
- `assistant-config.schema.json` — reusable Assistant definitions, modules, model catalogs/routing, credential/fallback policy, tool/RAG/memory/learning policy, and execution budgets;
- `assistant-runtime.schema.json` — per-run overrides plus normalized provider-adapter request/response metadata;
- `assistant-observability.schema.json` — provider quota observations, local usage attribution, content-free run traces, task signatures, and learning proposals.

`portable_json_schema/v1` is itself a semantic subset/profile defined under `spec/json/` and tested through `conformance/v0.3/portablejson.json`. It is not represented by pretending that one ordinary JSON Schema document can replace its strict lexical, exact-number, resource-bound, and instance-validation requirements.

Provider SDK objects, raw credentials, HTTP payloads, persistence, and provider-specific extension bags remain outside the portable Core shapes. `CredentialRef.reference` is an opaque runtime secret-store handle, never raw API-key material; persistent observability uses `CredentialIdentity`, which omits that handle.

Built-in Assistant skills are versioned semantic profiles. The schema currently defines `token_economy/v1`; it does not prescribe provider prompt text or a provider-specific tokenization strategy.

Model lists and free/paid availability are intentionally not frozen in schemas. Provider adapters supply current `ModelDescriptor` catalogs at runtime.

Future schemas may cover OCR contracts, stable error envelopes, streaming events after multi-provider validation, and other genuinely cross-language serialized values.
