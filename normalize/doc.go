// Package normalize provides deterministic, infrastructure-independent
// canonicalization and formatting primitives for text, identifiers, contact
// values, jurisdiction-aware values, and reusable structured fields.
//
// The package is the current Go implementation of normalization semantics that
// may also be specified under spec/ and exercised through shared conformance
// vectors. Consumer-specific persistence, transport, UI, and workflow behavior
// remain outside this package.
package normalize

// RulesVersion identifies the persisted canonicalization behavior implemented
// by this release. Consumers may store it with fingerprints that must be
// recomputed when normalization semantics change.
const RulesVersion = "normalize/v1"
