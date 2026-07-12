# Gymkhana Core — Orchestration Document

> **Planning version:** Stage 7  
> **Last synchronized:** 2026-07-12  
> **Current stage:** Stage 7 completed  
> **Primary source of truth:** `Pherlsz/Gymkhana-Database/docs/ORCHESTRATION.md`  
> **Repository responsibility:** reusable Go logic independent from infrastructure

This document records approved decisions that affect `Gymkhana-Core`. Product, persistence, HTTP, provider, and UI responsibilities remain in Gymkhana Database and Gymkhana UI.

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

- PostgreSQL repositories, SQL, pgx, sqlc, migrations, or indexes;
- HTTP handlers, sessions, product permissions, OpenAPI-generated product types;
- River, R2/S3, signed URLs, file downloads, PDF rendering, Cloud Run, Neon, or Vercel;
- OpenAI or Google SDK types;
- product prompts or model assignments;
- React, TypeScript, Data Grid, routes, UI copy, focus behavior, or responsive rules;
- complete Profile/document/bill persistence models;
- duplicate queue persistence or merge transactions;
- OCR operations, attachment access, or application of suggestions;
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

## 7. Normalization

Potential reusable normalization:

- names and search text;
- Unicode case, accents, spaces, and punctuation;
- CPF and official identifiers;
- phone and email;
- alphanumeric document values;
- custom option keys;
- civil dates and year-month;
- address fragments when generic enough.

Rules:

- preserve leading zeros;
- identifiers remain text;
- uppercase canonical letters when applicable;
- remove display separators only in canonical identifier forms;
- lowercase email;
- keep display value separate from normalized search value;
- keep `CivilDate`, `YearMonth`, and real instants distinct.

Core does not decide database representation or UI mask.

## 8. Temporal types

Conceptual public separation:

- `CivilDate`: year, month, day, no timezone;
- `YearMonth`: year and month;
- instant: consumer-owned UTC `time.Time`.

Core may provide parsing, validation, comparison, serialization, leap-year behavior, and month navigation.

It does not own locale formatting, date-picker behavior, or application timezone.

## 9. Query catalog contracts

The consumer supplies a filtered catalog. Core validates against it but never loads it.

Conceptual definitions:

```go
type EntityDefinition struct {
    Key          EntityKey
    Label        string
    Kind         EntityKind
    Fields       []FieldDefinition
    Relations    []RelationDefinition
    Capabilities EntityCapabilities
}

type FieldDefinition struct {
    Key          FieldKey
    Label        string
    Type         ValueType
    Source       FieldSource
    Nullable     bool
    Searchable   bool
    Filterable   bool
    Sortable     bool
    Groupable    bool
    Operators    []Operator
    Aggregations []Aggregation
    Sensitive    bool
}
```

Technical keys are stable. Labels and aliases are descriptive only.

Catalog metadata includes schema version and fingerprint. The consumer incorporates custom schema and permission scope into the supplied catalog/fingerprint.

## 10. Value types

Initial neutral value types:

```text
text
long_text
identifier
integer
decimal
money
boolean
civil_date
year_month
instant
enum
multi_enum
email
phone
url
uuid
attachment_reference
```

`identifier` is not numeric and preserves leading zeros and letters.

Typed values must have canonical JSON representation suitable for plan serialization and fingerprints.

## 11. Field paths and relations

A field reference is structured:

```go
type FieldPath struct {
    Root      EntityKey
    Relations []RelationKey
    Field     FieldKey
}
```

Relations declare source, destination, label, and cardinality:

```text
one_to_one
many_to_one
one_to_many
many_to_many
```

Every path is validated against the catalog. No arbitrary traversal or physical table/column name appears in Core.

Relation predicates support:

```text
any
all
none
```

## 12. Operators

Representative operators:

### Text

- equals/not equals;
- contains/not contains;
- starts/ends with;
- empty/not empty;
- normalized equals/contains;
- fuzzy match.

### Identifier

- equality;
- contains/prefix/suffix;
- empty;
- pattern match;
- allowed/required character classes;
- length equality/range;
- character-computation predicates.

### Numeric/money

- equality and inequality;
- comparison;
- between;
- empty.

### Dates/year-month/instant

- equality;
- before/after;
- between;
- empty.

### Boolean

- true/false/empty.

### Enum/relations

- is/is not;
- any/none of;
- empty/not empty.

Catalog definitions decide which operators are exposed per field.

## 13. Filter AST

Filter nodes:

```text
Predicate
RelationPredicate
And
Or
Not
```

The AST supports complete logical composition even if the first product filter UI exposes only a flat `AND` list.

Validation covers node kind, depth, required values, operator compatibility, relation path, and budget.

## 14. Query Plan

Conceptual contract:

```go
type QueryPlan struct {
    Version            string
    CatalogFingerprint string
    Root               EntityKey
    Projection         []Projection
    Filter             FilterNode
    Sort               []Sort
    Grouping           []Group
    Aggregations       []AggregationSpec
    Distinct           bool
    Page               *Pagination
    Limit              *int
    ResultMode         ResultMode
}
```

Result modes:

```text
records
references
count
aggregation
groups
exists
```

Projection requests only necessary fields. The consumer may add hidden technical identifiers needed for execution.

Sorting declares field, direction, and null placement. Grouping and aggregation are typed. Initial aggregations:

```text
count
count_distinct
sum
average
minimum
maximum
```

## 15. Execution Plan

Complex work uses a DAG of typed steps:

```go
type ExecutionPlan struct {
    Version            string
    CatalogFingerprint string
    Steps              []ExecutionStep
    Output             OutputDefinition
}
```

Step kinds:

```text
query
transform
set_operation
combination
rank
summarize
```

Dependencies reference structured step outputs. Cycles and unknown dependencies are invalid. Independent steps may be executed in parallel by the consumer.

Core does not persist intermediate results or choose SQL/temp-table strategy.

## 16. Set operations and result grain

Set operations:

```text
union
intersection
difference
symmetric_difference
```

Each set declares a stable key and result grain.

`ResultGrain` identifies whether a result represents a Profile, document, bill, custom record, or another entity. This prevents join multiplicity from changing semantic counts.

Distinctness is explicit. Core does not add `DISTINCT` to hide an incorrect plan.

## 17. Character patterns and computation

Core distinguishes:

```text
contiguous
ordered_subsequence
unordered_subset
permutation
```

Conceptual contracts:

```go
type CharacterSelection struct {
    Classes     []CharacterClass
    Literals    []rune
    Cardinality CardinalityConstraint
    Mode        CharacterSelectionMode
}

type CharacterTransform struct {
    Kind     CharacterTransformKind
    BitOrder BitOrder
    Target   TransformTarget
}
```

Character classes may include digit, letter, uppercase/lowercase, binary digit, whitespace, symbol, and alphanumeric.

Transformations may include binary-to-integer/byte, digits-to-integer, characters-to-text, and counts.

Result predicates may test letter, digit, alphanumeric, printable, exact character, integer range, or membership.

Mandatory use case: choose exactly eight binary digits from a CPF or another identifier, not necessarily contiguous or ordered, optionally permute them, and test whether they can form a letter or number.

Each position is used at most once by default.

Core should use counts, pruning, dynamic programming, and bounded search rather than blindly generating all permutations.

## 18. Combination planning

Conceptual contract:

```go
type CombinationPlan struct {
    Input       StepResultReference
    GroupSize   CardinalityConstraint
    DistinctBy  []FieldPath
    Constraints []CombinationConstraint
    Objective   *RankPlan
}
```

A combination may involve multiple records and owners. Distinctness may apply by entity ID, document ID, Profile owner, or another stable field path.

Typed transforms include normalize, extract/select/transform characters, map, filter, project, deduplicate, group, and aggregate.

No JavaScript, Python, expression evaluator, or arbitrary script execution.

## 19. Ranking

Ranking is structured and explainable:

```go
type RankPlan struct {
    Input     StepResultReference
    Criteria  []RankCriterion
    Direction RankDirection
    Limit     int
}
```

Weights are decimal and criteria are visible in explanations. Scores are not probabilities.

Stable tie-breakers are required.

## 20. Validation and budgets

Validation layers performed by Core and consumer together:

1. structural;
2. catalog;
3. value/type/operator;
4. logical and relation compatibility;
5. complexity budget.

Authorization and operational cost remain consumer responsibilities, though Core can represent neutral capability metadata and errors.

Neutral errors include:

```text
unknown_entity
unknown_field
unknown_relation
unsupported_operator
invalid_value
invalid_field_path
invalid_aggregation
catalog_changed
query_too_complex
result_too_large
post_processing_required
```

Budget fields may include relation depth, logical nodes, projection items, groups, aggregations, character selections, combination width, candidate values, sync rows, and execution time.

A budget result may advise asynchronous execution rather than declaring the user request functionally unsupported.

## 21. Canonical serialization and fingerprints

Query and Execution Plans use versioned canonical JSON:

- stable ordering where semantics allow;
- canonical dates and numbers;
- technical keys, never display labels;
- no SQL, credentials, provider types, or authentication data.

Fingerprints are derived from canonical plan plus catalog and permission-scope inputs supplied by the consumer.

Uses include all-matching selection, job identity, audit references, result-set follow-ups, idempotency, and future caching.

## 22. Query explanations

Core may produce structured explanation parts from a validated plan.

Explanations describe:

- root entity;
- filters and relation quantifiers;
- selection/permutation semantics;
- sets and bindings;
- grouping, aggregation, and ranking;
- result mode.

The consumer localizes labels and messages.

No SQL or private chain of thought is exposed.

## 23. Neutral result contracts

Conceptual result types:

```go
type QueryResult struct {
    Mode       ResultMode
    Columns    []ResultColumn
    Rows       []ResultRow
    Groups     []ResultGroup
    Total      *int64
    HasMore    bool
    References []ResultReference
    Warnings   []QueryWarning
}
```

Values remain typed.

References contain entity, ID, readable label metadata, and optional field path. The product creates URLs and checks visibility.

Core may represent summaries and samples but does not fetch or persist records.

## 24. Semantic Assistant contracts

Core must support natural-language planning without encoding a keyword-only parser.

Different phrasing, synonyms, sentence order, colloquial language, abbreviations, and minor typing errors may represent the same intent.

Exact words may be useful aliases in the catalog but are never the sole semantic rule.

The AI provider or product orchestration performs language interpretation. Core provides typed targets, validation, correction errors, and operational explanations.

A plan explanation allows the user to verify what was interpreted without exposing internal reasoning.

## 25. Assistant catalog and discovery

Neutral contracts may represent:

- assistant-visible entities;
- fields and aliases;
- types and allowed operators;
- relations;
- capabilities;
- examples;
- catalog fingerprint.

Discovery tools may include list/describe/search entities, fields, relations, and operators.

The product supplies a permission-filtered contextual catalog and decides expansion strategy.

Core never includes physical schema names or data samples by default.

## 26. Neutral AI provider contracts

No provider SDK type appears in public APIs.

Conceptual interfaces:

```go
type AIProvider interface {
    Generate(ctx context.Context, request GenerateRequest, sink EventSink) (GenerateResult, error)
}
```

Capabilities may declare streaming, structured output, tools, vision, prompt caching, and native cancellation.

The consumer maps OpenAI, Google, or future providers.

Core does not select configured models, prices, credentials, or fallback policy.

## 27. Tool contracts

Tools are small, explicit, typed, and versioned.

Potential neutral tools:

- discover catalog;
- validate query plan;
- execute query plan;
- start query operation;
- get operation status/result;
- get record details.

Never define `execute_sql`, `run_code`, `list_tables`, or direct database access.

Each tool declares stable name, description, input/output schema, limits, and neutral permission metadata. The consumer performs actual authorization.

Invalid calls are rejected completely, not partially executed.

## 28. Plan correction

Structured validation errors may be returned to the AI for correction.

Initial product policy allows up to two plan corrections, but Core only represents attempts and errors; it does not create autonomous unbounded loops.

No invalid partial plan may execute.

## 29. Assistant messages, runs, and events

Core may define neutral contracts for:

- structured response text;
- explanation;
- references;
- tables;
- warnings;
- operation references;
- declarative actions;
- run/attempt/tool-call states;
- streaming events.

Neutral events may include:

```text
run.started
run.status
plan.ready
plan.warning
tool.started
tool.completed
operation.started
text.delta
references.ready
table.ready
message.completed
run.failed
run.cancelled
heartbeat
```

Core does not know SSE or HTTP; Database maps events to transport.

## 30. Conversation context references

Core may represent structured thread summaries, active constraints, referenced records, open questions, result-set IDs, query fingerprints, result grain, and operation IDs.

This supports follow-ups such as “from those” without relying only on text.

Core does not persist threads, decide retention, or create global user memory.

## 31. Prompt-injection boundary

Core treats user text, stored database content, PDFs, OCR text, labels, and examples as untrusted data.

Security is structural:

- typed tools;
- validated plans;
- catalog scope;
- no arbitrary execution;
- no provider credentials or SQL;
- closed output schemas;
- consumer authorization.

Core must never convert record text into instructions or grant capabilities based on content.

## 32. OCR neutral contracts

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

Suggested statuses:

```text
found
not_found
unclear
conflicting
invalid
not_applicable
```

Clarity:

```text
clear
review_recommended
unclear
conflicting
```

Comparison:

```text
empty_current
same
different
current_invalid
suggestion_invalid
conflict_between_sources
```

Core does not call providers, inspect R2, render PDFs, persist operations, authorize users, or apply accepted values.

## 33. OCR evidence

Evidence may contain attachment reference, page, normalized box, and limited snippet.

Core validates numeric ranges and structural consistency but never invents coordinates.

Structured output is preferred. Raw provider responses and full document transcription are not Core contracts.

## 34. Matching and duplicate assessment

Core receives normalized values or compact evidence vectors and returns deterministic assessment.

Visible levels:

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

Evidence includes rule key, field keys, strength, and message key. Internal scores may order candidates but are never exposed as probability.

Name alone cannot produce a strong duplicate result. Strong identifiers must reject placeholders and invalid canonical values.

## 35. Duplicate rules and fingerprints

Conceptual contracts:

```go
type DuplicateRule struct {
    Key           string
    Version       string
    Required      []Condition
    Supporting    []Condition
    Contradicting []Condition
    ResultLevel   DuplicateLevel
}

type DuplicateAssessment struct {
    Pair                  ProfilePair
    Level                 DuplicateLevel
    SupportingEvidence    []DuplicateEvidence
    ContradictingEvidence []DuplicateEvidence
    RuleVersion           string
    NormalizationVersion  string
}
```

Same input, configuration, rule version, and normalization version must produce the same output.

Core may compute evidence fingerprints without embedding reversible sensitive values in ordinary logs.

## 36. Egress-aware duplicate boundary

Core must accept compact evidence vectors instead of requiring complete Profiles and all relations.

Example vector contents:

- canonical pair IDs;
- exact-match flags;
- limited similarity values;
- contradiction flags;
- evidence codes;
- fingerprint.

It must not require observations, full addresses, all documents, all bills, attachments, or custom records for every pair.

Candidate generation, SQL evidence calculation, staging, pagination, aggregation, and Neon egress measurement remain in Gymkhana Database.

This boundary allows PostgreSQL to reduce 10k–20k candidate pairs before application processing.

## 37. Complex Gymkhana task contracts

A task plan may combine multiple categories simultaneously:

- people;
- documents;
- bills;
- addresses;
- custom entities and fields;
- character patterns/transforms;
- combinations;
- external requirements.

No one-category restriction exists.

Conceptual contracts:

```go
type GymkhanaTaskPlan struct {
    Title        string
    Summary      string
    Requirements []TaskRequirement
    Constraints  []TaskConstraint
    Objectives   []TaskObjective
    Warnings     []TaskWarning
}
```

Requirement kinds may include person, document, bill, address, character pattern, character transformation, custom entity, combination, evidence, external, and unknown.

## 38. Requirement binding and cardinality

Bindings:

```text
same_entity
different_entities
same_owner
distinct_records
any_member
whole_combination
```

These bindings determine whether a person, document, and bill must belong to the same Profile or to different people/records.

Cardinality supports exact, minimum, and maximum.

Requirement status:

```text
satisfied
partially_satisfied
not_satisfied
not_verifiable
ambiguous
```

Core does not infer missing quantities silently.

## 39. Reference dictionaries

Core may define neutral versioned dictionary contracts and matching helpers for concepts such as saints, cities, animals, colors, professions, teams, and character classes.

The consumer owns source, persistence, administration, permissions, and localization.

Dictionaries must be versioned and explained in results. AI may not silently invent a dictionary during execution.

## 40. Gymkhana combination solver

Conceptual plan:

```go
type TaskCombinationPlan struct {
    Slots       []CombinationSlot
    Constraints []CombinationConstraint
    Objectives  []TaskObjective
    MaxResults  int
}
```

The solver should:

- start with the smallest candidate set;
- propagate constraints;
- prune incompatible candidates early;
- use set operations;
- apply branch-and-bound or dynamic programming where useful;
- use bounded memoization;
- respect distinct records and owner bindings;
- stop after the configured solution limit;
- produce main and materially different alternative solutions.

It does not fetch candidates or persist checkpoints. Database owns asynchronous execution and staging.

## 41. Task ranking and evidence

Ranking may consider mandatory coverage, optional coverage, evidence clarity, warnings, diversity, record reuse, and external dependencies.

Scores are explainable and not probability of competition success.

Evidence per requirement may include entity reference, field, matched value, transformation, dictionary version, selected positions, and selected characters.

Core structures this evidence; the product minimizes sensitive output and creates navigable routes.

## 42. Imports and pure conversion

Core may provide pure functions for header normalization, typed value conversion, schema validation, row errors, and duplicate candidate features.

It does not import Excelize, run `CopyFrom`, create staging rows, interact with Google Forms, or report job progress.

## 43. Errors

Neutral errors use stable codes, paths, and arguments:

```go
type FieldError struct {
    Field string
    Code  string
    Args  map[string]any
}
```

No HTTP status, localized message, SQL, stack trace, or provider payload in public Core errors.

## 44. Testing

### Unit tests

- Unicode normalization;
- CPF/documents/phone/email;
- civil dates and year-month;
- catalog and field paths;
- operators and AST;
- relation quantifiers;
- canonical serialization/fingerprint;
- set operations and grain;
- character selection, non-contiguous subsets, order, permutation, binary transforms;
- combinations and ranking;
- explanation generation;
- assistant/tool schemas;
- OCR comparison/evidence validation;
- duplicate rules, contradictions, determinism;
- mixed Gymkhana task requirements and bindings.

### Fuzzing

- Unicode and invalid identifiers;
- date and money values;
- arbitrary ASTs;
- deeply nested plans and budget boundaries;
- pattern and selection positions;
- cardinalities and combinations;
- serialization round trips;
- malformed tool/structured-output payloads;
- duplicate inputs.

### Integration boundary tests

Core tests only neutral contracts and pure algorithms. SQL, PostgreSQL, SSE, River, R2, providers, UI, merge transactions, and egress tests stay in Database/UI.

## 45. Versioning and releases

Start at `v0.1.0` and follow SemVer. Breaking changes are explicit even in `0.x`.

Go v2 uses `/v2` module suffix.

`CHANGELOG.md` contains Unreleased, Added, Changed, Fixed, and Security. No Changesets initially.

Manual-assisted release runs all checks, validates changelog, creates one release commit, tag, and GitHub Release. Do not release on every merge.

## 46. Branches and CI

Short-lived feature/fix/refactor/chore/agent branches from `main`, PR required during active development, resolved review threads, protected main, squash merge, and English commit/PR text.

CI:

```text
gofmt check
go vet ./...
staticcheck ./...
go test ./...
govulncheck ./...
OSV scan
```

Race runs on main/release/concurrency-relevant changes and justified periodic checks.

A Core release does not automatically update or deploy Gymkhana Database.

## 47. Extraction criteria

Move logic to Core only when it is:

- independent from PostgreSQL, HTTP, UI, and provider SDKs;
- not tied to a specific persisted entity implementation;
- reused by more than one flow;
- deterministic and easily tested;
- stable enough for a public package;
- removing real duplication.

Otherwise keep it in Gymkhana Database until reuse is proven.

## 48. Deferred decisions

- physical SQL compiler;
- final provider/model choices;
- product prompt content;
- natural-language model implementation;
- arbitrary code or SQL tools;
- public distribution;
- automatic cross-repository updates;
- Changesets;
- final optimizations before benchmarks;
- any React/Data Grid/mobile implementation in Core.

## 49. Next stage

**Stage 8 — APIs, permissions, and complete frontend/backend flows.**

Core impact should be limited to stable neutral contracts discovered while defining API resources, authorization-independent validation, idempotency inputs, result envelopes, and operation states.

Update this document again when Stage 8 is completed.