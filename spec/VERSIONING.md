# Specification versioning

`spec/VERSION` versions the language-neutral Core semantics independently from any language package release.

The specification uses `MAJOR.MINOR` versions.

## Before `1.0`

The Foundation specification is still stabilizing. During `0.x`:

- additive semantics and new operation families increment `MINOR` when they require a new conformance target;
- incompatible observable behavior also increments `MINOR` and must be called out explicitly as breaking for affected operations;
- editorial clarification that does not change observable behavior does not require a version bump.

Implementations must declare the exact `0.x` specification version they conform to; a newer `0.x` version must not be assumed compatible automatically.

## From `1.0`

After the specification is declared stable:

- `MAJOR` increments for incompatible semantic changes, removals, or reinterpretations of existing operations;
- `MINOR` increments for backward-compatible additions such as new operations, optional structured fields, or new jurisdiction modules;
- editorial clarification with no observable behavior change does not increment the version.

## Conformance suites

Versioned vectors live in `conformance/v<MAJOR.MINOR>/`.

An implementation claiming specification version `X.Y` must pass all applicable suites for `conformance/vX.Y/`. Supporting multiple specification versions is an implementation choice; Core does not require every package release to retain runners for every historical specification.

Existing conformance vectors are immutable in meaning. Fixing a typo or invalid fixture is permitted only when the previous vector could not correctly represent the already-documented semantics. Any real change in expected behavior requires a new specification version.

## Implementation versions

Language package versions remain independent. For example, Go `0.6.1` and TypeScript `0.4.0` may both conform to Core Spec `1.2`.

Package managers and implementation SemVer describe implementation delivery; `spec/VERSION` describes cross-language semantic compatibility.
