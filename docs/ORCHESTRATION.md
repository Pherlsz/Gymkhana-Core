# Gymkhana Core — Orchestration Document

> **Planning version:** Stage 5  
> **Last synchronized:** 2026-07-12  
> **Current stage:** Stage 5 completed  
> **Primary source of truth:** `Pherlsz/Gymkhana-Database/docs/ORCHESTRATION.md`  
> **Repository responsibility:** reusable Go logic independent from infrastructure

This document records every approved decision that affects `Gymkhana-Core`. Product and persistence decisions are owned by the `Gymkhana-Database` orchestration document.

## 1. Repository mission

Provide a small, deterministic, testable Go module for reusable logic that does not depend on Gymkhana Database persistence, HTTP, React, queue providers, storage providers, or AI SDKs.

Expected responsibilities:

- data normalization;
- civil date and year-month value types;
- typed query operators and AST;
- generic matching;
- duplicate scoring and structured reasons;
- neutral AI contracts;
- neutral tool orchestration contracts;
- result and reference formats;
- pure validation functions where reuse is proven.

## 2. Engineering language

The following must be written in English:

- source code identifiers and technical comments;
- documentation;
- commit messages;
- pull request titles and descriptions;
- changelog entries;
- release notes;
- workflow names and CI messages.

The module must not contain product-facing Portuguese copy. Consumer applications are responsible for localization.

## 3. What does not belong in Core

Do not include:

- PostgreSQL repositories or SQL queries;
- `pgx`, sqlc, or migrations;
- HTTP handlers or OpenAPI-generated product types;
- authentication, sessions, or product permissions;
- River or any queue implementation;
- R2, S3, or object storage;
- Google Forms integration;
- OpenAI or Google GenAI SDK types;
- Vercel, Cloud Run, Neon, or Cloudflare configuration;
- React, TypeScript, or UI components;
- complete product entities without real reuse value;
- MEMBER, ADMIN, or SUPERADMIN logic;
- persistence models from Gymkhana Database.

## 4. Core principles

- Keep domain logic and infrastructure separate.
- Add interfaces only at real substitution boundaries.
- Avoid premature abstraction.
- Do not install a dependency to save a few lines.
- Do not expose external package types through public APIs unless unavoidable.
- Do not upgrade only because a newer release exists.
- Do not release with a known and applicable vulnerability.
- Prefer pure functions and deterministic outputs.
- Explain matching results with structured reasons, not arbitrary percentages.
- Never produce or execute arbitrary SQL.
- Describe intent and plans; leave data access to the consumer.
- Keep the public API intentionally small.

## 5. Approved stack

- Go 1.26;
- Go modules;
- standard library first;
- `testing`;
- native fuzzing;
- race detector where concurrency exists;
- `gofmt`;
- `go vet`;
- `staticcheck`;
- `govulncheck`;
- OSV-Scanner;
- Dependabot Alerts;
- Dependency Review when available.

Do not use an application framework, ORM, generic agent framework, or assertion framework initially.

Exact patch versions are pinned during repository initialization after compatibility, changelog, and security review.

## 6. Repository structure

Public packages live directly at the repository root. Do not add a `pkg/` directory only by convention.

Approved initial layout:

```text
Gymkhana-Core/
├── normalize/
├── civiltime/
├── query/
├── matching/
├── duplicates/
├── assistant/
├── tools/
├── result/
├── internal/
│   └── testutil/
├── docs/
├── go.mod
├── go.sum
├── CHANGELOG.md
├── SECURITY.md
└── README.md
```

A package is public only when Gymkhana Database actually imports it.

Internal helpers stay under `internal/`.

Do not create broad packages named:

- `utils`;
- `helpers`;
- `common`;
- `shared`;
- `core`.

## 7. Module identity

Go module path:

```go
module github.com/Pherlsz/Gymkhana-Core
```

Example imports:

```go
import (
    "github.com/Pherlsz/Gymkhana-Core/normalize"
    "github.com/Pherlsz/Gymkhana-Core/query"
)
```

Gymkhana Database pins an exact tagged version.

## 8. Normalization

Potential reusable normalization areas:

- names;
- search text;
- whitespace;
- case;
- accents and diacritics;
- CPF;
- phone numbers;
- email;
- document identifiers;
- alphanumeric values;
- option keys;
- civil dates;
- year-month values;
- address fragments when sufficiently generic.

Important rules:

- preserve leading zeros;
- keep document identifiers as text;
- use uppercase canonical letters when applicable;
- remove visual separators from canonical identifiers;
- preserve display values outside search-normalization functions;
- lowercase email;
- keep civil dates separate from UTC instants;
- keep `YEAR_MONTH`, `DATE`, and `DATETIME` conceptually distinct.

Core does not decide persistence format or database columns.

## 9. Neutral temporal types

Required conceptual separation:

- `CivilDate`: year, month, and day without timezone;
- `YearMonth`: year and month;
- `Instant`: real instant represented by the consumer with UTC `time.Time`.

Core may provide:

- strict parsing;
- validation;
- comparison;
- month navigation;
- leap-year rules;
- canonical serialization;
- neutral formatting helpers when not tied to UI localization.

Core does not own `pt-BR` display formatting.

## 10. Query Engine model

Core defines the structured query language, not SQL.

Conceptual API:

```go
type QueryPlan struct {
    Entity     EntityRef
    Filters    []Filter
    Includes   []RelationRef
    Sort       []Sort
    Projection []FieldRef
    Grouping   []Group
    Limit      int
}
```

Potential elements:

- target entity;
- fields;
- typed values;
- operators;
- logical composition;
- included relations;
- projection;
- ordering;
- grouping;
- limits;
- intent metadata.

Generic operators may include:

- equal and not equal;
- contains;
- starts with and ends with;
- in and not in;
- greater/greater-or-equal;
- less/less-or-equal;
- between;
- is empty and is not empty;
- normalized match;
- approximate match when supported by the executor.

Core validates:

- operator compatibility with field type;
- required values;
- valid limits;
- logical composition;
- unknown fields and operators;
- configurable complexity and nesting limits.

Core does not know PostgreSQL column names, indexes, joins, or schemas.

## 11. Field catalog

The consumer provides a neutral catalog of available fields.

Conceptual model:

```go
type FieldDefinition struct {
    Key        string
    Label      string
    Type       ValueType
    Operators  []Operator
    Searchable bool
    Sortable   bool
    Filterable bool
}
```

Gymkhana Database builds the catalog from:

- native fields;
- document and bill types;
- custom entity types;
- custom fields;
- current user permissions.

Core validates plans against the catalog but never loads it from a database.

## 12. Results and references

Neutral result concepts:

```go
type ResultReference struct {
    EntityType string
    EntityID   string
    Label      string
    FieldKey   string
}

type MatchResult struct {
    Reference    ResultReference
    MatchedValue string
    Highlight    string
    Relevance    float64
}
```

The product may add URLs, labels, and permission-aware navigation afterward.

Core does not create web routes or absolute URLs.

## 13. Generic matching and duplicates

Approved product levels:

- `VERY_STRONG`;
- `PROBABLE`;
- `POSSIBLE`.

Core returns structured reasons, not a confidence percentage presented as truth.

Conceptual reason:

```go
type DuplicateReason struct {
    RuleKey    string
    FieldKey   string
    Strength   Strength
    MessageKey string
}
```

Potential reusable signals:

- identical CPF;
- identical document identifier;
- identical phone;
- identical email;
- similar normalized name;
- equal birth date;
- similar parent names;
- similar address;
- combinations of weaker signals.

Name alone must never produce a strong duplicate result.

Gymkhana Database decides:

- available fields;
- active rules and thresholds;
- persistence of reviews;
- merge behavior;
- permissions;
- when scans run.

Core calculates and explains only.

## 14. Rule versioning

Matching evaluation may accept:

- rule version;
- source data versions;
- normalization version;
- thresholds by level.

This enables the consumer to decide whether a dismissed review should return after data or rule changes.

Core does not persist versions or review records.

## 15. Neutral AI contracts

Core must not depend on provider SDKs.

Conceptual contracts:

```go
type GenerateRequest struct {
    Messages []Message
    Tools    []ToolDefinition
    Schema   *OutputSchema
}

type GenerateResult struct {
    Text       string
    ToolCalls  []ToolCall
    Structured any
    Usage      Usage
}
```

Gymkhana Database implements OpenAI, Google, or future provider adapters.

No `openai-go` or `go-genai` type may appear in Core public APIs.

## 16. Neutral AI orchestration

Approved flow:

1. receive a question;
2. interpret intent;
3. create a typed plan;
4. validate against catalog and policy;
5. request execution through a neutral tool;
6. receive structured results;
7. synthesize a response;
8. emit references.

Core may provide a small state machine and neutral interfaces.

Do not use:

- LangChain;
- LangGraph;
- CrewAI;
- Semantic Kernel;
- a generic agent framework;
- arbitrary SQL tools;
- database credentials;
- direct repository access.

Example neutral executor:

```go
type QueryExecutor interface {
    Execute(ctx context.Context, plan QueryPlan) (QueryResult, error)
}
```

Gymkhana Database implements the executor with safe domain modules and SQL.

## 17. Tool calling

Tools must be small, explicit, and typed.

Potential tools:

- query Profiles;
- query documents;
- query bills;
- query custom entities;
- retrieve record details;
- produce a navigable table result.

Never add a generic `execute_sql` tool.

Each tool declares:

- stable name;
- description;
- input schema;
- output schema;
- neutral required-permission metadata;
- result limits;
- stable technical errors.

The consumer performs real authorization.

## 18. Streaming events

Gymkhana Database uses SSE, but Core does not know HTTP or SSE.

Core may expose neutral events:

- `message.started`;
- `text.delta`;
- `tool.started`;
- `tool.completed`;
- `references`;
- `message.completed`;
- `message.failed`;
- `message.cancelled`.

The product maps them to its transport.

## 19. OCR-related logic

Core may provide:

- neutral extraction schemas;
- structured result validation;
- current-versus-suggested comparison;
- normalization of extracted values;
- classification of accepted, rejected, and conflicting fields.

Core does not:

- download files;
- access R2;
- call AI providers;
- persist results;
- modify entities;
- authorize actions.

Human review remains a product requirement.

## 20. Custom values

Core may define generic value types:

- short and long text;
- number;
- money;
- civil date;
- datetime;
- boolean;
- single and multi-select;
- email;
- phone;
- URL;
- attachment reference.

It may validate operator compatibility and value conversion.

It does not know database IDs, owner tables, storage, React widgets, or product import workflows.

## 21. Import-related pure logic

Core may provide pure functions for:

- header normalization;
- textual value conversion;
- row validation against a neutral schema;
- field-level validation errors;
- duplicate candidate generation from normalized values.

Do not include:

- Excelize;
- R2;
- `pgx.CopyFrom`;
- PostgreSQL staging;
- jobs and progress;
- Google Forms.

## 22. Errors

Core uses stable neutral technical errors.

Conceptual field error:

```go
type FieldError struct {
    Field string
    Code  string
    Args  map[string]any
}
```

Gymkhana Database converts them to localized messages and HTTP envelopes.

Core does not return HTTP status codes.

## 23. Testing

### 23.1 Unit tests

Prioritize:

- Unicode normalization;
- accents and whitespace;
- CPF and document identifiers;
- phones and email;
- civil dates and year-month;
- operators and AST validation;
- matching and duplicate levels;
- tool schemas;
- orchestration state transitions.

### 23.2 Fuzzing

Use native fuzzing for:

- Unicode;
- arbitrary document characters;
- phone numbers;
- dates;
- query parsing;
- normalizers;
- filter values;
- tool schemas.

### 23.3 Determinism

The same input, configuration, and rule version must produce the same output.

Tests must not rely on global clock state, machine locale, or map iteration order.

## 24. Dependency and security policy

Before adding a package:

- prove the need;
- review license;
- review maintenance activity;
- inspect transitive dependencies;
- check known vulnerabilities;
- assess whether the standard library is sufficient;
- prevent external types from leaking into the public API.

Checks:

```text
gofmt check
go vet ./...
staticcheck ./...
go test ./...
go test -race ./...
govulncheck ./...
```

Supply-chain controls:

- exact versions;
- committed `go.sum`;
- OSV-Scanner;
- Dependabot Alerts;
- Dependency Review when available;
- GitHub Actions pinned by commit SHA;
- no automatic merge of dependency updates.

## 25. Private consumption

Gymkhana Database consumes Core as a private Go module:

```go
require github.com/Pherlsz/Gymkhana-Core v0.1.0
```

Configuration:

```text
GOPRIVATE=github.com/Pherlsz/Gymkhana-Core
```

CI uses a fine-grained read-only token restricted to `Gymkhana-Core` with expiration and rotation.

Do not use permanent `replace` directives, branches, untagged commits, submodules, subtree, or source copying in production.

## 26. Local multi-repository development

Repositories may be checked out side by side:

```text
Gymkhana/
├── Gymkhana-Database/
├── Gymkhana-Core/
├── Gymkhana-UI/
└── go.work
```

A local `go.work` may connect Database and Core, but it stays outside repositories and is not committed.

The committed `go.mod` always points to a tagged exact version.

No fourth coordination repository is created.

## 27. Dependency direction

Approved graph:

```text
Gymkhana-Database → Gymkhana-Core
Gymkhana-Core     → no dependency on Database or UI
```

Core and UI have no direct dependency.

Go and TypeScript share product contracts through OpenAPI in Gymkhana Database, not through a multi-language “shared” package.

## 28. Versioning

Start at:

```text
v0.1.0
```

Use SemVer:

- patch: compatible fix;
- minor: compatible feature;
- major: public API break.

For Go major versions:

```text
v1: github.com/Pherlsz/Gymkhana-Core
v2: github.com/Pherlsz/Gymkhana-Core/v2
```

Even during `0.x`, breaking changes must be explicit and documented.

Reach `1.0.0` only after production use, stable public APIs, minimum documentation, and contract test coverage.

## 29. Releases and changelog

Do not use Changesets.

Maintain `CHANGELOG.md` with:

```text
Unreleased
Added
Changed
Fixed
Security
```

A manually triggered release workflow accepts `patch`, `minor`, or `major`, runs all checks, validates the changelog, creates one release commit, tags the module, and creates a GitHub Release.

Do not create a release for every merge.

## 30. Branches, commits, and pull requests

Development is based on `main` with short-lived branches:

- `feature/*`;
- `fix/*`;
- `refactor/*`;
- `chore/*`;
- `agent/*`.

No `develop` branch or Git Flow.

Main protection when active development begins:

- pull request required;
- required checks;
- resolved review threads;
- no force push;
- squash merge by default.

All commit messages, PR titles, and PR descriptions are written in English.

Remote pushes should be grouped to avoid unnecessary CI runs.

## 31. CI

PR checks:

- format check;
- `go vet`;
- `staticcheck`;
- unit tests;
- vulnerability scan;
- OSV scan.

Race detector runs:

- on `main`;
- before release;
- periodically when justified;
- on PRs that change concurrency.

New pushes cancel obsolete checks for the same branch. Release workflows are not cancelled this way.

A Core release does not automatically trigger Database builds, dependency PRs, Vercel deploys, or cross-repository workflows.

## 32. Compatibility strategy

Gymkhana Database supports one exact Core version before `1.0`, but breaking changes are never silent.

Safe incompatible change flow:

1. add a new API while keeping the old API;
2. publish a compatible release;
3. migrate Gymkhana Database;
4. validate production usage;
5. remove the old API in a later incompatible release.

Use Go deprecation comments:

```go
// Deprecated: use NewQueryPlan instead.
```

A breaking release documents:

- changed API;
- reason;
- introduced version;
- migration steps;
- minimum compatible Database version;
- deprecation window, when available.

## 33. Reproducible builds

Use:

```text
go mod download
go test -mod=readonly ./...
go build -mod=readonly
```

Rules:

- no floating dependencies;
- exact tool versions;
- GitHub Actions pinned by SHA;
- release built only by the official workflow;
- caches never replace module checksums.

## 34. Local commands

Approved Makefile commands:

```text
make test
make lint
make fuzz
make check
```

`make check` represents everything required before a pull request.

Do not add Taskfile, Just, Mage, or another orchestration runtime initially.

## 35. Documentation structure

- `README.md`: purpose, requirements, setup, commands, links;
- `docs/architecture/`: current architecture;
- `docs/guides/`: procedures such as adding a package and publishing a release;
- `docs/adr/`: only important architectural decisions;
- `docs/ORCHESTRATION.md`: approved planning decisions;
- `SECURITY.md`: private vulnerability reporting.

Do not automatically copy orchestration documents between repositories. Keep the same planning version, synchronization date, source-of-truth reference, and repository-specific responsibilities.

## 36. CODEOWNERS and review

Initial ownership may point to one maintainer, with paths prepared for future specialization:

```text
/query/
/duplicates/
/assistant/
```

Changes requiring extra review attention:

- Query AST and operator semantics;
- duplicate scoring behavior;
- neutral AI and tool contracts;
- public package exports;
- release workflows;
- dependency and security configuration.

No bot approval.

## 37. Pull request and issue templates

Use a short PR template:

```text
What changed
Why
How it was validated
Contract or migration impact
Security checklist
```

Issue templates:

- Bug;
- Feature;
- Security, redirected to private reporting;
- Technical debt.

Security reports must not be public issues.

## 38. Criteria for extracting logic to Core

Move logic from Gymkhana Database to Core only when it satisfies most of these conditions:

- independent from PostgreSQL;
- independent from HTTP;
- independent from UI;
- independent from a specific persisted entity;
- useful in more than one flow;
- easily testable as a neutral function or contract;
- sufficiently stable public API;
- removes real duplication.

When uncertain, keep the logic in Gymkhana Database and extract it after real reuse is demonstrated.

## 39. Version policy

Compare the most mature proven line with the latest stable line.

Adopt a newer version only for:

- applicable security fix;
- relevant bug fix;
- measurable performance gain;
- required compatibility;
- feature actually used;
- clear complexity reduction.

Betas and release candidates are limited to research branches. Production releases use stable versions.

## 40. Deferred decisions

- final complete operator catalog;
- natural-language parser owned by Core;
- final duplicate scoring rules;
- final tool schemas;
- support for additional AI provider concepts;
- optimization before benchmarks;
- a fourth workspace repository;
- automatic cross-repository updates;
- Changesets;
- public module distribution.

## 41. Next stage

**Stage 6 — Design system and Data Grid** primarily affects Gymkhana UI and Gymkhana Database.

Core should only be updated during Stage 6 if a truly reusable, infrastructure-independent contract is identified.

Update this document again when Stage 6 is completed.