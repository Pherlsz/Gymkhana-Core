# Machine-readable schemas

Schemas define portable serialized contracts that can be consumed by multiple language implementations and generators.

They complement, but do not replace, the normative prose in `spec/` and the observable behavior in `conformance/`.

Use schemas for data contracts where generation is appropriate. Do not attempt to generate nontrivial algorithms solely to avoid writing idiomatic implementations in each language.

Current schemas:

- `conformance.schema.json` — shape of shared conformance suites;
- `assistant.schema.json` — provider-neutral Assistant message, content, media-reference, tool, finish-reason, usage, and capability shapes.

The Assistant schema intentionally contains no provider SDK fields, credentials, model IDs, HTTP concerns, or provider-specific extension bag. Provider adapters keep such metadata outside the portable Core serialized shape.

Future schemas may cover OCR contracts, stable error envelopes, and other genuinely cross-language serialized values.
