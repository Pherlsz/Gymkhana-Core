package normalize

import (
	"strings"
	"unicode"
)

var personNameJunk = map[string]bool{
	"nome": true, "mae": true, "pai": true, "nome completo": true,
	"nome da mae": true, "nome do pai": true, "fulano": true,
	"nao informado": true, "desconhecido": true, "desconhecida": true,
	"xxx": true, "teste": true, "n a": true, "na": true, "sem nome": true,
}

var personParticles = map[string]bool{
	"de": true, "da": true, "do": true, "das": true, "dos": true,
	"e": true, "di": true, "du": true, "del": true,
}

// FormatPersonName returns the persisted person name. Placeholder dashes,
// digits, headers, spaced acronyms, and legal-entity names become empty.
func FormatPersonName(value string) string {
	value = DisplayText(value)
	value = strings.Trim(value, " \t.-_,;:[]{}()/\\'\"")
	if value == "" {
		return ""
	}
	if isPunctuationOrDigits(value) {
		return ""
	}
	key := SearchText(value)
	if key == "" || isDigitsOnly(key) || personNameJunk[key] || isAbsent(value) {
		return ""
	}
	tokens := strings.Fields(key)
	if len(tokens) >= 2 {
		allSingle := true
		for _, token := range tokens {
			if len([]rune(token)) != 1 {
				allSingle = false
				break
			}
		}
		if allSingle {
			return ""
		}
	}
	if !hasTwoLetters(key) {
		return ""
	}
	if isCompanyName(value) {
		return ""
	}
	return titlePerson(value)
}

var companyWords = map[string]bool{
	"ltda": true, "limitada": true, "eireli": true, "epp": true, "mei": true,
	"cnpj": true, "companhia": true, "cia": true, "associacao": true,
	"cooperativa": true, "sindicato": true, "fundacao": true, "instituto": true,
	"prefeitura": true, "empreendimentos": true, "holding": true,
	"construtora": true, "transportadora": true, "distribuidora": true,
	"imobiliaria": true, "metalurgica": true, "sociedade": true, "empresa": true,
	"comercio": true, "comercios": true, "industria": true, "industrias": true,
	"microempresa": true, "laticinios": true, "madeireira": true,
	"agropecuaria": true, "pecuaria": true, "mineracao": true, "mineradora": true,
	"carbonifera": true, "ceramica": true, "engenharia": true, "consultoria": true,
	"contabilidade": true, "advocacia": true, "informatica": true,
	"importacao": true, "exportacao": true, "corretora": true, "varejista": true,
	"associados": true, "delegacia": true, "autarquia": true, "cartorio": true,
	"clinica": true, "laboratorio": true, "hospital": true,
	"igreja": true, "paroquia": true, "diocese": true, "clube": true,
	"escola": true, "escolas": true, "colegio": true, "universidade": true,
	"faculdade": true, "padaria": true, "farmacia": true, "drogaria": true,
	"supermercado": true, "mercado": true, "loja": true, "restaurante": true,
	"oficina": true, "grafica": true, "funeraria": true, "borracharia": true,
	"vidracaria": true, "autopecas": true, "acougue": true, "lancheria": true,
	"hotel": true, "motel": true, "pousada": true, "comercial": true,
	"industrial": true, "servico": true, "servicos": true, "posto": true,
	"agencia": true, "templo": true, "capela": true, "orgao": true,
	"churrascaria": true, "pizzaria": true, "barbearia": true,
	"conveniencia": true, "otica": true, "optica": true, "joalheria": true,
	"confeccao": true, "calcados": true, "funilaria": true, "marcenaria": true,
	"serralheria": true, "concessionaria": true, "locadora": true,
	"administradora": true, "incorporadora": true, "seguradora": true,
	"beneficiadora": true, "atacadista": true, "recapadora": true,
	"serraria": true, "tecelagem": true, "malharia": true, "pastelaria": true,
	"sorveteria": true, "panificadora": true, "lavanderia": true,
}

var companySuffix = map[string]bool{
	"ltda": true, "limitada": true, "eireli": true, "epp": true, "mei": true,
}

func isCompanyName(value string) bool {
	key := SearchText(value)
	tokens := strings.Fields(key)
	if len(tokens) == 0 {
		return false
	}
	compact := strings.ReplaceAll(key, " ", "")
	if compact == "cnpj" || (len(compact) == 14 && isDigitsOnly(compact)) {
		return true
	}
	for _, token := range tokens {
		if companyWords[token] {
			return true
		}
	}
	last := tokens[len(tokens)-1]
	if companySuffix[last] {
		return true
	}
	if len(tokens) >= 3 && last == "me" {
		return true
	}
	if len(tokens) >= 2 && tokens[len(tokens)-2] == "s" && last == "a" {
		return true
	}
	if hasToken(tokens, "filhos") && (hasToken(tokens, "e") || strings.Contains(value, "&")) {
		return true
	}
	return false
}

func hasToken(tokens []string, want string) bool {
	for _, token := range tokens {
		if token == want {
			return true
		}
	}
	return false
}

func titlePerson(value string) string {
	fields := strings.Fields(DisplayText(value))
	out := make([]string, 0, len(fields))
	for i, field := range fields {
		key := SearchText(field)
		if i > 0 && personParticles[key] {
			out = append(out, key)
			continue
		}
		parts := strings.Split(field, "-")
		titled := make([]string, 0, len(parts))
		for _, part := range parts {
			if part == "" {
				continue
			}
			titled = append(titled, titleWord(part))
		}
		out = append(out, strings.Join(titled, "-"))
	}
	return strings.Join(out, " ")
}

func isPunctuationOrDigits(value string) bool {
	for _, r := range value {
		if unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func isDigitsOnly(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			if r == ' ' {
				continue
			}
			return false
		}
	}
	return strings.ContainsAny(value, "0123456789")
}

func hasTwoLetters(value string) bool {
	letters := 0
	for _, r := range value {
		if r >= 'a' && r <= 'z' {
			letters++
			if letters >= 2 {
				return true
			}
		}
	}
	return false
}
