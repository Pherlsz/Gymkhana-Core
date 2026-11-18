package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"unicode/utf8"
)

// Digest is a SHA-256 fingerprint.
type Digest [sha256.Size]byte

// Sum fingerprints value exactly as supplied.
func Sum(value []byte) Digest {
	return Digest(sha256.Sum256(value))
}

// SumText fingerprints the exact UTF-8 bytes of value without normalization.
// Invalid UTF-8 is rejected so the operation has the same semantics in
// runtimes whose string types cannot represent arbitrary byte sequences.
func SumText(value string) (Digest, error) {
	if !utf8.ValidString(value) {
		return Digest{}, validationError(CodeInvalidUTF8)
	}
	hash := sha256.New()
	_, _ = io.WriteString(hash, value)
	var digest Digest
	hash.Sum(digest[:0])
	return digest, nil
}

// Parse decodes a 64-character hexadecimal SHA-256 digest. Uppercase input is
// accepted; Digest.String always returns the canonical lowercase form.
func Parse(value string) (Digest, error) {
	if len(value) != sha256.Size*2 {
		return Digest{}, validationError(CodeInvalidLength)
	}

	var digest Digest
	if err := decodeHexInto(&digest, value); err != nil {
		return Digest{}, err
	}
	return digest, nil
}

// String returns the canonical lowercase hexadecimal representation.
func (d Digest) String() string {
	var buf [sha256.Size * 2]byte
	hex.Encode(buf[:], d[:])
	return string(buf[:])
}

// Bytes returns a copy of the digest bytes.
func (d Digest) Bytes() []byte {
	value := make([]byte, len(d))
	copy(value, d[:])
	return value
}

// AppendText implements encoding.TextAppender using canonical lowercase hex.
func (d Digest) AppendText(b []byte) ([]byte, error) {
	var buf [sha256.Size * 2]byte
	hex.Encode(buf[:], d[:])
	return append(b, buf[:]...), nil
}

// MarshalText implements encoding.TextMarshaler using canonical lowercase hex.
func (d Digest) MarshalText() ([]byte, error) {
	var buf [sha256.Size * 2]byte
	hex.Encode(buf[:], d[:])
	return buf[:], nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (d *Digest) UnmarshalText(text []byte) error {
	if len(text) != sha256.Size*2 {
		return validationError(CodeInvalidLength)
	}
	var parsed Digest
	if _, err := hex.Decode(parsed[:], text); err != nil {
		return validationError(CodeInvalidFormat)
	}
	*d = parsed
	return nil
}

func decodeHexInto(dst *Digest, value string) error {
	for i := 0; i < sha256.Size; i++ {
		hi, okHi := fromHexChar(value[i*2])
		lo, okLo := fromHexChar(value[i*2+1])
		if !okHi || !okLo {
			return validationError(CodeInvalidFormat)
		}
		dst[i] = (hi << 4) | lo
	}
	return nil
}

func fromHexChar(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	default:
		return 0, false
	}
}
