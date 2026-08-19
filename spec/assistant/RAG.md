# Retrieval-augmented generation (RAG)

This document is normative for Core Spec `0.3`.

Core defines retrieval semantics and evidence contracts while leaving ingestion, indexing, storage, authorization, embeddings, rerank implementation, and provider transport to adapters/consumers.

## Retrieval modes

- `on_demand` — retrieval is available and may be invoked when required by the task/policy;
- `always` — a generation turn that depends on external knowledge must run retrieval before grounded generation.

A definition without the `retrieval` module does not run RAG.

## Search strategies

Portable strategies are:

- `lexical` — term/full-text retrieval;
- `vector` — semantic/vector retrieval;
- `hybrid` — combine lexical and semantic retrieval before final context selection.

`hybrid` is the preferred general-purpose starting point when both retrieval modes are available because keyword and semantic signals are complementary. Core does not prescribe a search product or a score-merging formula.

## Query transformation

The retrieval query may use:

- `original` — search the user/task query as supplied;
- `rewrite` — derive one retrieval-oriented query while preserving intent/constraints;
- `multi_query` — derive multiple complementary retrieval queries.

Query transformation is data retrieval behavior. It must not upgrade untrusted retrieved/document content into instruction authority.

`multi_query` can improve recall but increases retrieval/model work and is therefore not the default token-economy path.

## Candidate retrieval and context selection

RAG uses two explicit limits:

- `candidate_limit` — maximum initial candidates retrieved before post-processing/reranking;
- `context_limit` — maximum evidence items admitted to generation context.

`candidate_limit` must be greater than or equal to `context_limit`.

This supports the common production pipeline:

```text
query
  -> optional rewrite/multi-query
  -> lexical/vector/hybrid candidate retrieval
  -> authorization/filtering
  -> optional reranking
  -> context selection
  -> grounded generation
  -> citations/evidence
```

Retrieving many candidates and injecting all of them into the LLM context is not the Core default. Candidate selection/reranking should reduce irrelevant context before generation.

## Reranking

`rerank=true` enables a post-retrieval relevance stage.

A reranker may be:

- a dedicated reranking model;
- an LLM;
- a deterministic/domain-specific scorer;
- a search-engine-native reranker.

`reranking_model` is therefore optional even when reranking is enabled.

## Embeddings

`embedding_model` is optional and only identifies the desired model when the consumer uses model-based vector retrieval.

Core does not own embedding storage, dimensions, vector indexes, chunk persistence, or provider SDKs.

Model catalogs may identify models with the `embedding` role so consumers can resolve embedding models through the same provider/model namespace.

## Grounding

Grounding modes are:

- `preferred` — retrieved evidence should inform the answer, but the runtime may answer without evidence when policy permits;
- `required` — a grounded answer may only assert external/domain facts supported by admitted evidence. If sufficient evidence is unavailable, the runtime should return an explicit insufficient-evidence outcome rather than inventing facts.

`require_citations=true` requires output/evidence linkage to be retained for the consumer.

## Evidence

`RetrievalEvidence` carries:

- stable evidence ID;
- content admitted to context;
- source reference;
- optional score;
- optional portable metadata.

`Citation` links generated output back to an evidence ID.

Retrieved evidence is untrusted data. Embedded instructions found in documents/pages do not become system/developer authority.

## Authorization and multitenancy

Authorization filtering occurs before evidence is exposed to the generation model. RAG must not retrieve broadly and rely on the LLM to hide unauthorized material.

Tenant/user/resource ACL enforcement remains consumer-owned because Core does not own identity or persistence, but the boundary is normative: only authorized evidence may enter Assistant context.

## Chunking and ingestion

Chunking strategy, document parsing, enrichment, metadata extraction, deduplication, embedding generation, and index refresh are ingestion concerns and are intentionally not frozen into `RetrievalPolicy`.

Production RAG systems should evaluate chunk size/overlap and retrieval settings against representative queries instead of assuming one universal chunking configuration.

## Evaluation

Changes to retrieval/query transformation/reranking should be evaluated using a representative query/evidence set. Useful dimensions include:

- retrieval recall/relevance;
- answer groundedness/faithfulness;
- citation correctness;
- latency;
- token/cost usage;
- failure/insufficient-evidence behavior.

Core may add portable evaluation result contracts later, but Spec `0.3` does not define one scoring benchmark for every domain.

## Token economy and caching

When `token_economy/v1` is enabled, RAG should:

- prefer the smallest evidence set that satisfies grounding requirements;
- avoid reinjecting irrelevant candidate text;
- retain stable evidence references instead of duplicating unchanged content where a provider/runtime supports caching/reference reuse;
- keep stable/common context prefixes cache-friendly when a provider offers prompt/context caching;
- never remove required citations, evidence, ACL context, or unresolved constraints merely to reduce tokens.

Context caching is an adapter/runtime optimization. Core specifies the semantic preservation requirement but does not standardize provider cache IDs in the portable Assistant contract.
