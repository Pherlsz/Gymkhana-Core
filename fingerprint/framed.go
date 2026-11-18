package fingerprint

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"
	"io"
	"unicode/utf8"
)

const MaxNamespaceBytes = 128

var frameMagic = [...]byte{'C', 'F', 'P', FrameVersion}

// SumFramed fingerprints an ordered sequence of exact byte parts using the
// portable Core FingerPrint frame. Part boundaries and namespaces are included
// in the hash input, preventing ambiguous concatenation and cross-domain reuse.
func SumFramed(namespace string, parts ...[]byte) (Digest, error) {
	if err := validateFrame(namespace, len(parts)); err != nil {
		return Digest{}, err
	}
	h := sha256.New()
	var scalar [8]byte
	writeFrameHeader(h, &scalar, namespace, len(parts))
	for _, value := range parts {
		binary.BigEndian.PutUint64(scalar[:], uint64(len(value)))
		_, _ = h.Write(scalar[:])
		_, _ = h.Write(value)
	}
	var digest Digest
	h.Sum(digest[:0])
	return digest, nil
}

// SumFramedText is SumFramed for exact UTF-8 string bytes. It performs no
// Unicode or application-level canonicalization and rejects invalid UTF-8.
func SumFramedText(namespace string, parts ...string) (Digest, error) {
	if err := validateFrame(namespace, len(parts)); err != nil {
		return Digest{}, err
	}
	for _, part := range parts {
		if !utf8.ValidString(part) {
			return Digest{}, validationError(CodeInvalidUTF8)
		}
	}
	h := sha256.New()
	var scalar [8]byte
	writeFrameHeader(h, &scalar, namespace, len(parts))
	for _, value := range parts {
		binary.BigEndian.PutUint64(scalar[:], uint64(len(value)))
		_, _ = h.Write(scalar[:])
		_, _ = io.WriteString(h, value)
	}
	var digest Digest
	h.Sum(digest[:0])
	return digest, nil
}

func validateFrame(namespace string, partCount int) error {
	if !validNamespace(namespace) {
		return validationError(CodeInvalidNamespace)
	}
	if uint64(partCount) > uint64(^uint32(0)) {
		return validationError(CodeTooManyParts)
	}
	return nil
}

func writeFrameHeader(h hash.Hash, scalar *[8]byte, namespace string, partCount int) {
	_, _ = h.Write(frameMagic[:])
	binary.BigEndian.PutUint16(scalar[:2], uint16(len(namespace)))
	_, _ = h.Write(scalar[:2])
	_, _ = io.WriteString(h, namespace)
	binary.BigEndian.PutUint32(scalar[:4], uint32(partCount))
	_, _ = h.Write(scalar[:4])
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
