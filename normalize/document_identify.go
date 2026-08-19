package normalize

import "strings"

// DocumentMatch is a unique catalog kind plus its canonical identifier.
type DocumentMatch struct {
	Kind      DocumentKind
	Canonical string
}

// identifyingDocumentKinds are the only types that may be inferred from an
// unlabeled number. Weak shapes (RG without X, CTPS, professional licenses,
// student cards, certificates) are excluded because they collide with almost
// any digit string. PIS is also excluded: it is 11 digits with a different
// check digit than CPF/CNH, and guessing it unlabeled would create false
// "RG is a CPF" / "CPF is a PIS" collisions.
var identifyingDocumentKinds = []DocumentKind{
	DocumentCPF,
	DocumentCNPJ,
	DocumentCNH,
	DocumentVoterID,
	DocumentSUSCard,
	DocumentPassport,
	DocumentRG,
}

var documentTypeLabels = map[string]DocumentKind{
	"CPF":         DocumentCPF,
	"CNPJ":        DocumentCNPJ,
	"CNH":         DocumentCNH,
	"RG":          DocumentRG,
	"IDENTIDADE":  DocumentRG,
	"TITULO":      DocumentVoterID,
	"ELEITOR":     DocumentVoterID,
	"SUS":         DocumentSUSCard,
	"CNS":         DocumentSUSCard,
	"PASSAPORTE":  DocumentPassport,
	"CTPS":        DocumentCTPS,
	"PIS":         DocumentPIS,
	"PASEP":       DocumentPIS,
	"NIT":         DocumentPIS,
	"NIS":         DocumentPIS,
	"OAB":         DocumentOAB,
	"CREA":        DocumentCREA,
	"COREN":       DocumentCOREN,
	"CRM":         DocumentCRM,
	"CRO":         DocumentCRO,
	"CIDADAO":     DocumentCitizenCard,
	"ESTUDANTIL":  DocumentStudentID,
	"CARTEIRINHA": DocumentStudentID,
	"NASCIMENTO":  DocumentBirthCertificate,
	"CASAMENTO":   DocumentMarriageCertificate,
}

// IdentifyDocument infers the catalog kind of an identifier when the caller
// does not already know the type. It succeeds only when exactly one kind
// matches. Unlabeled values use check digits or a rigid pattern; labeled
// values are pinned to that label and never reassigned to another type.
func IdentifyDocument(value string) (DocumentMatch, error) {
	matches, err := identifyDocuments(value)
	if err != nil {
		return DocumentMatch{}, err
	}
	switch len(matches) {
	case 0:
		return DocumentMatch{}, validationError(KindDocument, CodeUnknownKind)
	case 1:
		return matches[0], nil
	default:
		return DocumentMatch{}, validationError(KindDocument, CodeAmbiguous)
	}
}

// IdentifyDocumentMatches returns every strong (or label-pinned) match for
// value, in a stable order. Callers that need a review queue should use this
// when IdentifyDocument returns CodeAmbiguous.
func IdentifyDocumentMatches(value string) []DocumentMatch {
	matches, err := identifyDocuments(value)
	if err != nil {
		return nil
	}
	return matches
}

func identifyDocuments(value string) ([]DocumentMatch, error) {
	if DisplayText(value) == "" {
		return nil, validationError(KindDocument, CodeEmpty)
	}

	labels := labeledDocumentKinds(value)
	if len(labels) > 0 {
		kind, ok := uniqueLabeledKind(labels)
		if !ok {
			return nil, validationError(KindDocument, CodeAmbiguous)
		}
		canonical, err := CanonicalDocument(kind, identifierWithoutTypeLabels(value))
		if err != nil {
			return nil, err
		}
		return []DocumentMatch{{Kind: kind, Canonical: canonical}}, nil
	}

	var matches []DocumentMatch
	for _, kind := range identifyingDocumentKinds {
		canonical, err := CanonicalDocument(kind, value)
		if err != nil {
			continue
		}
		if kind == DocumentRG && !rgHasCheckLetter(canonical) {
			continue
		}
		matches = append(matches, DocumentMatch{Kind: kind, Canonical: canonical})
	}
	return matches, nil
}

func labeledDocumentKinds(value string) []DocumentKind {
	seen := map[DocumentKind]struct{}{}
	var kinds []DocumentKind
	for _, token := range scanIdentifierTokens(strings.ToUpper(SearchText(value))) {
		if token.kind != tokenLetters {
			continue
		}
		kind, ok := documentTypeLabels[token.text]
		if !ok {
			continue
		}
		if _, exists := seen[kind]; exists {
			continue
		}
		seen[kind] = struct{}{}
		kinds = append(kinds, kind)
	}
	return kinds
}

func uniqueLabeledKind(labels []DocumentKind) (DocumentKind, bool) {
	if len(labels) == 1 {
		return labels[0], true
	}
	if len(labels) == 2 && professionalPair(labels[0], labels[1]) {
		return DocumentCREAOAB, true
	}
	return "", false
}

func professionalPair(left, right DocumentKind) bool {
	return (left == DocumentOAB && right == DocumentCREA) || (left == DocumentCREA && right == DocumentOAB)
}

func identifierWithoutTypeLabels(value string) string {
	var parts []string
	for _, token := range scanIdentifierTokens(strings.ToUpper(SearchText(value))) {
		if token.kind == tokenLetters {
			if _, labeled := documentTypeLabels[token.text]; labeled {
				continue
			}
			if isIdentifierLabel(token.text) {
				continue
			}
		}
		parts = append(parts, token.text)
	}
	return strings.Join(parts, " ")
}

func rgHasCheckLetter(canonical string) bool {
	number, _, _ := splitStateCanonical(canonical)
	return strings.HasSuffix(number, "X")
}
