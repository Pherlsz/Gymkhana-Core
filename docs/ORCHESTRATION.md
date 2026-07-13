# Gymkhana Core — Orchestration Document

> **Planning version:** Stage 10  
> **Last synchronized:** 2026-07-13  
> **Current stage:** Stage 10 completed — M0 implementation authorized  
> **Primary source of truth:** `Pherlsz/Gymkhana-Database/docs/ORCHESTRATION.md`  
> **Repository responsibility:** reusable Go logic independent from infrastructure

This document records the approved implementation preparation for Gymkhana-Core. Product, persistence, HTTP, OpenAPI, provider adapters, workers, and UI remain outside this repository.

## 1. Mission

Provide a small, deterministic, testable Go module for reusable logic that does not depend on PostgreSQL, HTTP, React, River, R2, Cloud Run, Vercel, or provider SDKs.

Core exists to host stable reusable contracts and algorithms such as:

- normalization;
- civil date and year-month handling;
- canonical serialization and fingerprints;
- Query catalog, AST, QueryPlan, and ExecutionPlan;
- matching and duplicate assessment;
- provider-neutral Assistant/tool contracts;
- OCR suggestion/evidence contracts;
- character selection/transformation;
- combination solving and ranking.

Do not create broad packages named `utils`, `helpers`, `common`, `shared`, or `core`.

## 2. Approved principles

- Go 1.26;
- standard library first;
- pure functions and deterministic output where possible;
- interfaces only at real substitution boundaries;
- small public API;
- no infrastructure types in public contracts;
- no HTTP status, SQL, database schema, roles, permissions, localized text, or provider SDK types;
- no arbitrary SQL, code, expression runtime, or autonomous agent loop;
- structured reasons and errors;
- no score presented as probability without a validated probabilistic model;
- no premature extraction from Gymkhana Database;
- breaking changes before v1 remain explicit and documented.

## 3. What does not belong in Core

- PostgreSQL repositories, pgx, sqlc, migrations, indexes, or transactions;
- HTTP handlers, cookies, sessions, CSRF, idempotency persistence, or OpenAPI DTOs;
- roles, permissions, field visibility, resource ownership, or audit storage;
- River, R2, signed URLs, file downloads, PDF rendering, or Cloud Run;
- OpenAI/Google SDK types, prompts, secrets, model choices, quotas, or pricing;
- React, TypeScript, Data Grid, routes, forms, or Query Keys;
- complete Profile/document/bill persistence models;
- import execution, Forms synchronization, export generation, duplicate queue persistence, merge transactions, or OCR application.

## 4. Repository structure for M0

```text
Gymkhana-Core/
├── .github/
│   ├── workflows/
│   │   ├── ci.yml
│   │   ├── security.yml
│   │   └── release.yml
│   ├── dependabot.yml
│   └── pull_request_template.md
├── docs/ORCHESTRATION.md
├── internal/
│   └── testutil/
├── .editorconfig
├── .gitignore
├── CHANGELOG.md
├── CONTRIBUTING.md
├── LICENSE
├── README.md
├── SECURITY.md
├── go.mod
├── go.sum
├── Makefile
└── tools.go
```

`internal/testutil` is created only when real shared test support exists. Empty public packages are not committed as placeholders.

## 5. Module contract

```go
module github.com/Pherlsz/Gymkhana-Core

go 1.26
```

Rules:

- no dependency on Gymkhana Database or Gymkhana UI;
- no permanent `replace` directives;
- exact dependency versions in `go.sum`;
- no framework;
- no PostgreSQL or provider SDK dependency;
- private consumption through `GOPRIVATE=github.com/Pherlsz/Gymkhana-Core`;
- Gymkhana Database pins exact tags.

A local parent `go.work` may be used during development but is not committed as a production dependency mechanism.

## 6. Package creation order

M0 starts without artificial public packages.

M1 creates only stable foundation packages:

```text
normalize
civiltime
```

Canonical fingerprint helpers may become a separate `fingerprint` package only when they form a coherent reusable API. Otherwise they remain close to consuming types until stabilized.

Future packages are added by milestone, not in M0:

```text
query        M10
matching     M11
assistant    M12
ocr          M13
solver       M14
```

Package names remain concrete and responsibility-focused.

## 7. M1 normalization scope

Potential `normalize` APIs cover:

- trim and whitespace collapse;
- Unicode normalization;
- case and accent-insensitive search forms;
- punctuation handling;
- CPF canonicalization and validation helpers;
- alphanumeric document identifiers;
- phone;
- email;
- preservation of leading zeros;
- separate display and canonical values.

Normalization rules that affect persisted fingerprints, matching, or search receive explicit versions when needed.

## 8. M1 temporal scope

`civiltime` provides:

- `CivilDate`;
- `YearMonth`;
- parsing;
- validation;
- comparison;
- leap-year rules;
- month arithmetic;
- canonical serialization.

Civil dates and year-month values never depend on timezone. Real instants remain consumer-owned UTC `time.Time` values.

## 9. Future milestone delivery

### M10 — Query contracts

- value types;
- catalog entities, fields, relations, operators, capabilities, and sensitivity metadata;
- field paths and relation quantifiers;
- Predicate, RelationPredicate, And, Or, Not;
- QueryPlan;
- validation and budgets;
- canonical serialization and fingerprints;
- result and explanation contracts.

### M11 — Matching

- canonical ProfilePair;
- compact evidence vector;
- supporting/contradicting evidence;
- levels VERY_STRONG, PROBABLE, POSSIBLE;
- deterministic rules and fingerprints;
- name alone never producing a strong level.

### M12 — Assistant

- provider-neutral request/result/capabilities;
- typed tool schemas;
- structured result parts;
- streaming event contracts;
- no SDK types, SQL, code execution, or unbounded correction loops.

### M13 — OCR

- SuggestedValue;
- evidence and normalized bounding boxes;
- comparison status;
- review decisions;
- malformed structured-output validation;
- no provider, storage, or mutation behavior.

### M14 — Advanced execution and solver

- ExecutionPlan DAG;
- set operations and result grain;
- contiguous, ordered subsequence, unordered subset, and permutation semantics;
- binary/digit/character transforms;
- combination slots, bindings, constraints, pruning, branch-and-bound, and bounded memoization;
- deterministic ranking and evidence.

## 10. Makefile commands

Initial targets:

```text
make setup
make format
make format-check
make vet
make lint
make test
make test-race
make fuzz-smoke
make vuln
make security
make build
make check
```

`make check` runs all mandatory local PR checks.

Principles:

- simple, discoverable targets;
- fail fast;
- no hidden network side effects in ordinary tests;
- exact tool versions;
- README documents prerequisites and commands;
- race and fuzz smoke targets stay bounded for developer use.

## 11. Tool management

`tools.go` uses a tools build tag only for Go tools that are appropriate to pin through the module.

Expected tools when implementation confirms compatibility:

- staticcheck;
- govulncheck;
- any additional Go-native generator or analyzer actually used by CI.

Do not add unused tools or a dependency merely to populate `go.sum`.

## 12. CI workflows

### `ci.yml`

- exact Go version;
- dependency cache;
- `gofmt` check;
- `go vet ./...`;
- `staticcheck ./...`;
- `go test ./...`;
- module/build verification.

Race runs on main, release, scheduled/manual runs, and concurrency-relevant changes rather than every trivial PR when cost is unjustified.

### `security.yml`

- govulncheck;
- OSV scan;
- dependency review on PRs;
- secret scan according to the approved tool;
- GitHub Actions pinned by commit SHA.

### `release.yml`

Manual release only:

1. validate clean main and version;
2. validate CHANGELOG;
3. run full checks;
4. create release commit if required;
5. create annotated tag;
6. create GitHub Release;
7. verify the module can be consumed by Gymkhana Database.

No release on every merge.

## 13. Testing strategy

### M0

- module builds;
- package documentation and examples compile when present;
- CI and security commands execute;
- clean consumer import is verified.

### M1

- unit tests for Unicode, identifiers, dates, and year-month;
- fuzzing for malformed Unicode, identifiers, dates, and serialization;
- stable test vectors;
- deterministic round trips.

### Later milestones

- Query AST/property tests;
- plan serialization/fingerprint stability;
- matching determinism and contradiction tests;
- malformed Assistant/OCR schema fuzzing;
- exhaustive small-domain character/solver tests;
- algorithm benchmarks and budget boundaries.

Core tests do not spin up PostgreSQL, HTTP servers, provider APIs, or browser environments.

## 14. Documentation files

- README: mission, boundaries, setup, commands, consumption, and status;
- CONTRIBUTING: language, branch/commit policy, public API review, tests, fuzz, benchmark, and changelog;
- SECURITY: private vulnerability reporting without inventing contact details;
- CHANGELOG: Unreleased, Added, Changed, Fixed, Security;
- proprietary/all-rights-reserved license while private;
- `.editorconfig` with UTF-8/LF/final newline;
- `.gitignore` for actual artifacts only.

Documentation must not describe nonexistent APIs as available.

## 15. Branches and first PRs

Branches:

```text
chore/bootstrap-core
feat/core-normalization
```

Initial PR sequence:

1. module bootstrap, root files, Makefile, and documentation;
2. CI and security workflows;
3. normalization and civiltime foundations;
4. release workflow and v0.1.0 preparation;
5. v0.2.0 preparation when M1 APIs are stable and Gymkhana Database is ready to consume them.

A smaller number of PRs is acceptable when grouping improves review and keeps `main` coherent.

## 16. Initial issues

M0 issues:

- bootstrap Go module and repository root;
- add local quality commands;
- add CI workflow;
- add security workflow;
- add documentation and contribution policy;
- add release workflow;
- verify clean consumer import;
- publish and validate v0.1.0.

M1 issues:

- implement text normalization;
- implement CPF/document normalization;
- implement phone/email normalization;
- implement CivilDate;
- implement YearMonth;
- add fuzz and deterministic test vectors;
- prepare foundation release.

Each issue defines scope, exclusions, dependencies, acceptance criteria, tests, public API impact, and consumer impact.

## 17. Release roadmap

Conceptual sequence:

```text
v0.1.0 module and release foundation
v0.2.0 normalization, civil dates, and canonical foundations
v0.3.0 Query catalog, AST, and QueryPlan
v0.4.0 matching and duplicate assessment
v0.5.0 Assistant/provider/tool contracts
v0.6.0 OCR contracts
v0.7.0 ExecutionPlan, character engine, and solver
```

Exact version numbers follow actual SemVer impact. Releases occur only at real Gymkhana Database consumption points.

## 18. Pull request requirements

Core PRs include:

- objective and scope;
- public API changes;
- deterministic behavior;
- tests and fuzzing;
- benchmarks for algorithms;
- structured errors;
- dependency impact;
- infrastructure-independence review;
- CHANGELOG entry when public behavior changes;
- Gymkhana Database consumer impact.

## 19. Definition of Done for Core algorithms

- stable and intentionally small contract;
- deterministic output;
- unit tests;
- fuzz/property tests when appropriate;
- benchmark for nontrivial algorithms;
- explicit complexity/budget behavior;
- structured neutral errors;
- documentation and examples;
- no infrastructure or provider leakage;
- release consumed by Gymkhana Database when required.

## 20. Definition of Done for M0

M0 is complete when:

- clean clone and module download work;
- build and tests pass;
- formatting, vet, staticcheck, vulnerability, and security workflows pass;
- README, CONTRIBUTING, SECURITY, CHANGELOG, and license exist;
- release workflow is validated;
- no artificial public package exists;
- v0.1.0 is tagged and importable by Gymkhana Database;
- no PostgreSQL, HTTP, provider, UI, or product concern leaked into the module.

## 21. Implementation authorization

Stage 10 is complete. Implementation starts here:

```text
repository: Pherlsz/Gymkhana-Core
branch: chore/bootstrap-core
milestone: M0 Bootstrap
first scope: module, root files, Makefile, CI/security, and release foundation
```

After Core and UI publish verified minimum releases, Gymkhana Database integrates exact versions and continues the M0 bootstrap.

The orchestration document is updated again when a milestone closes or an approved architectural decision changes.