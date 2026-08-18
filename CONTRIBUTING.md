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

Every functional pull request must carry its own version and changelog updates when the change is releasable.

Versioning follows Semantic Versioning and is derived from the **final pull request title** using Conventional Commits.

### Version rules

Given a base version such as `0.2.2`:

| Pull request title                           | Version impact | Result                 |
| -------------------------------------------- | -------------- | ---------------------- |
| `fix(normalize): correct CPF formatting`     | Patch          | `0.2.3`                |
| `feat(normalize): add address normalization` | Minor          | `0.3.0`                |
| `feat(normalize)!: replace address contract` | Breaking       | `0.3.0` before `1.0.0` |
| `feat(api)!: replace public API`             | Breaking       | `2.0.0` from `1.x.x`   |
| `chore: update tooling`                      | None           | unchanged              |
| `docs: improve documentation`                | None           | unchanged              |
| `test: add normalization coverage`           | None           | unchanged              |
| `refactor: reorganize internals`             | None           | unchanged              |

A pull request is also treated as breaking when its body contains:

```text
BREAKING CHANGE:
```

Breaking changes increment the minor version while the project is below `1.0.0`. Starting from `1.0.0`, they increment the major version.

### Preparing a release locally

Before the final commit, determine the version currently present on the pull request base branch, normally `main`.

For example, if `main` currently contains:

```text
0.2.2
```

and the final pull request title will be:

```text
feat(normalize): add canonical address normalization
```

run:

```sh
python scripts/release-version.py write \
  --base-version 0.2.2 \
  --title "feat(normalize): add canonical address normalization"
```

The command prints the calculated version:

```text
0.3.0
```

and updates the repository release metadata.

For a releasable change, `write` updates:

* `VERSION` with the calculated semantic version;
* `CHANGELOG.md` with a release section for that version.

Review the generated changes before committing:

```sh
git diff -- VERSION CHANGELOG.md
```

The generated `VERSION` and `CHANGELOG.md` changes belong to the same functional pull request. Do not create a separate release-preparation pull request.

### Base version

`--base-version` is always the version from the pull request base branch **before the current pull request is applied**.

It is not necessarily the value currently present in your working tree after running the script.

For example:

```text
main VERSION:          0.2.2
feature branch VERSION after write: 0.3.0
--base-version:        0.2.2
```

Do not run the next command using `0.3.0` as the base version for the same pull request.

When needed, inspect the version on `main` directly:

```sh
git show origin/main:VERSION
```

### Pull request title consistency

The title passed to `release-version.py write` must match the final semantic meaning of the pull request title.

For example:

```sh
python scripts/release-version.py write \
  --base-version 0.2.2 \
  --title "feat(normalize): add canonical address normalization"
```

expects the pull request to remain a `feat`.

If the pull request is later changed to:

```text
fix(normalize): correct canonical address normalization
```

the expected version changes from:

```text
0.3.0
```

to:

```text
0.2.3
```

Run `write` again using the same original base version and the new final title before marking the pull request ready for review.

### Pull requests with breaking changes

Breaking changes can be declared using `!` in the title:

```sh
python scripts/release-version.py write \
  --base-version 0.2.2 \
  --title "feat(normalize)!: replace the public address contract"
```

or through the pull request body:

```text
BREAKING CHANGE: callers must migrate to the new address representation.
```

When the breaking declaration exists only in the pull request body, pass it locally with `--body`:

```sh
python scripts/release-version.py write \
  --base-version 0.2.2 \
  --title "feat(normalize): replace the public address contract" \
  --body "BREAKING CHANGE: callers must migrate to the new address representation."
```

### Changelog generation

The script derives the changelog category from the change type:

| Change          | Changelog section |
| --------------- | ----------------- |
| `feat`          | `Added`           |
| `fix`           | `Fixed`           |
| Breaking change | `Changed`         |

For example:

```text
feat(normalize): add canonical address normalization
```

may generate:

```md
## [0.3.0] - YYYY-MM-DD

### Added

- Add canonical address normalization.
```

The current date is used by default.

A specific date can be supplied when necessary:

```sh
python scripts/release-version.py write \
  --base-version 0.2.2 \
  --title "feat(normalize): add canonical address normalization" \
  --date 2026-08-18
```

Do not manually create another section for the same version after the script has generated one.

### Validation mode

The script also provides a non-mutating validation mode:

```sh
python scripts/release-version.py check \
  --base-version 0.2.2 \
  --title "feat(normalize): add canonical address normalization"
```

`check` recalculates the expected version and verifies that:

1. `VERSION` contains a valid semantic version;
2. `VERSION` matches the version required by the pull request title and body;
3. `CHANGELOG.md` contains the corresponding release section when the change requires a new release.

A successful validation prints information similar to:

```json
{"base":"0.2.2","expected":"0.3.0","actual":"0.3.0"}
```

`check` does not prepare a release and should not be used instead of `write` when release metadata still needs to be generated.

### CI validation

CI performs the same version calculation using:

* the `VERSION` value from the pull request base commit;
* the current pull request title;
* the current pull request body.

A mismatch between the calculated version and the committed `VERSION`, or a missing changelog section for a releasable change, blocks the pull request.

This means changing the pull request title or adding/removing a breaking-change declaration can change the required version.

When that happens, run `write` again locally using the original base version and commit the resulting metadata changes.

### Recommended workflow

For a releasable pull request:

```sh
# 1. Determine the version currently on main.
git show origin/main:VERSION

# 2. Prepare VERSION and CHANGELOG.md using the final PR title.
python scripts/release-version.py write \
  --base-version 0.2.2 \
  --title "feat(normalize): add canonical address normalization"

# 3. Review the generated release metadata.
git diff -- VERSION CHANGELOG.md

# 4. Run repository validation.
make check

# 5. Commit VERSION and CHANGELOG.md together with the functional change.
git add .
git commit
```

After the protected pull request is merged, release automation observes the new `VERSION` on `main` and creates the corresponding tag and GitHub Release.

`release-version.py` itself does **not** create commits, tags, pushes, or GitHub Releases. Its responsibility is limited to preparing and validating the semantic version and changelog metadata carried by the pull request.

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
