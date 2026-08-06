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

// MaskCPF returns a display-safe form of a canonical CPF by hiding its first
// three digits: "***.XXX.XXX-XX". The input must be exactly eleven ASCII
// digits (the output of CanonicalCPF); any other input is returned unchanged
// so callers can safely pass user-supplied text without panicking.
func MaskCPF(canonical string) string {
	if len(canonical) != 11 || !containsOnly(canonical, func(r rune) bool { return r >= '0' && r <= '9' }) {
		return canonical
	}
	return "***." + canonical[3:6] + "." + canonical[6:9] + "-" + canonical[9:11]
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
