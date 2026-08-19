package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"unicode/utf8"
)

// Digest is a SHA-256 fingerprint.
type Digest [sha256.Size]byte

// Bytes fingerprints value exactly as supplied.
func Bytes(value []byte) Digest {
	return Digest(sha256.Sum256(value))
}

// String fingerprints the exact UTF-8 bytes of value without normalization.
// Invalid UTF-8 is rejected so the operation has the same semantics in
// runtimes whose string types cannot represent arbitrary byte sequences.
func String(value string) (Digest, error) {
	if !utf8.ValidString(value) {
		return Digest{}, validationError(CodeInvalidUTF8)
	}
	return Bytes([]byte(value)), nil
}

// Parse decodes a 64-character hexadecimal SHA-256 digest. Uppercase input is
// accepted; String always returns the canonical lowercase form.
func Parse(value string) (Digest, error) {
	if len(value) != sha256.Size*2 {
		return Digest{}, validationError(CodeInvalidLength)
	}

	decoded, err := hex.DecodeString(value)
	if err != nil {
		return Digest{}, validationError(CodeInvalidFormat)
	}

	var digest Digest
	copy(digest[:], decoded)
	return digest, nil
}

// String returns the canonical lowercase hexadecimal representation.
func (d Digest) String() string {
	return hex.EncodeToString(d[:])
}

// Bytes returns a copy of the digest bytes.
func (d Digest) Bytes() []byte {
	value := make([]byte, len(d))
	copy(value, d[:])
	return value
}

// MarshalText implements encoding.TextMarshaler using canonical lowercase hex.
func (d Digest) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (d *Digest) UnmarshalText(text []byte) error {
	parsed, err := Parse(string(text))
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}
