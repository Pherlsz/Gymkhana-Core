# Contributing

## Before starting

Read `docs/ORCHESTRATION.md`. Gymkhana Core only accepts reusable, deterministic logic that is independent from product infrastructure.

When work originates from Gymkhana Database, read and link the originating Database issue before creating Core scope. Database issue #31 remains the cross-repository roadmap and continuation tracker.

## Language

Code, identifiers, comments, documentation, commits, pull requests, changelog entries, workflows, and CI messages are written in English.

## Branches, milestones, and commits

Create one short-lived branch from `main` for each active workstream, such as:

```text
feature/core-normalization
fix/civil-date-validation
refactor/query-contracts
chore/bootstrap-core
```

Open one draft pull request for that branch. Do not create helper PRs for formatting, diagnostics, validation, export, or squash operations.

A Core milestone is appropriate only when the Core portion has multiple executable issues or an independently meaningful repository delivery. It must link the originating Database issue, contain its Core parent and executable issues, and avoid duplicating each issue with its delivery PR in milestone progress. One-off dependencies use one issue and one PR without an empty milestone.

Use Conventional Commits for the final pull request title and squash commit:

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

CI runs the same calculation in check mode. A version mismatch or missing changelog section blocks the pull request. After a protected PR merge changes `VERSION`, the release workflow creates the tag and GitHub Release directly from `main`; it never opens a second pull request or reruns the complete quality suite.

## Local checks and hosted Actions

During development, run targeted commands locally. Before marking the PR ready for review, run:

```sh
make check
```

Run `make fuzz-smoke` when the changed package contains fuzz tests. Algorithm changes should include benchmarks where relevant.

Hosted Actions are final gates:

- draft PR pushes do not start runners;
- meaningful corrections return the PR to draft;
- repeated pushes must not be used as a remote test loop;
- only transient infrastructure failures justify rerunning a failed job;
- temporary workflows, dummy commits, and close/reopen cycles used to force CI are prohibited.

## Pull requests

Pull requests must describe:

- objective and scope;
- originating Database issue when applicable;
- public API impact;
- determinism and limits;
- tests, fuzzing, and benchmarks;
- consumer impact;
- security implications;
- migration notes for breaking public behavior;
- expected Actions impact when workflow frequency, jobs, timeout, matrix, artifacts, or runner type changes.

Do not merge infrastructure dependencies, localized product text, HTTP status codes, SQL, provider SDK types, or arbitrary execution facilities into Core.

## Public API

Keep public APIs small. Do not create placeholder packages or speculative abstractions. Breaking changes require migration notes for Gymkhana Database.

## Security

Never commit secrets, private data, prompts containing real data, signed URLs, credentials, or provider payloads. Follow `SECURITY.md` for vulnerability reports.
