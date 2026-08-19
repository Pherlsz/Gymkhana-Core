package fingerprint

import (
	"crypto/sha256"
	"encoding/binary"
	"unicode/utf8"
)

const MaxNamespaceBytes = 128

var frameMagic = [...]byte{'C', 'F', 'P', FrameVersion}

// Framed fingerprints an ordered sequence of exact byte parts using the
// portable Core FingerPrint frame. Part boundaries and namespaces are included
// in the hash input, preventing ambiguous concatenation and cross-domain reuse.
func Framed(namespace string, parts ...[]byte) (Digest, error) {
	return framed(namespace, len(parts), func(index int) []byte {
		return parts[index]
	})
}

// FramedStrings is Framed for exact UTF-8 string bytes. It performs no Unicode
// or application-level canonicalization and rejects invalid UTF-8 strings.
func FramedStrings(namespace string, parts ...string) (Digest, error) {
	for _, part := range parts {
		if !utf8.ValidString(part) {
			return Digest{}, validationError(CodeInvalidUTF8)
		}
	}
	return framed(namespace, len(parts), func(index int) []byte {
		return []byte(parts[index])
	})
}

func framed(namespace string, partCount int, part func(index int) []byte) (Digest, error) {
	if !validNamespace(namespace) {
		return Digest{}, validationError(CodeInvalidNamespace)
	}
	if uint64(partCount) > uint64(^uint32(0)) {
		return Digest{}, validationError(CodeTooManyParts)
	}

	hash := sha256.New()
	_, _ = hash.Write(frameMagic[:])

	var scalar [8]byte
	binary.BigEndian.PutUint16(scalar[:2], uint16(len(namespace)))
	_, _ = hash.Write(scalar[:2])
	_, _ = hash.Write([]byte(namespace))

	binary.BigEndian.PutUint32(scalar[:4], uint32(partCount))
	_, _ = hash.Write(scalar[:4])

	for index := 0; index < partCount; index++ {
		value := part(index)
		binary.BigEndian.PutUint64(scalar[:], uint64(len(value)))
		_, _ = hash.Write(scalar[:])
		_, _ = hash.Write(value)
	}

	var digest Digest
	copy(digest[:], hash.Sum(nil))
	return digest, nil
}

func validNamespace(value string) bool {
	if len(value) == 0 || len(value) > MaxNamespaceBytes {
		return false
	}
	if value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for index := 1; index < len(value); index++ {
		character := value[index]
		if character >= 'a' && character <= 'z' {
			continue
		}
		if character >= '0' && character <= '9' {
			continue
		}
		switch character {
		case '.', '_', ':', '/', '-':
			continue
		default:
			return false
		}
	}
	return true
}
