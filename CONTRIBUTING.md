# Contributing

## Before starting

Read `docs/ORCHESTRATION.md` and `spec/README.md`.

Gymkhana Core accepts coherent reusable contracts and deterministic/infrastructure-independent algorithms. It is language-neutral, product-neutral, provider-neutral, and jurisdiction-neutral.

Work may originate from Gymkhana Database, another service, a future AI Assistant, or an independently approved Core capability. When a consumer motivates the work, link its issue or requirement for context. The consumer is not the source of truth for Core semantics.

## Language

Code, identifiers, comments, documentation, commits, pull requests, changelog entries, workflows, specification prose, and CI messages are written in English.

## Branches, milestones, and commits

Create one short-lived branch from `main` for each active workstream, for example:

```text
feat/core-normalization
feat/assistant-contracts
fix/civil-date-validation
refactor/fingerprint-canonicalization
```

Open one draft pull request for that branch. Do not create helper PRs for formatting, diagnostics, validation, export, or squash operations.

A Core milestone is appropriate only when a Core-owned delivery has multiple independently executable issues. Consumer milestones may link Core work but do not own Core architecture or versioning.

Use Conventional Commits for the final pull request title and squash commit:

```text
feat(normalize): add canonical identifier normalization
fix(civiltime): reject invalid leap days
feat(assistant)!: replace the public tool contract
```

Version impact for the current Go implementation is derived from that final title:

- `fix` produces a patch release;
- `feat` produces a minor release;
- breaking changes produce a minor release before `1.0.0` and a major release afterwards;
- non-releasable types such as `chore`, `docs`, and `test` keep the current version.

The Core specification version in `spec/VERSION` is separate from the current Go implementation version in `VERSION`.

## Specification changes

Portable behavior is specified under `spec/` and demonstrated through `conformance/` vectors.

When a change affects cross-language observable semantics:

1. update the relevant specification;
2. add or update shared conformance vectors;
3. update the current implementation;
4. update language-local tests/fuzz/property coverage;
5. determine whether `spec/VERSION` must change;
6. document consumer migration impact when behavior is breaking.

Do not make Go-specific API ergonomics normative unless they represent a true language-neutral semantic requirement.

## Jurisdiction and standards

Country-specific rules must be explicitly namespaced in the specification. Brazil is one jurisdiction, not a Core default.

Country, jurisdiction, language, locale, currency, and timezone must not be silently conflated.

Application-specific aliases, migration cleanup mappings, and dirty-data heuristics should remain configurable/consumer-owned unless they form a stable reusable contract or dataset.

International standards belong under `spec/standards/` only when their contract and maintenance/versioning model are understood.

## Release preparation

Every functional pull request must carry its own version and changelog updates when the current implementation change is releasable.

Versioning follows Semantic Versioning and is derived from the **final pull request title** using Conventional Commits.

### Version rules

Given a base version such as `0.2.2`:

| Pull request title | Version impact | Result |
| --- | --- | --- |
| `fix(normalize): correct CPF formatting` | Patch | `0.2.3` |
| `feat(normalize): add address normalization` | Minor | `0.3.0` |
| `feat(normalize)!: replace address contract` | Breaking | `0.3.0` before `1.0.0` |
| `feat(api)!: replace public API` | Breaking | `2.0.0` from `1.x.x` |
| `chore: update tooling` | None | unchanged |
| `docs: improve documentation` | None | unchanged |
| `test: add normalization coverage` | None | unchanged |
| `refactor: reorganize internals` | None | unchanged |

A pull request is also treated as breaking when its body contains:

```text
BREAKING CHANGE:
```

Breaking changes increment the minor version while the implementation is below `1.0.0`. Starting from `1.0.0`, they increment the major version.

### Preparing a release locally

Before the final commit, determine the version currently present on the pull request base branch, normally `main`.

For example, if `main` contains `0.2.2` and the final pull request title is:

```text
feat(normalize): add canonical address normalization
```

run:

```sh
python scripts/release-version.py write \
  --base-version 0.2.2 \
  --title "feat(normalize): add canonical address normalization"
```

The command prints the calculated version, such as `0.3.0`, and updates release metadata.

For a releasable change, `write` updates:

- `VERSION` with the calculated semantic version;
- `CHANGELOG.md` with a release section for that version.

Review the generated changes before committing:

```sh
git diff -- VERSION CHANGELOG.md
```

The generated `VERSION` and `CHANGELOG.md` changes belong to the same functional pull request. Do not create a separate release-preparation pull request.

### Base version

`--base-version` is always the implementation version from the pull request base branch **before the current pull request is applied**.

For example:

```text
main VERSION:                    0.2.2
feature branch VERSION:          0.3.0
--base-version for the same PR:  0.2.2
```

Do not use the already-bumped feature-branch version as the base for the same pull request.

When needed, inspect the version on `main` directly:

```sh
git show origin/main:VERSION
```

### Pull request title consistency

The title passed to `release-version.py write` must match the final semantic meaning of the pull request title.

If a pull request prepared as:

```text
feat(normalize): add canonical address normalization
```

is later changed to:

```text
fix(normalize): correct canonical address normalization
```

the required version changes. Run `write` again using the same original base version and the new final title before marking the pull request ready for review.

### Breaking changes

Breaking changes can be declared with `!` in the title:

```sh
python scripts/release-version.py write \
  --base-version 0.2.2 \
  --title "feat(normalize)!: replace the public address contract"
```

or through the pull request body:

```text
BREAKING CHANGE: callers must migrate to the new address representation.
```

When the breaking declaration exists only in the body, pass the same body locally:

```sh
python scripts/release-version.py write \
  --base-version 0.2.2 \
  --title "feat(normalize): replace the public address contract" \
  --body "BREAKING CHANGE: callers must migrate to the new address representation."
```

### Changelog generation

The current script derives the changelog category from the change type:

| Change | Changelog section |
| --- | --- |
| `feat` | `Added` |
| `fix` | `Fixed` |
| Breaking change | `Changed` |

The current date is used by default. A specific date can be supplied with `--date` when necessary.

Do not create another section for the same version after the script has generated one. If a release section already exists and requires richer release notes, edit that section deliberately and keep the version marker intact.

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
3. `CHANGELOG.md` contains the corresponding release section when a new release is required.

A successful validation prints information similar to:

```json
{"base":"0.2.2","expected":"0.3.0","actual":"0.3.0"}
```

`check` does not prepare release metadata and should not replace `write` when the files still need to be generated.

### CI validation

CI performs the same version calculation using:

- the `VERSION` value from the pull request base commit;
- the current pull request title;
- the current pull request body.

A mismatch between the calculated version and committed `VERSION`, or a missing changelog section for a releasable change, blocks the pull request.

### Recommended release workflow

```sh
# 1. Determine the implementation version currently on main.
git show origin/main:VERSION

# 2. Prepare VERSION and CHANGELOG.md using the final PR title.
python scripts/release-version.py write \
  --base-version 0.2.2 \
  --title "feat(normalize): add canonical address normalization"

# 3. Review generated release metadata.
git diff -- VERSION CHANGELOG.md

# 4. Run repository validation.
make check

# 5. Commit release metadata with the functional change.
git add .
git commit
```

After the protected pull request is merged, release automation observes the new `VERSION` on `main` and creates the corresponding tag and GitHub Release.

`release-version.py` itself does **not** create commits, tags, pushes, or GitHub Releases. Its responsibility is limited to preparing and validating implementation release metadata carried by the pull request.

Future language ports may use independent ecosystem package versions while declaring the Core specification version they implement.

## Local checks and conformance

During development, run targeted commands locally. Before marking a PR ready for review, run:

```sh
make check
```

Shared Go conformance can be run directly with:

```sh
make conformance
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

- objective and reusable scope;
- originating consumer requirement when applicable;
- specification/conformance impact;
- public API impact;
- determinism, complexity, and limits;
- tests, fuzzing, property tests, and benchmarks;
- consumer impact;
- security/privacy implications;
- migration notes for breaking public behavior;
- expected Actions impact when workflow frequency, jobs, timeout, matrix, artifacts, or runner type changes.

Do not merge infrastructure dependencies, localized product text, HTTP status codes, SQL, provider SDK types, credentials, or arbitrary execution facilities into Core.

## Public API and multi-language design

Keep public APIs small. Do not create placeholder packages or speculative abstractions.

A language implementation may be idiomatic. Cross-language consistency applies to semantics and serialized contracts, not to forcing identical method names or class/module shapes across ecosystems.

Do not build every language port preemptively. Add a new implementation when a real consumer or portability requirement exists, and make it consume the shared conformance vectors.

## Security

Never commit secrets, private production data, prompts containing real sensitive data, signed URLs, credentials, or raw provider payloads. Follow `SECURITY.md` for vulnerability reports.
