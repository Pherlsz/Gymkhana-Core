package normalize

import (
	"net/mail"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// CanonicalEmail validates a single email address, preserves the local-part,
// and lowercases the ASCII domain. Display names and surrounding metadata are
// rejected.
func CanonicalEmail(value string) (string, error) {
	value = strings.TrimSpace(norm.NFC.String(value))
	if value == "" {
		return "", validationError(KindEmail, CodeEmpty)
	}
	for _, r := range value {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return "", validationError(KindEmail, CodeInvalidFormat)
		}
	}

	parsed, err := mail.ParseAddress(value)
	if err != nil || parsed.Name != "" || parsed.Address != value {
		return "", validationError(KindEmail, CodeInvalidFormat)
	}

	separator := strings.LastIndexByte(value, '@')
	if separator <= 0 || separator == len(value)-1 {
		return "", validationError(KindEmail, CodeInvalidFormat)
	}

	localPart := value[:separator]
	domain := value[separator+1:]
	if !validASCIIDomain(domain) {
		return "", validationError(KindEmail, CodeInvalidFormat)
	}

	return localPart + "@" + strings.ToLower(domain), nil
}

// CanonicalBrazilPhone validates a Brazilian landline or mobile number and
// returns E.164 format with the +55 country code.
func CanonicalBrazilPhone(value string) (string, error) {
	if !validPhoneCharacters(value) {
		return "", validationError(KindBrazilPhone, CodeInvalidFormat)
	}

	digits := Digits(value)
	if digits == "" {
		return "", validationError(KindBrazilPhone, CodeEmpty)
	}

	national := digits
	if strings.HasPrefix(national, "55") && (len(national) == 12 || len(national) == 13) {
		national = national[2:]
	}
	if len(national) != 10 && len(national) != 11 {
		return "", validationError(KindBrazilPhone, CodeInvalidLength)
	}
	if !validBrazilAreaCode(national[:2]) {
		return "", validationError(KindBrazilPhone, CodeInvalidAreaCode)
	}

	subscriber := national[2:]
	if len(subscriber) == 8 {
		if subscriber[0] < '2' || subscriber[0] > '5' {
			return "", validationError(KindBrazilPhone, CodeInvalidNumber)
		}
	} else if subscriber[0] != '9' {
		return "", validationError(KindBrazilPhone, CodeInvalidNumber)
	}

	return "+55" + national, nil
}

func validASCIIDomain(domain string) bool {
	if domain == "" || strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}

	labels := strings.Split(domain, ".")
	for _, label := range labels {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for i := 0; i < len(label); i++ {
			character := label[i]
			if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') || (character >= '0' && character <= '9') || character == '-' {
				continue
			}
			return false
		}
	}

	return len(domain) <= 253
}

func validPhoneCharacters(value string) bool {
	seenContent := false
	for _, r := range value {
		switch {
		case unicode.IsSpace(r):
			continue
		case r == '+':
			if seenContent {
				return false
			}
			seenContent = true
		case (r >= '0' && r <= '9') || r == '-' || r == '(' || r == ')' || r == '.':
			seenContent = true
		default:
			return false
		}
	}
	return true
}
