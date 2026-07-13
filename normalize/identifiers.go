package normalize

import "unicode"

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
