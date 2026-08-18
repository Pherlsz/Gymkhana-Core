package normalize

import (
	"strings"
	"unicode"
)

// CanonicalCPF strips accepted formatting, validates the Brazilian CPF check
// digits, and returns exactly eleven ASCII digits.
func CanonicalCPF(value string) (string, error) {
	if !containsOnly(value, func(r rune) bool {
		return (r >= '0' && r <= '9') || r == '.' || r == '-' || unicode.IsSpace(r)
	}) {
		return "", validationError(KindCPF, CodeInvalidFormat)
	}

	digits := Digits(value)
	if digits == "" {
		return "", validationError(KindCPF, CodeEmpty)
	}
	if len(digits) != 11 {
		return "", validationError(KindCPF, CodeInvalidLength)
	}
	if allBytesEqual(digits) || cpfCheckDigit(digits[:9], 10) != digits[9]-'0' || cpfCheckDigit(digits[:10], 11) != digits[10]-'0' {
		return "", validationError(KindCPF, CodeInvalidChecksum)
	}

	return digits, nil
}

// CanonicalCNPJ strips accepted formatting, uppercases ASCII letters, validates
// the Receita Federal check digits, and returns fourteen characters. The first
// twelve may be digits or A–Z; the last two are always digits. Numeric CNPJs
// remain valid because digit values are ASCII minus 48.
func CanonicalCNPJ(value string) (string, error) {
	if !containsOnly(value, cnpjRune) {
		return "", validationError(KindCNPJ, CodeInvalidFormat)
	}

	body := cnpjBody(value)
	if body == "" {
		return "", validationError(KindCNPJ, CodeEmpty)
	}
	if len(body) != 14 {
		return "", validationError(KindCNPJ, CodeInvalidLength)
	}
	if body[12] < '0' || body[12] > '9' || body[13] < '0' || body[13] > '9' {
		return "", validationError(KindCNPJ, CodeInvalidFormat)
	}
	if allBytesEqual(body) || cnpjCheckDigit(body[:12]) != body[12] || cnpjCheckDigit(body[:13]) != body[13] {
		return "", validationError(KindCNPJ, CodeInvalidChecksum)
	}

	return body, nil
}

func cnpjRune(r rune) bool {
	return (r >= '0' && r <= '9') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= 'a' && r <= 'z') ||
		r == '.' || r == '/' || r == '-' ||
		unicode.IsSpace(r)
}

func cnpjBody(value string) string {
	var builder strings.Builder
	builder.Grow(14)
	for i := 0; i < len(value); i++ {
		r := value[i]
		switch {
		case r >= '0' && r <= '9':
			builder.WriteByte(r)
		case r >= 'A' && r <= 'Z':
			builder.WriteByte(r)
		case r >= 'a' && r <= 'z':
			builder.WriteByte(r - 'a' + 'A')
		}
	}
	return builder.String()
}

func cnpjCheckDigit(prefix string) byte {
	sum := 0
	weight := 2
	for i := len(prefix) - 1; i >= 0; i-- {
		sum += (int(prefix[i]) - '0') * weight
		weight++
		if weight > 9 {
			weight = 2
		}
	}
	remainder := sum % 11
	if remainder < 2 {
		return '0'
	}
	return byte('0' + (11 - remainder))
}

func cpfCheckDigit(prefix string, weight int) byte {
	sum := 0
	for i := 0; i < len(prefix); i++ {
		sum += int(prefix[i]-'0') * (weight - i)
	}

	remainder := sum % 11
	if remainder < 2 {
		return 0
	}
	return byte(11 - remainder)
}

func allBytesEqual(value string) bool {
	for i := 1; i < len(value); i++ {
		if value[i] != value[0] {
			return false
		}
	}
	return true
}

func containsOnly(value string, allowed func(rune) bool) bool {
	for _, r := range value {
		if !allowed(r) {
			return false
		}
	}
	return true
}
