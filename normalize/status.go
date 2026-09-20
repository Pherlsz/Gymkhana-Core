package normalize

import "strings"

var genderValues = map[string]string{
	"m": "M", "masculino": "M", "masc": "M",
	"f": "F", "feminino": "F", "fem": "F",
	"outro": "Outro", "outros": "Outro",
}

var maritalValues = map[string]string{
	"solteiro": "solteiro", "solteira": "solteiro",
	"casado": "casado", "casada": "casado",
	"separado": "separado", "separada": "separado",
	"divorciado": "divorciado", "divorciada": "divorciado",
	"viuvo": "viuvo", "viuva": "viuvo",
}

var maritalEmpty = map[string]bool{
	"namorando": true, "uniao estavel": true,
	"relacionamento serio": true, "relacionamento": true,
}

var bloodValues = map[string]string{
	"a+": "A+", "a-": "A-", "b+": "B+", "b-": "B-",
	"ab+": "AB+", "ab-": "AB-", "o+": "O+", "o-": "O-",
	"a": "A", "b": "B", "ab": "AB", "o": "O", "0": "O",
	"a positivo": "A+", "a negativo": "A-",
	"o positivo": "O+", "o negativo": "O-",
	"b positivo": "B+", "b negativo": "B-",
	"ab positivo": "AB+", "ab negativo": "AB-",
}

// FormatGender returns M, F, or Outro. Empty and refusals stay empty.
func FormatGender(value string) string {
	if isAbsent(value) {
		return ""
	}
	return genderValues[SearchText(value)]
}

// FormatMaritalStatus returns the CC 1.571 status. União estável and dating
// are not stored as casado.
func FormatMaritalStatus(value string) string {
	if isAbsent(value) {
		return ""
	}
	key := SearchText(value)
	if maritalEmpty[key] {
		return ""
	}
	return maritalValues[key]
}

// FormatBloodType returns A+/O- and the other ABO labels.
func FormatBloodType(value string) string {
	if isAbsent(value) {
		return ""
	}
	key := SearchText(value)
	key = strings.ReplaceAll(key, "positivo", "+")
	key = strings.ReplaceAll(key, "negativo", "-")
	key = strings.ReplaceAll(key, " ", "")
	if label, ok := bloodValues[key]; ok {
		return label
	}
	return bloodValues[SearchText(value)]
}
