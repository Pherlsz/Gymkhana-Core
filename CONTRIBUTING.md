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

The functional pull request must include its own `VERSION` and `CHANGELOG.md` updates when the Go implementation release changes. Prepare them before the final commit:

```sh
python scripts/release-version.py write \
  --base-version <version-on-main> \
  --title "<final-pr-title>"
```

CI runs the same calculation in check mode. A version mismatch or missing changelog section blocks the pull request. After a protected PR merge changes `VERSION`, the release workflow creates the tag and GitHub Release directly from `main`.

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
