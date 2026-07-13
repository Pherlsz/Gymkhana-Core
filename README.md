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

Releases are manual and grouped by coherent capabilities. `v0.1.0` established the private module and release foundation; `v0.2.0` adds the first reusable public packages.
