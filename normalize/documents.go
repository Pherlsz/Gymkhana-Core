package normalize

import (
	"strings"
	"unicode"
)

// DocumentKind is a stable technical key for a catalog document type.
type DocumentKind string

const (
	DocumentCPF                 DocumentKind = "cpf"
	DocumentCNPJ                DocumentKind = "cnpj"
	DocumentRG                  DocumentKind = "rg"
	DocumentIdentidade          DocumentKind = "identidade"
	DocumentVoterID             DocumentKind = "voter_id"
	DocumentCNH                 DocumentKind = "cnh"
	DocumentCTPS                DocumentKind = "ctps"
	DocumentPIS                 DocumentKind = "pis"
	DocumentPassport            DocumentKind = "passport"
	DocumentStudentID           DocumentKind = "student_id"
	DocumentCitizenCard         DocumentKind = "citizen_card"
	DocumentSUSCard             DocumentKind = "sus_card"
	DocumentOAB                 DocumentKind = "oab"
	DocumentCREA                DocumentKind = "crea"
	DocumentCOREN               DocumentKind = "coren"
	DocumentCRM                 DocumentKind = "crm"
	DocumentCRO                 DocumentKind = "cro"
	DocumentCREAOAB             DocumentKind = "crea_oab"
	DocumentBirthCertificate    DocumentKind = "birth_certificate"
	DocumentMarriageCertificate DocumentKind = "marriage_certificate"
)

type documentSpec struct {
	canonical func(kind DocumentKind, value string) (string, error)
	format    func(canonical string) string
}

var documentSpecs = map[DocumentKind]documentSpec{
	DocumentCPF:                 {canonical: canonicalNationalDigits, format: formatCPF},
	DocumentCNPJ:                {canonical: canonicalCNPJDocument, format: formatCNPJ},
	DocumentCNH:                 {canonical: canonicalNationalDigits, format: formatCNH},
	DocumentPIS:                 {canonical: canonicalNationalDigits, format: formatPIS},
	DocumentVoterID:             {canonical: canonicalNationalDigits, format: formatVoterID},
	DocumentSUSCard:             {canonical: canonicalNationalDigits, format: formatSUSCard},
	DocumentPassport:            {canonical: canonicalPassport, format: formatPassport},
	DocumentRG:                  {canonical: canonicalStateDocument, format: formatRG},
	DocumentIdentidade:          {canonical: canonicalStateDocument, format: formatRG},
	DocumentCTPS:                {canonical: canonicalStateDocument, format: formatCTPS},
	DocumentOAB:                 {canonical: canonicalStateDocument, format: formatProfessional},
	DocumentCREA:                {canonical: canonicalStateDocument, format: formatProfessional},
	DocumentCOREN:               {canonical: canonicalStateDocument, format: formatProfessional},
	DocumentCRM:                 {canonical: canonicalStateDocument, format: formatProfessional},
	DocumentCRO:                 {canonical: canonicalStateDocument, format: formatProfessional},
	DocumentCREAOAB:             {canonical: canonicalStateDocument, format: formatProfessional},
	DocumentStudentID:           {canonical: canonicalLooseDocument, format: formatLoose},
	DocumentCitizenCard:         {canonical: canonicalLooseDocument, format: formatLoose},
	DocumentBirthCertificate:    {canonical: canonicalLooseDocument, format: formatLoose},
	DocumentMarriageCertificate: {canonical: canonicalLooseDocument, format: formatLoose},
}

// DocumentKinds returns the catalog kinds Core can canonicalize, in a stable order.
func DocumentKinds() []DocumentKind {
	return []DocumentKind{
		DocumentCPF,
		DocumentCNPJ,
		DocumentRG,
		DocumentIdentidade,
		DocumentVoterID,
		DocumentCNH,
		DocumentCTPS,
		DocumentPIS,
		DocumentPassport,
		DocumentStudentID,
		DocumentCitizenCard,
		DocumentSUSCard,
		DocumentOAB,
		DocumentCREA,
		DocumentCOREN,
		DocumentCRM,
		DocumentCRO,
		DocumentCREAOAB,
		DocumentBirthCertificate,
		DocumentMarriageCertificate,
	}
}

// KnownDocumentKind reports whether Core has a validator and formatter for kind.
func KnownDocumentKind(kind DocumentKind) bool {
	_, ok := documentSpecs[kind]
	return ok
}

// CanonicalDocument validates value for kind and returns the persisted identifier.
//
// National numbers (CPF, CNH, título, CNS) are digits only, with leading zeros
// kept. Professional licenses and other state-issued numbers fold the UF into
// the identifier as NUMBER/UF so callers do not need a second column for state.
func CanonicalDocument(kind DocumentKind, value string) (string, error) {
	spec, ok := documentSpecs[kind]
	if !ok {
		return "", validationError(ValueKind(kind), CodeUnknownKind)
	}
	return spec.canonical(kind, value)
}

// FormatDocument canonicalizes value for kind and returns the display form.
func FormatDocument(kind DocumentKind, value string) (string, error) {
	spec, ok := documentSpecs[kind]
	if !ok {
		return "", validationError(ValueKind(kind), CodeUnknownKind)
	}
	canonical, err := spec.canonical(kind, value)
	if err != nil {
		return "", err
	}
	return spec.format(canonical), nil
}

func canonicalNationalDigits(kind DocumentKind, value string) (string, error) {
	if kind == DocumentCPF {
		return CanonicalCPF(value)
	}
	if !containsOnly(value, digitSeparator) {
		return "", validationError(ValueKind(kind), CodeInvalidFormat)
	}
	digits := Digits(value)
	if digits == "" {
		return "", validationError(ValueKind(kind), CodeEmpty)
	}

	switch kind {
	case DocumentCNH:
		if len(digits) != 11 {
			return "", validationError(ValueKind(kind), CodeInvalidLength)
		}
		if !validCNH(digits) {
			return "", validationError(ValueKind(kind), CodeInvalidChecksum)
		}
	case DocumentPIS:
		if len(digits) != 11 {
			return "", validationError(ValueKind(kind), CodeInvalidLength)
		}
		if !validPIS(digits) {
			return "", validationError(ValueKind(kind), CodeInvalidChecksum)
		}
	case DocumentVoterID:
		if len(digits) != 12 {
			return "", validationError(ValueKind(kind), CodeInvalidLength)
		}
		if !validVoterID(digits) {
			return "", validationError(ValueKind(kind), CodeInvalidChecksum)
		}
	case DocumentSUSCard:
		if len(digits) != 15 {
			return "", validationError(ValueKind(kind), CodeInvalidLength)
		}
		if !validCNS(digits) {
			return "", validationError(ValueKind(kind), CodeInvalidChecksum)
		}
	default:
		return "", validationError(ValueKind(kind), CodeUnknownKind)
	}
	return digits, nil
}

func digitSeparator(r rune) bool {
	return (r >= '0' && r <= '9') || r == '.' || r == '-' || unicode.IsSpace(r)
}

func canonicalCNPJDocument(_ DocumentKind, value string) (string, error) {
	return CanonicalCNPJ(value)
}

func formatCPF(canonical string) string {
	return canonical[:3] + "." + canonical[3:6] + "." + canonical[6:9] + "-" + canonical[9:]
}

func formatCNPJ(canonical string) string {
	return canonical[:2] + "." + canonical[2:5] + "." + canonical[5:8] + "/" + canonical[8:12] + "-" + canonical[12:]
}

func formatCNH(canonical string) string {
	return canonical[:3] + " " + canonical[3:6] + " " + canonical[6:9] + " " + canonical[9:]
}

func formatPIS(canonical string) string {
	return canonical[:3] + "." + canonical[3:8] + "." + canonical[8:10] + "-" + canonical[10:]
}

func formatVoterID(canonical string) string {
	return canonical[:4] + " " + canonical[4:8] + " " + canonical[8:]
}

func formatSUSCard(canonical string) string {
	return canonical[:3] + " " + canonical[3:7] + " " + canonical[7:11] + " " + canonical[11:]
}

func formatPassport(canonical string) string {
	return canonical
}

func formatRG(canonical string) string {
	return formatStateCanonical(canonical, true)
}

func formatCTPS(canonical string) string {
	return formatStateCanonical(canonical, false)
}

func formatProfessional(canonical string) string {
	return formatStateCanonical(canonical, false)
}

func formatLoose(canonical string) string {
	return canonical
}

func formatStateCanonical(canonical string, hyphenLast bool) string {
	number, series, state := splitStateCanonical(canonical)
	body := number
	suffix := ""
	if strings.HasSuffix(number, "X") {
		body = number[:len(number)-1]
		suffix = "-X"
	} else if hyphenLast && len(number) >= 2 {
		body = number[:len(number)-1]
		suffix = "-" + number[len(number)-1:]
	}
	formatted := groupThousands(body) + suffix
	if series != "" {
		formatted += "-" + series
	}
	if state != "" {
		formatted += "/" + state
	}
	return formatted
}

func groupThousands(digits string) string {
	if len(digits) <= 3 {
		return digits
	}
	var parts []string
	for len(digits) > 3 {
		parts = append(parts, digits[len(digits)-3:])
		digits = digits[:len(digits)-3]
	}
	parts = append(parts, digits)
	var builder strings.Builder
	for i := len(parts) - 1; i >= 0; i-- {
		if builder.Len() > 0 {
			builder.WriteByte('.')
		}
		builder.WriteString(parts[i])
	}
	return builder.String()
}
