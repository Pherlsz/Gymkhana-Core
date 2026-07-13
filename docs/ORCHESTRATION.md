# Gymkhana Core — Orchestration Document

> **Planning version:** Stage 9  
> **Last synchronized:** 2026-07-13  
> **Current stage:** Stage 9 completed  
> **Primary source of truth:** `Pherlsz/Gymkhana-Database/docs/ORCHESTRATION.md`  
> **Repository responsibility:** reusable Go logic independent from infrastructure

This document records approved decisions that affect `Gymkhana-Core`. Product, persistence, HTTP, OpenAPI, provider, worker, and UI responsibilities remain in Gymkhana Database and Gymkhana UI.

## 1. Mission

Provide a small, deterministic, testable Go module for reusable logic that does not depend on PostgreSQL, HTTP, React, River, R2, Cloud Run, Vercel, or AI-provider SDKs.

Expected public areas:

```text
normalize/
civiltime/
query/
matching/
duplicates/
assistant/
tools/
result/
internal/
```

Do not create broad packages named `utils`, `helpers`, `common`, `shared`, or `core`.

## 2. Engineering language

Source code, identifiers, comments, documentation, commits, PRs, changelog, releases, workflows, and CI messages are written in English.

Core does not contain product-facing Portuguese strings. Consumers own localization.

## 3. What does not belong in Core

- PostgreSQL repositories, SQL, pgx, sqlc, migrations, indexes, RLS, or transaction management;
- HTTP handlers, status codes, cookies, sessions, CSRF, routing, or OpenAPI-generated product types;
- product roles, permissions, authorization service, field visibility, or resource ownership;
- River, R2/S3, signed URLs, file downloads, PDF rendering, Cloud Run, Neon, or Vercel;
- OpenAI or Google SDK types;
- product prompts, provider secrets, model assignments, quotas, or prices;
- React, TypeScript, Data Grid, routes, Query Keys, cache invalidation, forms, or responsive rules;
- complete Profile/document/bill persistence models;
- duplicate queue persistence, merge transaction, lixeira, restore, or audit storage;
- OCR operations, attachment access, or application of suggestions;
- imports, XLSX parsing, Forms synchronization, exports, notifications, or operation polling;
- direct execution of arbitrary code or SQL.

## 4. Principles

- Standard library first.
- Prefer pure functions and deterministic outputs.
- Add interfaces only at real substitution boundaries.
- Keep public APIs intentionally small.
- Avoid premature extraction from Gymkhana Database.
- Do not expose external dependency types in public contracts.
- Never interpret a score as probability without a validated probabilistic model.
- Return structured reasons and warnings.
- Describe intent and plans; leave data access and authorization to the consumer.
- Support broad capability through typed extensibility, not arbitrary scripts.
- Do not rely on exact user keywords as a semantic contract.
- Treat HTTP, OpenAPI, persistence, idempotency storage, and localization as adapters outside Core.

## 5. Stack and quality

- Go 1.26;
- Go modules;
- `testing` and native fuzzing;
- race detector where concurrency exists;
- `gofmt`, `go vet`, `staticcheck`, `govulncheck`;
- OSV-Scanner and dependency review;
- exact dependency versions and committed `go.sum`;
- GitHub Actions pinned by commit SHA.

No assertion framework, ORM, application framework, generic agent framework, or scripting runtime initially.

## 6. Module and consumption

```go
module github.com/Pherlsz/Gymkhana-Core
```

Gymkhana Database pins an exact tag and configures:

```text
GOPRIVATE=github.com/Pherlsz/Gymkhana-Core
```

No permanent `replace`, branch, untagged commit, submodule, subtree, or copied source in production.

A local uncommitted parent `go.work` may connect Core and Database.

## 7. Normalization and temporal types

Potential reusable normalization:

- names and search text;
- Unicode case, accents, spaces, and punctuation;
- CPF and official identifiers;
- phone and email;
- alphanumeric document values;
- custom option keys;
- address fragments when generic enough.

Rules:

- preserve leading zeros;
- identifiers remain text;
- uppercase canonical letters when applicable;
- remove display separators only in canonical forms;
- lowercase email;
- keep display value separate from normalized search value;
- version normalizers when changes affect persisted fingerprints, matching, or search.

Temporal separation:

- `CivilDate`: year, month, day, no timezone;
- `YearMonth`: year and month;
- instant: consumer-owned UTC `time.Time`.

Core may provide parsing, validation, comparison, serialization, leap-year behavior, month navigation, and deterministic fingerprint helpers. It does not own locale formatting, HTTP representation, or persistence.

## 8. Query catalog and plans

The consumer supplies a permission-filtered catalog. Core validates against it but never loads it.

Catalog concepts:

- entities;
- fields;
- sources;
- value types;
- relations and cardinalities;
- capabilities;
- operators;
- aggregations;
- aliases;
- sensitivity metadata;
- schema version and fingerprint.

Neutral value types include text, long text, identifier, integer, decimal, money, boolean, civil date, year month, instant, enum, multi-enum, email, phone, URL, UUID, and attachment reference.

Field paths are structured and never contain arbitrary table or column names. Relation predicates support `any`, `all`, and `none`.

Filter AST nodes:

```text
Predicate
RelationPredicate
And
Or
Not
```

`QueryPlan` contains version, catalog fingerprint, root, projection, filter, sort, grouping, aggregations, distinct, page/limit, and result mode.

Result modes:

```text
records
references
count
aggregation
groups
exists
```

Core validates structure, catalog, types, operators, relation compatibility, and neutral complexity budgets. Authorization and physical execution remain in Database.

## 9. ExecutionPlan, sets, patterns, and combinations

Complex work uses a typed DAG:

```text
query
transform
set_operation
combination
rank
summarize
```

Set operations:

```text
union
intersection
difference
symmetric_difference
```

Result grain identifies Profile, document, bill, custom record, or another entity and prevents join multiplicity from changing semantic counts.

Character selection distinguishes:

```text
contiguous
ordered_subsequence
unordered_subset
permutation
```

Transforms may include binary-to-integer/byte, digits-to-integer, characters-to-text, and counts. Predicates may test letter, digit, alphanumeric, printable, exact character, range, or set membership.

Each position is used at most once by default. Algorithms use counts, pruning, dynamic programming, bounded permutation, branch-and-bound, and bounded memoization instead of blind enumeration.

Combination plans support cardinality, distinctness, owner/entity bindings, constraints, ranking, and stable tie-breakers.

## 10. Canonical serialization, fingerprints, and results

Plans use versioned canonical JSON:

- stable ordering where semantics allow;
- canonical dates and numbers;
- technical keys, never labels;
- no SQL, credentials, provider types, auth data, or locale strings.

Fingerprints may combine canonical plan with catalog and permission-scope inputs supplied by the consumer.

Neutral results can represent columns, rows, groups, totals, references, and warnings with typed values.

Core may generate structured operational explanations of roots, filters, relation quantifiers, patterns, sets, bindings, grouping, aggregation, ranking, and result mode. It never exposes SQL or private reasoning.

## 11. Assistant and tool contracts

Core supports semantic planning without a keyword-only parser. Equivalent phrasing, synonyms, sentence order, colloquial language, abbreviations, and minor errors may map to equivalent typed plans.

No provider SDK type appears in public APIs.

Neutral provider contracts may expose messages, structured output, tools, usage, finish status, streaming events, capabilities, and cancellation.

Potential tools:

- discover catalog;
- validate query plan;
- execute query plan;
- start query operation;
- get operation status/result;
- get record details.

Never define `execute_sql`, `run_code`, `list_tables`, or direct database access.

Invalid calls are rejected completely. Core represents correction attempts and errors but does not create unbounded autonomous loops.

Structured Assistant results may contain text, explanation, references, tables, warnings, operation references, and declarative actions.

Core may represent run/tool/event contracts but does not know SSE, sessions, thread ownership, persistence, quotas, or provider credentials.

## 12. OCR contracts

Core may define:

- document classification suggestions;
- extraction schemas;
- `SuggestedValue`;
- suggestion status and warnings;
- evidence references;
- normalized bounding boxes;
- clarity levels;
- current-versus-suggested comparison;
- human decision enums;
- deterministic merge/validation helpers;
- neutral OCR fingerprint inputs.

Core does not call providers, inspect R2, render PDFs, persist operations, authorize users, handle resource versions, or apply accepted values.

Evidence may contain attachment reference, page, normalized box, and limited snippet. Core validates ranges and structure but never invents coordinates.

## 13. Matching and duplicate assessment

Core receives normalized values or compact evidence vectors and returns deterministic assessment.

Levels:

```text
VERY_STRONG
PROBABLE
POSSIBLE
```

Evidence direction:

```text
supports_duplicate
supports_distinct
neutral
```

Evidence includes rule key, field keys, strength, and message key. Scores are never exposed as probability.

Name alone cannot produce a strong result. Strong identifiers reject placeholders and invalid canonical values.

Same input, configuration, rule version, and normalization version must produce the same output.

Core accepts compact evidence vectors rather than complete Profiles and relations. Candidate generation, SQL evidence, staging, pagination, egress measurement, queue persistence, review, and merge remain in Database.

## 14. Complex Gymkhana task contracts

A task plan may combine people, documents, bills, addresses, custom data, character patterns/transforms, combinations, evidence, and external requirements.

Bindings:

```text
same_entity
different_entities
same_owner
distinct_records
any_member
whole_combination
```

Requirement status:

```text
satisfied
partially_satisfied
not_satisfied
not_verifiable
ambiguous
```

Core may define versioned dictionaries and matching helpers, task plans, requirements, constraints, objectives, warnings, solutions, evidence, and alternatives.

The solver starts with the smallest candidate set, propagates constraints, prunes early, respects owner/distinct bindings, limits solutions, and produces deterministic materially different alternatives.

Database owns candidate fetching, asynchronous execution, persistence, checkpoints, budgets, retention, and user-facing minimization.

## 15. Imports and neutral operation concepts

Core may provide pure functions for header normalization, typed conversion, schema validation, row errors, mapping transformation definitions, candidate features, and canonical row fingerprints when generic enough.

Allowed transformations may include trim, whitespace normalization, case conversion, mask removal, typed date/year-month/decimal parsing, option mapping, split, join, and constants.

Core does not import Excelize, create staging, interact with Google Forms, run batches, or report jobs.

Neutral operation values may represent queued, running, waiting for review, completed, completed with errors, failed, cancel requested, cancelled, expired, determinate/indeterminate progress, and stage metadata. Database owns operation resources and infrastructure.

## 16. Errors and transport boundary

Neutral errors use stable codes, paths, and arguments. They contain no HTTP status, localized message, SQL, stack trace, request ID, or provider payload.

Core public APIs do not depend on `/api/v1`, OpenAPI operation IDs, generated DTOs, session/role models, CSRF/idempotency headers, Query Keys, route schemas, or localization.

Database maps neutral contracts to API schemas and revalidates authorization, permissions, row scope, and physical execution.

Core may produce canonical request/fingerprint inputs but never stores idempotency records or repeats mutations.

## 17. Stage 9 implementation strategy

Core is implemented incrementally and released only when Gymkhana Database has a real consumer.

Do not extract speculative logic. Preferred flow:

```text
local Database implementation
→ proven reuse/stability
→ neutral Core contract
→ Core implementation and release
→ Database adoption
```

Exceptions that belong in Core from the beginning because their neutral boundary is already stable:

- normalization;
- CivilDate and YearMonth;
- canonical serialization/fingerprints;
- Query catalog/AST/plan contracts;
- deterministic matching contracts;
- provider/tool/OCR contracts once their milestone begins;
- ExecutionPlan and solver algorithms once Query Engine foundations are stable.

Core forms one of five implementation tracks: Core, UI, Backend/Database, Frontend/Product, and Infrastructure/Quality.

## 18. Core milestone roadmap

### M0 — Module and release foundation

Work packages:

```text
C0.1 module bootstrap
C0.2 quality commands
C0.3 CI
C0.4 release skeleton
```

Deliver Go module, intentionally small package tree, tests, static analysis, security checks, CHANGELOG, and manual release workflow.

Conceptual release: `v0.1.0`.

### M1 — Normalization, civil time, and fingerprints

Work packages:

```text
C1.1 text normalization
C1.2 identifiers
C1.3 versioned normalization
C1.4 CivilDate
C1.5 YearMonth
C1.6 neutral value errors
C1.7 canonical fingerprints
```

Require fuzz tests, examples, deterministic test vectors, and no database dependency.

Conceptual release: `v0.2.0`.

### M2–M9 — Consumer foundations

Core should not accumulate product-specific auth, Profile CRUD, document/bill persistence, custom schema storage, attachments, Search SQL, operations, imports, Forms, or exports.

It may receive small additional neutral parsing/normalization helpers only when proven reusable and stable.

### M10 — Query Engine base

Work packages:

```text
C10.1 value types
C10.2 catalog definitions
C10.3 relations
C10.4 catalog fingerprint
C10.5 operators
C10.6 Filter AST
C10.7 QueryPlan
C10.8 validation
C10.9 canonical serialization
C10.10 result contracts
C10.11 explanations
```

Gates include round-trip tests, fuzzing of arbitrary ASTs, budget boundaries, examples, and compatibility review.

Conceptual release: `v0.3.0`.

### M11 — Matching and duplicate assessment

Work packages:

```text
C11.1 ProfilePair
C11.2 evidence
C11.3 levels
C11.4 assessment
C11.5 deterministic rules
C11.6 compact evidence vector
```

Gates include determinism, invalid-identifier handling, contradictions, fuzzing, test vectors, and rule benchmarks.

Conceptual release: `v0.4.0`.

### M12 — Assistant/provider/tool contracts

Work packages:

```text
C12.1 provider request/result
C12.2 capabilities
C12.3 events
C12.4 tool schemas
C12.5 structured Assistant result
```

No provider SDK or transport dependency. Schemas are closed and versioned.

Conceptual release: `v0.5.0`.

### M13 — OCR contracts

Work packages:

```text
C13.1 SuggestedValue
C13.2 evidence
C13.3 comparison
C13.4 review decisions
```

Gates include malformed-output fuzzing, normalized-box validation, comparison determinism, and no provider/storage dependency.

Conceptual release: `v0.6.0`.

### M14 — ExecutionPlan, character engine, and solver

Work packages:

```text
C14.1 DAG
C14.2 step kinds
C14.3 result references
C14.4 CharacterSelection
C14.5 transforms
C14.6 predicates
C14.7 bounded algorithms
C14.8 combination slots
C14.9 constraints
C14.10 solver
C14.11 ranking
C14.12 task contracts
```

Gates include exhaustive small-domain tests, property tests, fuzzing, deterministic output, budget tests, no unbounded combinatorial execution, and performance benchmarks.

Conceptual release: `v0.7.0`.

### M15 — Hardening

- API compatibility review;
- benchmark/fuzz/race review;
- vulnerability scan;
- consumer upgrade validation;
- no large new public API family.

## 19. Conceptual release sequence

```text
v0.1.0 — module and release foundation
v0.2.0 — normalization, civil dates and fingerprints
v0.3.0 — Query catalog, AST and QueryPlan
v0.4.0 — matching and duplicate assessment
v0.5.0 — Assistant/provider/tool contracts
v0.6.0 — OCR contracts
v0.7.0 — ExecutionPlan, character engine and solver
```

Exact versions follow actual SemVer changes. Before `v1.0.0`, breaking changes remain explicit and require CHANGELOG, migration notes, coordinated Database update, compatibility tests, and no silent removal.

Release only when:

- a real Database consumer exists;
- contract is stable enough;
- tests/fuzz/benchmarks pass;
- examples and CHANGELOG are current;
- no infrastructure dependency leaked;
- Database is ready to update.

## 20. Dependencies and parallel work

Core and UI are independent and may progress in parallel.

Critical order for Core consumption:

```text
M0 foundation
→ M1 normalization/time
→ M10 Query contracts
→ M11 matching
→ M12 Assistant contracts
→ M13 OCR contracts
→ M14 advanced execution/solver
```

M10 must precede final AI Chat integration. M12 provider/tool contracts may proceed while M11 Database duplicate work continues. M14 algorithms may begin after Query contracts are stable, but product integration waits for Query Engine, operations, result sets, and Assistant orchestration.

Database must not implement final adapters against untagged or unstable Core APIs. Local `go.work` is allowed during coordinated development, but merge uses an exact released tag.

## 21. Work packages, issues, and PRs

Use milestone → epic → work-package issues.

Each Core issue records:

- context and neutral-use case;
- scope/out-of-scope;
- public API impact;
- dependencies and consumer;
- determinism requirements;
- acceptance criteria;
- unit/fuzz/property/benchmark plan;
- limits and failure behavior;
- documentation and changelog;
- infrastructure-independence check.

Prefer reviewable PRs per contract/algorithm family, not one milestone-sized PR. A small coherent vertical package may combine API, implementation, tests, examples, and docs.

Normally 1–3 intentional commits per PR, followed by squash merge.

A Core PR must state public API changes, determinism, tests/fuzz, benchmark when algorithmic, dependency impact, infrastructure independence, changelog entry, and Database consumer impact.

## 22. CI and release gates

CI:

```text
gofmt check
go vet ./...
staticcheck ./...
go test ./...
govulncheck ./...
OSV scan
```

Race runs on main/releases/concurrency-relevant changes and justified periodic checks.

Algorithm releases additionally require:

- deterministic test vectors;
- fuzz/property tests;
- budget/limit tests;
- benchmark evidence;
- examples;
- public API compatibility review;
- successful Database integration branch or consuming PR.

A Core release does not automatically update or deploy Database.

## 23. Definition of Done

A Core work package is done when scope/acceptance are met, API is reviewed, implementation is deterministic, tests/docs are complete, no critical TODO remains, and the next dependent package can consume it.

A Core algorithm is done when it has:

- stable neutral contract;
- deterministic output;
- structured errors/warnings;
- unit tests;
- fuzz/property tests;
- benchmark where performance matters;
- explicit limits and cancellation/budget representation where applicable;
- no infrastructure dependency;
- documentation/examples;
- release consumed by Database when required by a milestone.

A Core milestone is done when all required work packages are complete, CI/security checks are green, benchmark/fuzz gates pass, CHANGELOG is current, exact tag exists, Database adoption is validated, and residual risks are recorded.

## 24. Performance and safety policy

- Do not optimize before measuring, except to prevent clearly unbounded algorithms.
- Character and combination algorithms require explicit budgets.
- Prefer counts, pruning, dynamic programming, and branch-and-bound over enumeration.
- Avoid copying complete product records into neutral algorithms.
- Keep result/reference structures compact.
- Never expose sensitive raw values in ordinary logs or fingerprints when avoidable.
- No arbitrary script, SQL, expression evaluator, or provider instruction execution.

## 25. Deferred decisions

- physical SQL compiler;
- final provider/model choices;
- product prompt content;
- natural-language model implementation;
- arbitrary code or SQL tools;
- public distribution;
- automatic cross-repository updates;
- Changesets;
- final optimizations before benchmarks;
- HTTP/OpenAPI contracts in Core;
- authorization and role models in Core;
- React/Data Grid/mobile implementation in Core;
- speculative packages without a real Database consumer;
- fixed milestone dates before measuring early implementation velocity.

## 26. Next stage

**Stage 10 — implementation-start preparation.**

Core planning must define:

- exact initial module/package tree;
- `go.mod` and tool-version policy;
- commands and CI workflow files;
- first public APIs, tests, examples, and docs;
- release workflow and CHANGELOG skeleton;
- M0/M1 issue and PR sequence;
- local `go.work` integration instructions;
- checklist for publishing `v0.1.0` and `v0.2.0`.

Update this document again when Stage 10 is completed.