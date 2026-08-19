package normalize

var brazilStates = map[string]struct{}{
	"AC": {}, "AL": {}, "AP": {}, "AM": {}, "BA": {}, "CE": {}, "DF": {}, "ES": {}, "GO": {},
	"MA": {}, "MT": {}, "MS": {}, "MG": {}, "PA": {}, "PB": {}, "PR": {}, "PE": {}, "PI": {},
	"RJ": {}, "RN": {}, "RS": {}, "RO": {}, "RR": {}, "SC": {}, "SP": {}, "SE": {}, "TO": {},
}

func validBrazilState(value string) bool {
	_, ok := brazilStates[value]
	return ok
}
