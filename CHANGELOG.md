# Changelog

All notable changes to this project will be documented in this file.

The project follows Semantic Versioning. Breaking changes remain explicit before `v1.0.0`.

## [Unreleased]

### Added

- `normalize` package for Unicode-aware text, CPF, email, Brazilian phone, digit, and alphanumeric canonicalization.
- `civiltime` package with immutable `CivilDate` and `YearMonth` values, strict parsing, canonical text/JSON serialization, comparison, leap-year rules, and month arithmetic.
- Table-driven unit tests and bounded fuzz targets for normalization and civil-time round trips.

### Changed

- `make fuzz-smoke` now executes every declared fuzz target independently and is part of `make check`.

## [0.1.0] - 2026-07-13

### Added

- Milestone 0 repository bootstrap.
- Pinned Go quality and security tooling.
- CI, security, and manual release workflows.
- Contribution, security, and repository-boundary documentation.
