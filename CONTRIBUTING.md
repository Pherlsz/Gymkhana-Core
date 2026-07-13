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

Use Conventional Commits for the final pull request title and squash or merge commit:

```text
feat(normalize): add canonical CPF normalization
fix(civiltime): reject invalid leap days
feat(query)!: replace the public predicate contract
```

Version impact is derived from that final title:

- `fix` produces a patch release;
- `feat` produces a minor release;
- breaking changes produce a minor release before `1.0.0` and a major release afterwards;
- non-releasable types such as `chore`, `docs`, and `test` keep the current version.

## Release preparation

The functional pull request must include its own `VERSION` and `CHANGELOG.md` updates. Prepare them before the final commit:

```sh
python scripts/release-version.py write \
  --base-version <version-on-main> \
  --title "<final-pr-title>"
```

CI runs the same calculation in check mode. A version mismatch or missing changelog section blocks the pull request. After merge, the release workflow creates the tag and GitHub Release directly from `main`; it never opens a second pull request.

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
- migration notes for breaking public behavior.

Do not merge infrastructure dependencies, localized product text, HTTP status codes, SQL, provider SDK types, or arbitrary execution facilities into Core.

## Public API

Keep public APIs small. Do not create placeholder packages or speculative abstractions. Breaking changes require migration notes for Gymkhana Database.

## Security

Never commit secrets, private data, prompts containing real data, signed URLs, credentials, or provider payloads. Follow `SECURITY.md` for vulnerability reports.
