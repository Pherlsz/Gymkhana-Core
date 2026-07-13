package normalize

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// DisplayText returns NFC-normalized text with leading/trailing whitespace
// removed and every Unicode whitespace run collapsed to a single ASCII space.
// Invalid UTF-8 sequences become separators so malformed input cannot merge
// adjacent Unicode compatibility characters into unstable search forms.
func DisplayText(value string) string {
	if value == "" {
		return ""
	}

	value = strings.ToValidUTF8(value, " ")
	return strings.Join(strings.Fields(norm.NFC.String(value)), " ")
}

// SearchText returns a deterministic case- and accent-insensitive search form.
// Letters and digits are preserved; punctuation and whitespace become separators.
func SearchText(value string) string {
	value = DisplayText(value)
	if value == "" {
		return ""
	}

	decomposed := norm.NFKD.String(value)
	var builder strings.Builder
	builder.Grow(len(decomposed))
	pendingSpace := false

	for _, r := range decomposed {
		if unicode.Is(unicode.Mn, r) {
			continue
		}

		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if pendingSpace && builder.Len() > 0 {
				builder.WriteByte(' ')
			}
			builder.WriteRune(unicode.ToLower(r))
			pendingSpace = false
			continue
		}

		if builder.Len() > 0 {
			pendingSpace = true
		}
	}

	return norm.NFC.String(builder.String())
}

// Digits keeps ASCII decimal digits only and preserves their original order,
// including leading zeros.
func Digits(value string) string {
	var builder strings.Builder
	builder.Grow(len(value))

	for i := 0; i < len(value); i++ {
		if value[i] >= '0' && value[i] <= '9' {
			builder.WriteByte(value[i])
		}
	}

	return builder.String()
}

// Alphanumeric keeps Unicode letters and digits, removes separators, and
// uppercases letters. Leading zeroes are preserved.
func Alphanumeric(value string) string {
	value = norm.NFC.String(strings.ToValidUTF8(value, " "))
	var builder strings.Builder
	builder.Grow(len(value))

	for _, r := range value {
		if unicode.IsLetter(r) {
			builder.WriteRune(unicode.ToUpper(r))
			continue
		}
		if unicode.IsDigit(r) {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}
