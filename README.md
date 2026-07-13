# Gymkhana Core

Private Go module containing deterministic, infrastructure-independent contracts and algorithms reused by Gymkhana Database.

## Status

Milestone 1 foundations. The first public packages are `normalize` and `civiltime`; the first release containing them is planned as `v0.2.0`.

## Available packages

### `normalize`

Deterministic normalization for:

- display and search text;
- Unicode letters and digits;
- CPF validation and canonicalization;
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

Useful commands:

```sh
make format
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

After `v0.2.0` is released, Gymkhana Database pins it explicitly:

```sh
go get github.com/Pherlsz/Gymkhana-Core@v0.2.0
```

Permanent `replace` directives, copied source, submodules, and branch dependencies are not supported.

## Development

Read [CONTRIBUTING.md](CONTRIBUTING.md) and [docs/ORCHESTRATION.md](docs/ORCHESTRATION.md) before changing public contracts.

## Releases

Release Please derives versions from the Conventional Commit used as the final squash or merge title:

- `fix` increments the patch version;
- `feat` increments the minor version;
- a `!` or `BREAKING CHANGE` increments the minor version before `1.0.0` and the major version afterwards;
- `chore`, `docs`, `test`, and other non-releasable commits do not create a version bump by themselves.

A push to `main` validates the repository and creates or updates a release pull request. Merging that release pull request updates `CHANGELOG.md`, creates the `vX.Y.Z` tag, and publishes the GitHub Release. Releases remain grouped deliberately: ordinary feature merges prepare the release pull request but do not publish immediately.
