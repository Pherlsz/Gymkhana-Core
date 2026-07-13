# Contributing

## Before starting

Read `docs/ORCHESTRATION.md`. Gymkhana Core only accepts reusable, deterministic logic that is independent from product infrastructure.

## Language

Code, identifiers, comments, documentation, commits, pull requests, changelog entries, workflows, and CI messages are written in English.

## Branches and commits

Create short-lived branches from `main`, such as:

```text
feature/core-normalization
fix/civil-date-validation
refactor/query-contracts
chore/bootstrap-core
```

Use Conventional Commits when practical:

```text
feat(normalize): add canonical CPF normalization
fix(civiltime): reject invalid leap days
```

## Local checks

```sh
make check
```

Run `make fuzz-smoke` when the changed package contains fuzz tests. Algorithm changes should include benchmarks where relevant.

## Pull requests

Pull requests must describe:

- objective and scope;
- public API impact;
- determinism and limits;
- tests, fuzzing, and benchmarks;
- consumer impact;
- security implications;
- changelog entry when public behavior changes.

Do not merge infrastructure dependencies, localized product text, HTTP status codes, SQL, provider SDK types, or arbitrary execution facilities into Core.

## Public API

Keep public APIs small. Do not create placeholder packages or speculative abstractions. Breaking changes require a changelog entry and migration notes for Gymkhana Database.

## Security

Never commit secrets, private data, prompts containing real data, signed URLs, credentials, or provider payloads. Follow `SECURITY.md` for vulnerability reports.
