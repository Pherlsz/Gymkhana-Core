# Gymkhana Core

Private Go module containing deterministic, infrastructure-independent contracts and algorithms reused by Gymkhana Database.

## Status

Milestone 0 bootstrap. Public domain packages are intentionally not created until their contracts are implemented in Milestone 1.

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

Gymkhana Database consumes exact tagged versions. Permanent `replace` directives, copied source, submodules, and branch dependencies are not supported.

## Development

Read [CONTRIBUTING.md](CONTRIBUTING.md) and [docs/ORCHESTRATION.md](docs/ORCHESTRATION.md) before changing public contracts.

## Releases

Releases are manual and grouped by coherent capability. The first `v0.1.0` release will only be created after the module has CI, documentation, a verified release workflow, and minimum real reusable content.
