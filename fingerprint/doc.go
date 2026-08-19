// Package fingerprint provides deterministic SHA-256 fingerprints for exact
// bytes and explicitly framed semantic parts.
//
// The package does not normalize caller data. Consumers must canonicalize
// semantic values before fingerprinting them when representation equivalence
// matters.
package fingerprint

const (
	// Algorithm identifies the digest algorithm used by the current contract.
	Algorithm = "sha256"
	// RulesVersion identifies the portable framed-fingerprint encoding.
	RulesVersion = "fingerprint/v1"
	// FrameVersion is encoded into every framed fingerprint payload.
	FrameVersion byte = 1
)
