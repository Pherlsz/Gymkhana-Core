# Gymkhana Core

Private Go module containing deterministic, infrastructure-independent contracts and algorithms reused by Gymkhana Database.

## Repository role

Gymkhana Core owns reusable Go logic that does not depend on product infrastructure.

This repository receives work only when a Database delivery requires a reusable contract or when Core has an independently approved maintenance need. It does not maintain a competing product roadmap.

## Available packages

### `normalize`

Deterministic normalization for:

- display and search text;
- Unicode letters and digits;
- person names;
- street, house number, block (`bl. A`), apartment (`apt. 202`), CEP, and UF;
- city, neighborhood, country, and nationality;
- gender, marital status, blood type, and club membership type;
- team, sector, football club, health plan, card, collection, animal, and vehicle;
- CPF validation and canonicalization;
- CNPJ validation and canonicalization, including the alphanumeric Receita Federal form;
- catalog document identifiers (canonical form, display formatting, UF folded into the identifier when the type is state-issued, and unique type inference from unlabeled numbers);
- email canonicalization;
- Brazilian phone validation and E.164 canonicalization.

Normalization never logs or embeds the original value in validation errors.

### `civiltime`

Timezone-free value types for:

- strict `CivilDate` parsing and serialization;
- strict `YearMonth` parsing and serialization;
- comparison, leap-year rules, month arithmetic, and month boundaries.

Zero values represent absence and serialize as empty text or JSON `null`.

## Boundaries

Gymkhana Core may contain normalization, civil date/year-month handling, canonical fingerprints, typed query plans, duplicate assessment, provider-neutral Assistant/OCR contracts, character transforms, and combination algorithms.

It must not contain PostgreSQL, HTTP, OpenAPI DTOs, sessions, permissions, React, provider SDKs, workers, R2, or product-specific persistence.

## Requirements

- Go 1.26.5
- GNU Make or a compatible implementation
- GitHub authentication for private module consumption

## Setup

```sh
make setup
make check
```

Useful targeted commands:

```sh
make format
make format-check
make vet
make lint
make test
make test-race
make fuzz-smoke
make security
```

Pinned tool binaries are installed into `./bin`, which is ignored by Git.

## Private consumption

Consumers configure:

```sh
export GOPRIVATE=github.com/Pherlsz/Gymkhana-Core
```

Gymkhana Database pins an exact release:

```sh
go get github.com/Pherlsz/Gymkhana-Core@v0.2.1
```

Permanent `replace` directives, copied source, submodules, and branch dependencies are not supported.

## Development governance

Read [CONTRIBUTING.md](CONTRIBUTING.md) and [docs/ORCHESTRATION.md](docs/ORCHESTRATION.md) before changing public contracts.

Cross-repository delivery follows these rules:

- the Database issue and milestone remain the source of product scope, sequencing, and acceptance;
- every Core issue or PR created for Database work links the originating Database issue;
- do not mirror the entire Database milestone into Core;
- create a Core milestone only when the Core portion contains multiple independently executable issues or represents a distinct repository-level delivery stage;
- name a supporting milestone `M<n> support — <outcome>` and link its parent Database issue in the milestone description;
- the Core milestone contains its parent issue and executable issues; PRs normally close those issues with `Closes #...` and are not also assigned to the milestone, avoiding duplicated progress;
- one-off dependency work uses one issue and one draft PR without creating an otherwise empty milestone;
- newly discovered scope must be added to an issue, explicitly deferred in Database #31, or rejected as out of scope;
- releases and version tags remain separate from milestones.

A supporting milestone closes only after its executable issues are delivered, the Database consumer can pin the published version, and the originating Database issue records the dependency result.

## CI and GitHub Actions budget policy

The personal GitHub Pro allowance is shared across private repositories owned by the account. Hosted Actions are final integration and publication gates, not a remote development environment.

Workflow behavior:

- draft PRs do not start runner jobs;
- non-draft synchronization runs formatting, vet, unit tests, and semantic-version validation;
- Staticcheck, race tests, and fuzz smoke run when a PR first becomes reviewable, when a ready PR opens or reopens, or through manual execution;
- meaningful corrections return the PR to draft and trigger one new full run only after local validation;
- Security runs for relevant reviewable PRs, manually, and on the 1st and 15th of each month;
- successful PR quality checks are not repeated after merge;
- Release runs only when `VERSION` changes on `main` or when manually requested;
- Release creates the missing tag and GitHub Release without rerunning the complete quality suite;
- obsolete runs on the same PR are canceled;
- setting repository variable `CI_RUNNER` to a trusted Linux self-hosted runner label moves jobs away from GitHub-hosted minutes.

Mandatory usage rules:

1. Keep one implementation PR per active Core workstream. Helper, diagnostic, formatting, export, validation, and squash PRs are prohibited.
2. Open implementation PRs as drafts and run relevant commands locally before the first ready-for-review transition.
3. Run `make check` locally before a public-contract release or milestone acceptance.
4. Do not use repeated pushes, dummy commits, PR close/reopen cycles, or temporary workflows as a remote test loop.
5. Re-run only the failed job and only when the failure was caused by transient infrastructure.
6. Review account-wide Actions usage before starting a new supporting milestone. At 70% usage, move heavy gates to local or self-hosted execution; at 85%, reserve hosted runners for final merge gates and releases; at 95%, hosted execution requires owner approval.
7. Any workflow change that increases frequency, job count, timeout, matrix size, artifact retention, or runner cost must explain its monthly impact.
8. Dependabot changes require functional review and compatibility validation before merge.

## GitHub Pro protection for `main`

After GitHub Pro is active, use a branch ruleset that:

- blocks branch deletion and force pushes;
- requires changes through pull requests;
- requires review conversations to be resolved;
- requires linear history;
- does not require an independent approval while there is only one maintainer;
- does not require branches to be up to date before merge, avoiding unnecessary repeated CI;
- does not make path-filtered workflows required until a universal low-cost aggregate gate always reports a status;
- grants no routine bypass.

Allow squash merge only and automatically delete merged branches.

## Releases

The functional pull request carries its own semantic version and changelog section. Before the final commit, run:

```sh
python scripts/release-version.py write \
  --base-version <version-on-main> \
  --title "<final-pr-title>"
```

CI recalculates the expected version from the final pull request title:

- `fix` increments the patch version;
- `feat` increments the minor version;
- a `!` or `BREAKING CHANGE` increments the minor version before `1.0.0` and the major version afterwards;
- `chore`, `docs`, `test`, and other non-releasable changes keep the current version.

After a protected PR merge changes `VERSION`, the Release workflow creates the missing `vX.Y.Z` tag and GitHub Release directly. It is safe to rerun and does not create a release pull request.
