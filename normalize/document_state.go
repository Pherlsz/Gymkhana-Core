package normalize

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

type stateDocumentOptions struct {
	requireState bool
	allowX       bool
	allowSeries  bool
	minDigits    int
	maxDigits    int
}

func stateOptions(kind DocumentKind) stateDocumentOptions {
	switch kind {
	case DocumentRG, DocumentIdentidade:
		return stateDocumentOptions{allowX: true, minDigits: 5, maxDigits: 12}
	case DocumentCTPS:
		return stateDocumentOptions{allowSeries: true, minDigits: 4, maxDigits: 12}
	default:
		return stateDocumentOptions{requireState: true, minDigits: 3, maxDigits: 10}
	}
}

func canonicalStateDocument(kind DocumentKind, value string) (string, error) {
	parsed, err := parseStateDocument(kind, value, stateOptions(kind))
	if err != nil {
		return "", err
	}
	canonical := parsed.number
	if parsed.series != "" {
		canonical += "-" + parsed.series
	}
	if parsed.state != "" {
		canonical += "/" + parsed.state
	}
	return canonical, nil
}

type parsedStateDocument struct {
	number string
	series string
	state  string
}

func parseStateDocument(kind DocumentKind, value string, options stateDocumentOptions) (parsedStateDocument, error) {
	value = DisplayText(value)
	if value == "" {
		return parsedStateDocument{}, validationError(ValueKind(kind), CodeEmpty)
	}
	if !containsOnly(value, stateIdentifierRune) {
		return parsedStateDocument{}, validationError(ValueKind(kind), CodeInvalidFormat)
	}

	tokens := scanIdentifierTokens(strings.ToUpper(norm.NFC.String(value)))
	var states []string
	var numbers []string
	xCount := 0

	for _, token := range tokens {
		switch token.kind {
		case tokenLetters:
			if isIdentifierLabel(token.text) {
				continue
			}
			if validBrazilState(token.text) {
				states = append(states, token.text)
				continue
			}
			if options.allowX && token.text == "X" {
				xCount++
				continue
			}
			if len(token.text) == 2 {
				return parsedStateDocument{}, validationError(ValueKind(kind), CodeInvalidState)
			}
			return parsedStateDocument{}, validationError(ValueKind(kind), CodeInvalidFormat)
		case tokenDigits:
			numbers = append(numbers, token.text)
		}
	}

	if xCount > 1 {
		return parsedStateDocument{}, validationError(ValueKind(kind), CodeInvalidFormat)
	}
	if len(states) > 1 {
		return parsedStateDocument{}, validationError(ValueKind(kind), CodeInvalidFormat)
	}
	if options.requireState && len(states) == 0 {
		return parsedStateDocument{}, validationError(ValueKind(kind), CodeMissingState)
	}

	number, series, err := assignStateNumbers(kind, numbers, options)
	if err != nil {
		return parsedStateDocument{}, err
	}
	if xCount == 1 {
		number += "X"
	}

	digitLen := digitCount(number)
	if digitLen < options.minDigits || digitLen > options.maxDigits {
		return parsedStateDocument{}, validationError(ValueKind(kind), CodeInvalidLength)
	}
	if options.allowSeries && series != "" && (len(series) < 1 || len(series) > 8) {
		return parsedStateDocument{}, validationError(ValueKind(kind), CodeInvalidLength)
	}

	parsed := parsedStateDocument{number: number, series: series}
	if len(states) == 1 {
		parsed.state = states[0]
	}
	return parsed, nil
}

func assignStateNumbers(kind DocumentKind, numbers []string, options stateDocumentOptions) (string, string, error) {
	if len(numbers) == 0 {
		return "", "", validationError(ValueKind(kind), CodeInvalidFormat)
	}
	if options.allowSeries && len(numbers) >= 2 {
		return strings.Join(numbers[:len(numbers)-1], ""), numbers[len(numbers)-1], nil
	}
	return strings.Join(numbers, ""), "", nil
}

const (
	tokenLetters = 'L'
	tokenDigits  = 'D'
)

type identifierToken struct {
	kind rune
	text string
}

func scanIdentifierTokens(value string) []identifierToken {
	var tokens []identifierToken
	var builder strings.Builder
	kind := rune(0)

	flush := func() {
		if builder.Len() == 0 {
			return
		}
		tokens = append(tokens, identifierToken{kind: kind, text: builder.String()})
		builder.Reset()
		kind = 0
	}

	for _, r := range value {
		switch {
		case r >= '0' && r <= '9':
			if kind == tokenLetters {
				flush()
			}
			kind = tokenDigits
			builder.WriteByte(byte(r))
		case r >= 'A' && r <= 'Z':
			if kind == tokenDigits {
				flush()
			}
			kind = tokenLetters
			builder.WriteByte(byte(r))
		case r == '.' || r == ':':
			continue
		default:
			flush()
		}
	}
	flush()
	return tokens
}

func stateIdentifierRune(r rune) bool {
	switch {
	case r >= '0' && r <= '9':
		return true
	case unicode.IsLetter(r):
		return true
	case unicode.IsSpace(r):
		return true
	case r == '.' || r == '-' || r == '/' || r == ':':
		return true
	default:
		return false
	}
}

func isIdentifierLabel(value string) bool {
	switch value {
	case "OAB", "CREA", "COREN", "CRM", "CRO", "RG", "CNH", "CPF", "CNPJ", "CTPS", "SUS",
		"CNS", "TITULO", "PASSAPORTE", "IDENTIDADE", "CARTEIRINHA", "CARTEIRA",
		"ESTUDANTIL", "ELEITOR", "CERTIDAO", "NASCIMENTO", "CASAMENTO", "CARTAO",
		"CIDADAO", "DOCUMENTO", "NUMERO", "NUM", "NRO", "NR", "N", "DE", "DA",
		"DO", "DOS", "DAS", "E", "EM", "NO", "NA", "NOS", "NAS":
		return true
	default:
		return false
	}
}

func splitStateCanonical(value string) (number, series, state string) {
	rest := value
	if len(value) >= 3 && value[len(value)-3] == '/' {
		state = value[len(value)-2:]
		rest = value[:len(value)-3]
	}
	if index := strings.LastIndexByte(rest, '-'); index >= 0 {
		return rest[:index], rest[index+1:], state
	}
	return rest, "", state
}

func digitCount(value string) int {
	count := 0
	for i := 0; i < len(value); i++ {
		if value[i] >= '0' && value[i] <= '9' {
			count++
		}
	}
	return count
}
