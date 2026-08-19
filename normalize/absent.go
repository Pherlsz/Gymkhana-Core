package normalize

import "strings"

var absentValues = map[string]bool{
	"nao": true, "n": true, "nao tenho": true, "nao possuo": true, "nao tem": true,
	"nenhum": true, "nenhuma": true, "nao possui": true, "nops": true, "no": true,
	"nunca": true, "sem": true, "nao sei": true, "0": true,
}

func isAbsent(value string) bool {
	key := SearchText(value)
	if key == "" || absentValues[key] {
		return true
	}
	return strings.HasPrefix(key, "nao ") || strings.HasPrefix(key, "nenhum")
}
