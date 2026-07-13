// Package normalize provides deterministic, infrastructure-independent
// canonicalization for text, identifiers, and common contact values.
package normalize

// RulesVersion identifies the persisted canonicalization behavior implemented
// by this release. Consumers may store it with fingerprints that must be
// recomputed when normalization semantics change.
const RulesVersion = "normalize/v1"
