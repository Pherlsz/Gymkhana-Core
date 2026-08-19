# Retrieval-augmented generation (RAG)

This document is normative for Core Spec `0.3`.

Core defines retrieval semantics/evidence contracts while ingestion, indexing, storage, authorization, embedding execution, rerank execution, and transport remain consumer/adapter concerns.

## Retrieval modes

- `on_demand` — retrieval is available and may be invoked when task/policy requires it;
- `always` — external/domain factual generation requires retrieval before grounded generation.

An Assistant that declares the retrieval module must carry a valid `RetrievalPolicy`.

## Search strategies

Portable strategies:

- `lexical`;
- `vector`;
- `hybrid`.

Hybrid retrieval is useful when lexical and semantic signals are complementary, but Core does not mandate one score-normalization or search product.

## Query transformation

- `original` — retrieve using the supplied task query;
- `rewrite` — derive one retrieval-oriented query while preserving intent/constraints;
- `multi_query` — derive multiple complementary queries.

Query transformation is retrieval behavior, not instruction authority. Untrusted user/document content cannot become system/developer instructions through rewriting.

`multi_query` increases recall opportunities and also work/token/latency cost; it is not the token-economy default.

## Candidate/context bounds

`candidate_limit` is the maximum initial result set before post-processing. `context_limit` is the maximum evidence-item count admitted to generation context.

Portable bounds in Spec `0.3`:

- `1 <= context_limit <= 1024`;
- `context_limit <= candidate_limit <= 10000`.

The intended pipeline is:

```text
query
  -> optional rewrite/multi-query
  -> lexical/vector/hybrid candidates
  -> authorization/filtering
  -> optional reranking
  -> bounded context selection
  -> grounded generation
  -> citation/evidence linkage
```

Retrieving a large candidate set does not imply forwarding the entire set into LLM context.

## Embedding and reranking models

`embedding_model` and `reranking_model` use the same `(provider, model)` namespace as generation models.

When a live catalog is available:

- `embedding_model` must advertise `embedding` role;
- `reranking_model` must advertise `reranking` role.

A `reranking_model` is invalid when `rerank=false`. `rerank=true` does not require a model because the consumer may use deterministic/search-native/domain-specific reranking.

Core owns neither vector dimensions nor embedding indexes.

## Grounding

Modes:

- `preferred` — evidence should inform the answer when available;
- `required` — unsupported external/domain assertions are not permitted. If sufficient evidence is unavailable, the runtime returns an insufficient-evidence outcome rather than inventing support.

`require_citations=true` requires the consumer to preserve output/evidence linkage.

## Evidence

`RetrievalEvidence` contains stable evidence ID, admitted content, source, optional finite score, and optional bounded metadata.

Portable evidence hardening includes:

- non-empty bounded IDs/source/content;
- content capped to the portable evidence byte limit;
- finite scores only — NaN/Infinity are invalid;
- bounded metadata entry count/key/value size;
- duplicate evidence IDs rejected inside a set.

`Citation` references an evidence ID rather than duplicating evidence content.

Retrieved evidence remains untrusted data. Instructions found inside documents/pages/OCR/tool output never inherit system/developer authority.

## Authorization and multitenancy

Authorization filtering occurs before evidence is exposed to the generation model. RAG must not retrieve unauthorized data broadly and ask an LLM to hide it afterwards.

Tenant/user/resource ACL enforcement remains consumer-owned because Core does not own identity/persistence.

## Chunking and ingestion

Chunk parsing, chunk size/overlap, enrichment, deduplication, metadata extraction, embedding refresh, and index lifecycle are intentionally not frozen into `RetrievalPolicy`.

These parameters should be evaluated against representative domain queries rather than assumed universal.

## Evaluation

Useful RAG evaluation dimensions include retrieval recall/relevance, groundedness/faithfulness, citation correctness, latency, token/cost usage, and insufficient-evidence behavior.

Core Spec `0.3` does not define one universal quality score. Task-scoped `SkillLearningProposal` may later reuse evidence-backed retrieval lessons without globally changing every Assistant.

## Token economy and caching

With `token_economy/v1`, RAG should prefer the smallest evidence set satisfying grounding, avoid injecting irrelevant candidates, and avoid duplicating stable context where runtime/provider reuse is safe.

Provider caching is not standardized in Spec `0.3`: cache semantics, TTL, minimum token thresholds, storage/retention, and privacy guarantees differ. A later context-planning/caching contract must preserve instruction hierarchy, evidence, ACL context, and unresolved constraints.
