# Changelog

All notable changes to this project will be documented in this file.

The project follows Semantic Versioning. Breaking changes remain explicit before `v1.0.0`.

## [0.3.0] - 2026-08-18

### Added

- `FormatStreet`, `FormatHouseNumber`, `FormatBlock`, `FormatApartment`, `FormatCEP`, `FormatUF`, `FormatCity`, `FormatNeighborhood`, `FormatCountry`, and `FormatNationality` for address and origin slots.
- `FormatPersonName`, `FormatGender`, `FormatMaritalStatus`, `FormatBloodType`, and `FormatMembershipType` for profile civil fields.
- `FormatTeam`, `FormatSector`, `FormatClub`, `FormatHealthPlan`, `FormatCard`, `FormatCollection`, `FormatAnimal`, `FormatVehiclePlate`, `FormatVehicleModel`, `FormatVehicleColor`, and `FormatVehicleYear` for gincana catalogs. Empty, nenhum, and s/n stay empty. Formatters are idempotent where a canonical form exists.
- `CanonicalDocument` and `FormatDocument` for catalog document kinds, with check digits where they exist and UF folded into the persisted identifier (`114458/RS`) so state is not a separate column.
- Document kind `pis` (PIS/PASEP/NIT/NIS): 11 digits with the Caixa modulo-11 weights `3,2,9,8,7,6,5,4,3,2`. It is a separate type from CTPS and is not inferred from an unlabeled number.
- `IdentifyDocument` infers a catalog kind from an unlabeled number only when a single strong validator matches, so weak shapes such as RG or OAB are not guessed.
- `CanonicalCNPJ` and document kind `cnpj` accept the Receita Federal alphanumeric CNPJ (12 letters or digits plus two numeric check digits) while remaining compatible with existing numeric CNPJs.

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
