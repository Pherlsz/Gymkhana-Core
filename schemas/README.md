# Machine-readable schemas

Schemas define portable serialized contracts that can be consumed by multiple language implementations and generators.

They complement, but do not replace, the normative prose in `spec/` and the observable behavior in `conformance/`.

Use schemas for data contracts where generation is appropriate. Do not attempt to generate nontrivial algorithms solely to avoid writing idiomatic implementations in each language.

Current schema:

- `conformance.schema.json` — shape of shared conformance suites.

Future schemas may cover provider-neutral Assistant/OCR contracts, stable error envelopes, and other genuinely cross-language serialized values.
