// Package normalize provides deterministic, infrastructure-independent
// canonicalization for text, identifiers, contact values, and profile field
// formatters used by forms, mass import, and manual create.
package normalize

// RulesVersion identifies the persisted canonicalization behavior implemented
// by this release. Consumers may store it with fingerprints that must be
// recomputed when normalization semantics change.
const RulesVersion = "normalize/v1"
