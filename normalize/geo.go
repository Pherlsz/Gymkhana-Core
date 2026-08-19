package normalize

import (
	"slices"
	"strings"
)

var cityHeaderJunk = map[string]bool{
	"cidade": true, "cidade de nascimento": true, "cidade que reside": true,
	"cidade que nasceu": true, "municipio": true, "bacanal": true,
}

var cityAliases = map[string]string{
	"poa": "Porto Alegre", "porto alegre": "Porto Alegre", "porto alegre rs": "Porto Alegre",
	"sao jeronimo": "São Jerônimo", "sao jeronimo rs": "São Jerônimo", "s jeronimo": "São Jerônimo",
	"sao jeo": "São Jerônimo", "sj": "São Jerônimo",
	"butia": "Butiá", "butia rs": "Butiá",
	"minas do leao": "Minas do Leão", "leao": "Minas do Leão", "minas leao": "Minas do Leão",
	"arroio dos ratos": "Arroio dos Ratos", "ratos": "Arroio dos Ratos",
	"charqueadas": "Charqueadas", "charqueadas rs": "Charqueadas",
	"chaqueadas": "Charqueadas", "charqueads": "Charqueadas", "charquedas": "Charqueadas",
	"sao leopoldo": "São Leopoldo", "sao leopoldo rs": "São Leopoldo", "sl": "São Leopoldo",
	"novo hamburgo": "Novo Hamburgo", "nh": "Novo Hamburgo",
	"sapucaia do sul": "Sapucaia do Sul", "sapucaia": "Sapucaia do Sul",
	"canoas": "Canoas", "esteio": "Esteio", "gravatai": "Gravataí",
	"montenegro": "Montenegro", "cachoeirinha": "Cachoeirinha",
	"sapiranga": "Sapiranga", "campo bom": "Campo Bom", "portao": "Portão",
	"dois irmaos": "Dois Irmãos", "estancia velha": "Estância Velha",
	"ivoti": "Ivoti", "carlos barbosa": "Carlos Barbosa",
	"sao sebastiao do cai": "São Sebastião do Caí", "sao sebastiao cai": "São Sebastião do Caí",
	"bento goncalves": "Bento Gonçalves", "garibaldi": "Garibaldi",
	"guaiba": "Guaíba", "osorio": "Osório", "gramado": "Gramado",
	"alvorada": "Alvorada", "taquara": "Taquara", "farroupilha": "Farroupilha",
	"bom principio": "Bom Princípio", "feliz": "Feliz",
	"santo antonio da patrulha": "Santo Antônio da Patrulha",
	"viamao":                    "Viamão", "tramandai": "Tramandaí", "triunfo": "Triunfo",
	"igrejinha": "Igrejinha", "vacaria": "Vacaria", "rio pardo": "Rio Pardo",
	"passo fundo": "Passo Fundo", "parobe": "Parobé", "santa maria": "Santa Maria",
	"sao francisco de paula": "São Francisco de Paula",
	"caxias do sul":          "Caxias do Sul", "caxias": "Caxias do Sul",
	"camaqua": "Camaquã", "seberi": "Seberi", "rolante": "Rolante",
	"general camara": "General Câmara", "tres coroas": "Três Coroas",
	"cachoeira do sul": "Cachoeira do Sul", "pelotas": "Pelotas",
	"rio grande": "Rio Grande", "santa cruz do sul": "Santa Cruz do Sul",
	"lajeado": "Lajeado", "arroio do sal": "Arroio do Sal",
	"capao da canoa": "Capão da Canoa", "eldorado do sul": "Eldorado do Sul",
	"sao jeronino": "São Jerônimo", "sao geronimo": "São Jerônimo",
}

var neighborhoodAliases = map[string]string{
	"bella vista": "Bela Vista", "bela vista": "Bela Vista",
	"sao thomas": "São Thomas", "sao tomas": "São Tomás",
	"sao jose": "São José", "sao francisco": "São Francisco",
	"sao miguel":              "São Miguel",
	"nossa senhora aparecida": "Nossa Senhora Aparecida",
	"passo d areia":           "Passo d'Areia", "passo da areia": "Passo d'Areia",
	"cidade alta": "Cidade Alta", "cidade baixa": "Cidade Baixa",
	"sta teresa": "Santa Teresa", "santa tereza": "Santa Teresa",
	"santa teresa": "Santa Teresa",
	"centro":       "Centro", "navegantes": "Navegantes",
}

var countryByCode = map[string]string{
	"bra": "Brasil", "br": "Brasil", "ury": "Uruguai", "uy": "Uruguai",
	"arg": "Argentina", "ar": "Argentina", "chl": "Chile", "cl": "Chile",
	"pry": "Paraguai", "py": "Paraguai", "usa": "Estados Unidos", "us": "Estados Unidos",
	"deu": "Alemanha", "de": "Alemanha", "fra": "França", "fr": "França",
	"prt": "Portugal", "pt": "Portugal", "per": "Peru", "pe": "Peru",
	"jpn": "Japão", "jp": "Japão", "ven": "Venezuela", "ve": "Venezuela",
	"ago": "Angola", "ao": "Angola", "ita": "Itália", "it": "Itália",
	"esp": "Espanha", "es": "Espanha", "mex": "México", "mx": "México",
	"can": "Canadá", "ca": "Canadá",
}

var countryByName = map[string]string{
	"brasil": "Brasil", "brazil": "Brasil",
	"uruguai": "Uruguai", "uruguay": "Uruguai",
	"argentina": "Argentina", "chile": "Chile",
	"paraguai": "Paraguai", "paraguay": "Paraguai",
	"estados unidos": "Estados Unidos", "eua": "Estados Unidos",
	"alemanha": "Alemanha", "franca": "França", "portugal": "Portugal",
	"peru": "Peru", "japao": "Japão", "venezuela": "Venezuela",
	"angola": "Angola", "italia": "Itália", "espanha": "Espanha",
	"mexico": "México", "canada": "Canadá",
}

var cityCanonFold map[string]string
var cityCompact map[string]string

func init() {
	cityCanonFold = map[string]string{}
	cityCompact = map[string]string{}
	for key, label := range cityAliases {
		cityCanonFold[SearchText(label)] = label
		cityCompact[strings.ReplaceAll(key, " ", "")] = label
	}
}

// FormatCity returns a catalog city, a one-edit match, or a titled name.
func FormatCity(value string) string {
	if isAbsent(value) {
		return ""
	}
	key := SearchText(value)
	key = strings.TrimSpace(strings.TrimSuffix(key, " rs"))
	key = strings.TrimSpace(strings.TrimSuffix(key, " sc"))
	key = strings.TrimSpace(strings.TrimSuffix(key, " pr"))
	if key == "" || cityHeaderJunk[key] || containsDigit(key) {
		return ""
	}
	if label, ok := cityAliases[key]; ok {
		return label
	}
	compact := strings.ReplaceAll(key, " ", "")
	if label, ok := cityCompact[compact]; ok {
		return label
	}
	if len(compact) >= 6 {
		best, bestD := "", 3
		for alias, label := range cityCompact {
			d := levenshtein(compact, alias)
			if d < bestD {
				best, bestD = label, d
			}
		}
		if best != "" && (bestD == 1 || (bestD == 2 && len(compact) >= 10)) {
			return best
		}
	}
	return titleStreetName(value)
}

// FormatMunicipality returns FormatCity only when the result is a known city.
func FormatMunicipality(value string) string {
	city := FormatCity(value)
	if city == "" {
		return ""
	}
	if _, ok := cityCanonFold[SearchText(city)]; ok {
		return city
	}
	return ""
}

// FormatNeighborhood canonicalizes known bairros; unknown names are titled.
func FormatNeighborhood(value string) string {
	if isAbsent(value) {
		return ""
	}
	key := SearchText(value)
	if key == "" || isDigitsOnly(key) {
		return ""
	}
	if label, ok := neighborhoodAliases[key]; ok {
		return label
	}
	return titleStreetName(value)
}

// FormatCountry returns a single country from ISO code or name.
func FormatCountry(value string) string {
	if isAbsent(value) {
		return ""
	}
	key := SearchText(value)
	if label, ok := countryByCode[key]; ok {
		return label
	}
	return countryByName[key]
}

// FormatNationality returns one or more countries, semicolon-separated.
func FormatNationality(value string) string {
	one := FormatCountry(value)
	if one != "" {
		return one
	}
	if isAbsent(value) {
		return ""
	}
	key := " " + SearchText(value) + " "
	var found []string
	seen := map[string]bool{}
	for name, label := range countryByName {
		if strings.Contains(key, " "+name+" ") && !seen[label] {
			found = append(found, label)
			seen[label] = true
		}
	}
	if len(found) == 0 {
		return ""
	}
	sortStrings(found)
	return strings.Join(found, "; ")
}

func containsDigit(value string) bool {
	for _, r := range value {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

func levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	if abs(len(a)-len(b)) > 2 {
		return 99
	}
	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 0; i < len(a); i++ {
		cur := make([]int, len(b)+1)
		cur[0] = i + 1
		for j := 0; j < len(b); j++ {
			ins := cur[j] + 1
			del := prev[j+1] + 1
			sub := prev[j]
			if a[i] != b[j] {
				sub++
			}
			cur[j+1] = min3(ins, del, sub)
		}
		prev = cur
	}
	return prev[len(b)]
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func min3(a, b, c int) int {
	if a <= b && a <= c {
		return a
	}
	if b <= c {
		return b
	}
	return c
}

func sortStrings(values []string) {
	slices.Sort(values)
}
