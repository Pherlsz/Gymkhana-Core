package normalize

import (
	"slices"
	"strings"
)

var cityHeaderJunk = map[string]bool{
	"cidade": true, "cidade de nascimento": true, "cidade que reside": true,
	"cidade que nasceu": true, "municipio": true,
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

type cityCompactEntry struct {
	alias     string
	runes     []rune
	canonical string
	ambiguous bool
}

// CityCatalog is an immutable, concurrency-safe catalog of canonical city
// names and their accepted aliases. A nil *CityCatalog behaves as an empty
// catalog: FormatCity falls back to title-casing non-empty, non-junk inputs
// and FormatMunicipality always returns "".
type CityCatalog struct {
	aliases   map[string]string
	canonFold map[string]string
	compact   map[string]string
	entries   []cityCompactEntry
}

var defaultCityCatalog = NewCityCatalog(cityAliases)

// NewCityCatalog builds an immutable CityCatalog from an alias-to-canonical-label
// map. Alias keys are folded through SearchText (and spaces stripped for compact
// matching) deterministically so caller map iteration order never affects behavior.
// A nil or empty map produces a valid empty catalog.
func NewCityCatalog(aliases map[string]string) *CityCatalog {
	c := &CityCatalog{
		aliases:   make(map[string]string, len(aliases)),
		canonFold: make(map[string]string, len(aliases)),
		compact:   make(map[string]string, len(aliases)),
	}
	if len(aliases) == 0 {
		return c
	}

	rawKeys := make([]string, 0, len(aliases))
	for k := range aliases {
		rawKeys = append(rawKeys, k)
	}
	slices.Sort(rawKeys)

	type candidate struct {
		key       string
		compact   string
		canonical string
	}
	candidates := make([]candidate, 0, len(rawKeys))
	keyAmbiguous := make(map[string]bool)
	compactAmbiguous := make(map[string]bool)

	for _, rawKey := range rawKeys {
		label := DisplayText(aliases[rawKey])
		key := SearchText(rawKey)
		if key == "" || label == "" {
			continue
		}
		canonKey := SearchText(label)
		if canonKey != "" {
			if prev, exists := c.canonFold[canonKey]; !exists || label < prev {
				c.canonFold[canonKey] = label
			}
		}

		if prev, exists := c.aliases[key]; exists && prev != label {
			keyAmbiguous[key] = true
		} else {
			c.aliases[key] = label
		}

		comp := strings.ReplaceAll(key, " ", "")
		if comp != "" {
			if prev, exists := c.compact[comp]; exists && prev != label {
				compactAmbiguous[comp] = true
			} else {
				c.compact[comp] = label
			}
		}

		candidates = append(candidates, candidate{
			key:       key,
			compact:   comp,
			canonical: label,
		})
	}

	// Ensure canonical labels themselves are also resolvable by key and compact form.
	canonKeys := make([]string, 0, len(c.canonFold))
	for k := range c.canonFold {
		canonKeys = append(canonKeys, k)
	}
	slices.Sort(canonKeys)
	for _, canonKey := range canonKeys {
		label := c.canonFold[canonKey]
		if prev, exists := c.aliases[canonKey]; exists && prev != label {
			keyAmbiguous[canonKey] = true
		} else if !exists {
			c.aliases[canonKey] = label
		}
		comp := strings.ReplaceAll(canonKey, " ", "")
		if comp != "" {
			if prev, exists := c.compact[comp]; exists && prev != label {
				compactAmbiguous[comp] = true
			} else if !exists {
				c.compact[comp] = label
				candidates = append(candidates, candidate{
					key:       canonKey,
					compact:   comp,
					canonical: label,
				})
			}
		}
	}

	for k := range keyAmbiguous {
		delete(c.aliases, k)
	}
	for comp := range compactAmbiguous {
		delete(c.compact, comp)
	}

	byCompact := make(map[string]cityCompactEntry, len(candidates))
	for _, cand := range candidates {
		if cand.compact == "" {
			continue
		}
		if existing, ok := byCompact[cand.compact]; ok {
			if existing.canonical != cand.canonical {
				existing.ambiguous = true
				byCompact[cand.compact] = existing
			}
			continue
		}
		byCompact[cand.compact] = cityCompactEntry{
			alias:     cand.compact,
			runes:     []rune(cand.compact),
			canonical: cand.canonical,
			ambiguous: compactAmbiguous[cand.compact],
		}
	}

	entries := make([]cityCompactEntry, 0, len(byCompact))
	for _, entry := range byCompact {
		entries = append(entries, entry)
	}
	slices.SortFunc(entries, func(a, b cityCompactEntry) int {
		return strings.Compare(a.alias, b.alias)
	})
	c.entries = entries
	return c
}

// FormatCity returns a catalog city from c, a deterministic bounded-edit match
// (when unambiguous), or a titled name. If c is nil, it behaves as an empty
// catalog and falls back to titleStreetName for valid inputs.
func (c *CityCatalog) FormatCity(value string) string {
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
	if c != nil {
		if label, ok := c.aliases[key]; ok {
			return label
		}
		compact := strings.ReplaceAll(key, " ", "")
		if label, ok := c.compact[compact]; ok {
			return label
		}
		var compactStack [64]rune
		var compactRunes []rune
		if len(compact) <= len(compactStack) {
			n := 0
			for _, r := range compact {
				compactStack[n] = r
				n++
			}
			compactRunes = compactStack[:n]
		} else {
			compactRunes = []rune(compact)
		}
		runeLen := len(compactRunes)
		if runeLen >= 6 && len(c.entries) > 0 {
			maxDist := 1
			if runeLen >= 10 {
				maxDist = 2
			}
			best := ""
			bestD := maxDist + 1
			ambiguous := false
			for i := range c.entries {
				entry := &c.entries[i]
				d := boundedLevenshteinRunes(compactRunes, entry.runes, maxDist)
				if d < bestD {
					bestD = d
					best = entry.canonical
					ambiguous = entry.ambiguous
				} else if d == bestD && d <= maxDist {
					if entry.ambiguous || entry.canonical != best {
						ambiguous = true
					}
				}
			}
			if best != "" && !ambiguous && bestD <= maxDist {
				return best
			}
		}
	}
	return titleStreetName(value)
}

// FormatMunicipality returns c.FormatCity(value) only when the result is a
// canonical city present in c. If c is nil, FormatMunicipality returns "".
func (c *CityCatalog) FormatMunicipality(value string) string {
	if c == nil || len(c.canonFold) == 0 {
		return ""
	}
	city := c.FormatCity(value)
	if city == "" {
		return ""
	}
	if _, ok := c.canonFold[SearchText(city)]; ok {
		return city
	}
	return ""
}

// FormatCity returns a catalog city, a one-edit match, or a titled name.
func FormatCity(value string) string {
	return defaultCityCatalog.FormatCity(value)
}

// FormatMunicipality returns FormatCity only when the result is a known city.
func FormatMunicipality(value string) string {
	return defaultCityCatalog.FormatMunicipality(value)
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
	slices.Sort(found)
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

func boundedLevenshteinRunes(a, b []rune, maxDist int) int {
	lenA, lenB := len(a), len(b)
	if abs(lenA-lenB) > maxDist {
		return maxDist + 1
	}
	if slices.Equal(a, b) {
		return 0
	}

	var prevStack, curStack [64]int
	var prev, cur []int
	if lenB+1 <= len(prevStack) {
		prev = prevStack[:lenB+1]
		cur = curStack[:lenB+1]
	} else {
		prev = make([]int, lenB+1)
		cur = make([]int, lenB+1)
	}

	for j := 0; j <= lenB; j++ {
		prev[j] = j
	}
	for i := 0; i < lenA; i++ {
		cur[0] = i + 1
		rowMin := cur[0]
		ai := a[i]
		for j := 0; j < lenB; j++ {
			ins := cur[j] + 1
			del := prev[j+1] + 1
			sub := prev[j]
			if ai != b[j] {
				sub++
			}
			val := min(ins, del, sub)
			cur[j+1] = val
			if val < rowMin {
				rowMin = val
			}
		}
		if rowMin > maxDist {
			return maxDist + 1
		}
		prev, cur = cur, prev
	}
	if prev[lenB] > maxDist {
		return maxDist + 1
	}
	return prev[lenB]
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
