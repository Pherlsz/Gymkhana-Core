package normalize

import (
	"strings"
	"unicode"
)

var streetTypeLabels = map[string]string{
	"r":        "Rua",
	"rua":      "Rua",
	"av":       "Avenida",
	"avenida":  "Avenida",
	"tv":       "Travessa",
	"travessa": "Travessa",
	"al":       "Alameda",
	"alameda":  "Alameda",
	"praca":    "Praça",
	"estrada":  "Estrada",
	"rodovia":  "Rodovia",
	"beco":     "Beco",
	"largo":    "Largo",
	"viela":    "Viela",
	"servidao": "Servidão",
	"acesso":   "Acesso",
	"rs":       "RS",
	"br":       "BR",
}

var streetParticles = map[string]bool{
	"de": true, "da": true, "do": true, "das": true, "dos": true, "e": true,
}

var houseNumberEmpty = map[string]bool{
	"sn": true, "s n": true, "sem numero": true, "sem n": true,
}

var houseNumberLabels = map[string]bool{
	"n": true, "no": true, "num": true, "numero": true,
}

var blockLabels = []string{"bloco", "blc", "bl"}

var apartmentLabels = []string{"apartamento", "apart", "apto", "apt", "ap"}

// FormatStreet returns the persisted logradouro: canonical type prefix plus
// titled name. "r primavera" and "Rua Primavera" both become "Rua Primavera".
// Addresses without a recognized type keep the titled name and do not gain "Rua".
func FormatStreet(value string) string {
	value = DisplayText(value)
	if value == "" {
		return ""
	}

	fields := strings.Fields(value)
	label, fields := peelLabels(fields, streetTypeKey)
	name := titleStreetName(strings.Join(fields, " "))
	if label == "" {
		return name
	}
	if name == "" {
		return label
	}
	return label + " " + name
}

// FormatHouseNumber keeps letters and digits only (123A). Empty, s/n, and
// labels such as nº are dropped.
func FormatHouseNumber(value string) string {
	value = DisplayText(value)
	if value == "" {
		return ""
	}
	if houseNumberEmpty[SearchText(value)] {
		return ""
	}

	fields := strings.Fields(value)
	_, fields = peelLabels(fields, func(token string) (string, bool) {
		key := SearchText(strings.TrimRight(token, "º°."))
		if houseNumberLabels[key] {
			return "", true
		}
		return "", false
	})
	return Alphanumeric(strings.Join(fields, ""))
}

// FormatBlock returns "bl. A" from bloco/bl/blc variants, or empty.
func FormatBlock(value string) string {
	return formatPrefixedSlot(value, "bl. ", blockLabels)
}

// FormatApartment returns "apt. 202" from ap/apto/apt/apartamento variants, or empty.
func FormatApartment(value string) string {
	return formatPrefixedSlot(value, "apt. ", apartmentLabels)
}

func formatPrefixedSlot(value, prefix string, labels []string) string {
	value = DisplayText(value)
	if value == "" {
		return ""
	}

	// Preserve an already canonical prefix before label peeling. The payload may
	// itself equal a valid label (for example "bl. BL" or "apt. AP"), and
	// treating both tokens as labels would otherwise erase a valid canonical
	// value on the second pass.
	if len(value) >= len(prefix) && strings.EqualFold(value[:len(prefix)], prefix) {
		canonical := Alphanumeric(strings.TrimSpace(value[len(prefix):]))
		if canonical == "" {
			return ""
		}
		return prefix + canonical
	}

	labelSet := make(map[string]bool, len(labels))
	for _, label := range labels {
		labelSet[label] = true
	}

	fields := strings.Fields(value)
	_, fields = peelLabels(fields, func(token string) (string, bool) {
		key := SearchText(strings.TrimRight(token, "º°."))
		if labelSet[key] {
			return "", true
		}
		return "", false
	})
	canonical := Alphanumeric(strings.Join(fields, ""))
	if canonical == "" {
		return ""
	}
	return prefix + canonical
}

func streetTypeKey(token string) (string, bool) {
	key := SearchText(strings.TrimRight(token, "."))
	label, ok := streetTypeLabels[key]
	return label, ok
}

func peelLabels(fields []string, match func(string) (string, bool)) (string, []string) {
	var kept string
	for len(fields) > 0 {
		label, ok := match(fields[0])
		if !ok {
			break
		}
		if kept == "" {
			kept = label
		}
		fields = fields[1:]
	}
	return kept, fields
}

func titleStreetName(value string) string {
	value = DisplayText(value)
	if value == "" {
		return ""
	}

	fields := strings.Fields(value)
	out := make([]string, 0, len(fields))
	for i, field := range fields {
		key := SearchText(field)
		if i > 0 && streetParticles[key] {
			out = append(out, key)
			continue
		}
		out = append(out, titleWord(field))
	}
	return strings.Join(out, " ")
}

func titleWord(word string) string {
	runes := []rune(strings.ToLower(word))
	if len(runes) == 0 {
		return ""
	}
	runes[0] = unicode.ToTitle(runes[0])
	return string(runes)
}

var stateAliases = map[string]string{
	"rs": "RS", "rio grande do sul": "RS", "rio grande sul": "RS",
	"sc": "SC", "santa catarina": "SC",
	"pr": "PR", "parana": "PR",
	"sp": "SP", "sao paulo": "SP",
	"rj": "RJ", "rio de janeiro": "RJ",
	"mg": "MG", "minas gerais": "MG",
}

// CanonicalCEP returns the 8 persisted digits, keeping leading zeros.
func CanonicalCEP(value string) (string, error) {
	if isAbsent(value) || DisplayText(value) == "" {
		return "", validationError(KindPostalCode, CodeEmpty)
	}
	for _, r := range value {
		if r == '-' || unicode.IsSpace(r) || (r >= '0' && r <= '9') {
			continue
		}
		return "", validationError(KindPostalCode, CodeInvalidFormat)
	}
	digits := Digits(value)
	if len(digits) != 8 {
		return "", validationError(KindPostalCode, CodeInvalidLength)
	}
	return digits, nil
}

// FormatCEP returns 00000-000, or empty when the value is missing or invalid.
func FormatCEP(value string) string {
	canonical, err := CanonicalCEP(value)
	if err != nil {
		return ""
	}
	return canonical[:5] + "-" + canonical[5:]
}

// FormatUF returns the two-letter Brazilian state, or empty.
func FormatUF(value string) string {
	if isAbsent(value) {
		return ""
	}
	key := SearchText(value)
	if label, ok := stateAliases[key]; ok {
		return label
	}
	compact := strings.ToUpper(Alphanumeric(value))
	if validBrazilState(compact) {
		return compact
	}
	return ""
}
