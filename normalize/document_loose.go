package normalize

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

func joinIdentifierTokens(value string) string {
	var builder strings.Builder
	for _, token := range scanIdentifierTokens(value) {
		if token.kind == tokenLetters && isIdentifierLabel(token.text) {
			continue
		}
		builder.WriteString(token.text)
	}
	return builder.String()
}

func canonicalPassport(kind DocumentKind, value string) (string, error) {
	value = DisplayText(value)
	if value == "" {
		return "", validationError(ValueKind(kind), CodeEmpty)
	}
	if !containsOnly(value, passportRune) {
		return "", validationError(ValueKind(kind), CodeInvalidFormat)
	}

	canonical := Alphanumeric(value)
	if len(canonical) != 8 {
		return "", validationError(ValueKind(kind), CodeInvalidLength)
	}
	if !isASCIILetter(rune(canonical[0])) || !isASCIILetter(rune(canonical[1])) {
		return "", validationError(ValueKind(kind), CodeInvalidFormat)
	}
	for i := 2; i < 8; i++ {
		if canonical[i] < '0' || canonical[i] > '9' {
			return "", validationError(ValueKind(kind), CodeInvalidFormat)
		}
	}
	return canonical, nil
}

func passportRune(r rune) bool {
	return unicode.IsLetter(r) || (r >= '0' && r <= '9') || unicode.IsSpace(r)
}

func isASCIILetter(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

func canonicalLooseDocument(kind DocumentKind, value string) (string, error) {
	value = DisplayText(value)
	if value == "" {
		return "", validationError(ValueKind(kind), CodeEmpty)
	}
	if !containsOnly(value, looseIdentifierRune) {
		return "", validationError(ValueKind(kind), CodeInvalidFormat)
	}

	canonical := joinIdentifierTokens(strings.ToUpper(norm.NFC.String(value)))
	if canonical == "" {
		return "", validationError(ValueKind(kind), CodeEmpty)
	}
	if !hasDigit(canonical) {
		return "", validationError(ValueKind(kind), CodeInvalidFormat)
	}

	minLen, maxLen := 4, 24
	switch kind {
	case DocumentBirthCertificate, DocumentMarriageCertificate:
		minLen, maxLen = 6, 40
	case DocumentCitizenCard:
		minLen, maxLen = 8, 24
	}
	if len(canonical) < minLen || len(canonical) > maxLen {
		return "", validationError(ValueKind(kind), CodeInvalidLength)
	}
	return canonical, nil
}

func looseIdentifierRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) || r == '.' || r == '-' || r == '/'
}

func hasDigit(value string) bool {
	for i := 0; i < len(value); i++ {
		if value[i] >= '0' && value[i] <= '9' {
			return true
		}
	}
	return false
}
