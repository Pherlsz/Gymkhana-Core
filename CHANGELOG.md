# Changelog

All notable changes to this project will be documented in this file.

The project follows Semantic Versioning. Breaking changes remain explicit before `v1.0.0`.

## [0.2.1] - 2026-07-13

### Fixed

- Updated `golang.org/x/text` to `v0.40.0` for current Unicode normalization and malformed-input fixes.
- Removed nonexistent custom labels from Dependabot configuration so scheduled updates open without warnings.

## [0.2.0] - 2026-07-13

### Added

- `normalize` package for Unicode-aware text, CPF, email, Brazilian phone, digit, and alphanumeric canonicalization.
- `civiltime` package with immutable `CivilDate` and `YearMonth` values, strict parsing, canonical serialization, comparison, leap-year rules, and month arithmetic.
- Table-driven unit tests and bounded fuzz targets for normalization and civil-time round trips.
- In-PR semantic version preparation with automatic GitHub Releases after merge.

### Changed

- `make fuzz-smoke` executes every declared fuzz target independently and is part of `make check`.
- Release versions and changelog entries are carried by the functional pull request instead of a generated release pull request.

## [0.1.0] - 2026-07-13

### Added

- Milestone 0 repository bootstrap.
- Pinned Go quality and security tooling.
- CI, security, and release workflow foundations.
- Contribution, security, and repository-boundary documentation.
