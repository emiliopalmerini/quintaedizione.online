package parsers

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

var (
	// Animal stat block patterns (similar to monsters but adapted)
	AnimalACRE    = regexp.MustCompile(`(?i)\*\*classe\s+armatura:\*\*\s*(\d+)`)
	AnimalHPRE    = regexp.MustCompile(`(?i)\*\*punti\s+ferita:\*\*\s*(\d+)\s*\((\d+)d(\d+)(?:\s*\+\s*(\d+))?\)`)
	AnimalSpeedRE = regexp.MustCompile(`(?i)\*\*velocità:\*\*\s*(\d+)\s*(?:metri?|m)`)
	AnimalSizeRE  = regexp.MustCompile(`(?i)(minuscola|piccola|media|grande|enorme|colossale)`)
	AnimalTypeRE  = regexp.MustCompile(`(?i)(animale|bestia)`)
	AnimalPBRE    = regexp.MustCompile(`(?i)PB\s*\+(\d+)`)
)

// ParseAnimali parses Italian D&D 5e animal data from markdown into proper domain objects
func ParseAnimali(lines []string) ([]map[string]interface{}, error) {
	items := splitItemsByH2(lines)
	var animali []map[string]interface{}

	for _, item := range items {
		if len(item.lines) == 0 {
			continue
		}

		animale := parseAnimaleItem(item.title, item.lines)
		if animale != nil {
			animali = append(animali, animale)
		}
	}

	return animali, nil
}

// parseAnimaleItem parses a single animal item into the correct domain structure
func parseAnimaleItem(title string, lines []string) map[string]interface{} {
	name := strings.TrimSpace(title)
	if name == "" {
		return nil
	}

	// Create slug from name
	slug, _ := domain.NewSlug(name)

	content := strings.Join(lines, "\n")
	fullContent := "## " + title + "\n" + content

	// Parse basic stats
	ac := parseAnimalAC(content)
	hp := parseAnimalHP(content)
	velocita := parseAnimalSpeed(content)
	taglia := parseAnimalSize(content)
	tipo := parseAnimalType(content)
	caratteristiche := parseAnimalAbilities(lines)
	tratti := parseAnimalTraits(lines)
	azioni := parseAnimalActions(lines)
	bonusCompetenza := parseAnimalProficiencyBonus(content)

	// Build the animal document matching the domain model exactly
	// Note: _id will be generated automatically by MongoDB
	animale := map[string]interface{}{
		"slug":             string(slug),
		"nome":             name,
		"taglia":           string(taglia),
		"tipo":             string(tipo),
		"ca":               ac,
		"pf":               hp,
		"velocita":         velocita,
		"caratteristiche":  caratteristiche,
		"tratti":           tratti,
		"azioni":           azioni,
		"contenuto":        fullContent,
		"bonus_competenza": bonusCompetenza,
	}

	return animale
}

// parseAnimalAC extracts armor class
func parseAnimalAC(content string) int {
	if match := AnimalACRE.FindStringSubmatch(content); len(match) > 1 {
		if ac, err := strconv.Atoi(match[1]); err == nil {
			return ac
		}
	}
	return 10 // default AC
}

// parseAnimalHP extracts hit points and creates PuntiFerita structure
func parseAnimalHP(content string) map[string]interface{} {
	valore := 1 // default HP
	numero := 1 // default dice count
	facce := 6  // default dice faces
	bonus := 0  // default bonus

	if match := AnimalHPRE.FindStringSubmatch(content); len(match) >= 4 {
		// Extract HP value
		if hp, err := strconv.Atoi(match[1]); err == nil {
			valore = hp
		}

		// Extract dice number
		if diceNum, err := strconv.Atoi(match[2]); err == nil {
			numero = diceNum
		}

		// Extract dice faces
		if diceFaces, err := strconv.Atoi(match[3]); err == nil {
			facce = diceFaces
		}

		// Extract bonus (optional, might be empty)
		if len(match) > 4 && match[4] != "" {
			if diceBonus, err := strconv.Atoi(match[4]); err == nil {
				bonus = diceBonus
			}
		}
	}

	return map[string]interface{}{
		"valore": valore,
		"dadi": map[string]interface{}{
			"numero": numero,
			"facce":  facce,
			"bonus":  bonus,
		},
	}
}

// parseAnimalSpeed extracts speed and creates Velocita structure
func parseAnimalSpeed(content string) map[string]interface{} {
	valore := 9 // default speed in meters
	if match := AnimalSpeedRE.FindStringSubmatch(content); len(match) > 1 {
		if speed, err := strconv.Atoi(match[1]); err == nil {
			valore = speed
		}
	}

	return map[string]interface{}{
		"valore": valore,
		"unita":  "m",
	}
}

// parseAnimalSize extracts creature size
func parseAnimalSize(content string) domain.Taglia {
	if match := AnimalSizeRE.FindStringSubmatch(content); len(match) > 1 {
		switch strings.ToLower(match[1]) {
		case "minuscola":
			return domain.TagliaMinuscola
		case "piccola":
			return domain.TagliaPiccola
		case "media":
			return domain.TagliaMedia
		case "grande":
			return domain.TagliaGrande
		case "enorme":
			return domain.TagliaEnorme
		case "colossale":
			return domain.TagliaColossale
		}
	}
	return domain.TagliaMedia // default size
}

// parseAnimalType extracts creature type
func parseAnimalType(content string) domain.TipoAnimale {
	if match := AnimalTypeRE.FindStringSubmatch(content); len(match) > 1 {
		switch strings.ToLower(match[1]) {
		case "bestia":
			return domain.TipoB
		case "animale":
			return domain.TipoA
		}
	}
	return domain.TipoA // default to Animale
}

// parseAnimalAbilities parses ability scores into Caratteristica array
func parseAnimalAbilities(lines []string) []map[string]interface{} {
	var caratteristiche []map[string]interface{}

	// Default ability scores if not found
	defaultAbilities := map[string]int{
		"FOR": 10,
		"DES": 10,
		"COS": 10,
		"INT": 2,
		"SAG": 10,
		"CAR": 6,
	}

	// Track saving throw proficiencies and modifiers
	savingThrows := make(map[string]int)
	modifiers := make(map[string]int)

	// Look for table-based ability scores (| FOR | 19 | +4 | +4 |)
	tablePattern := regexp.MustCompile(`\|\s*(FOR|DES|CON|INT|SAG|CAR)\s*\|\s*(\d+)\s*\|\s*([+-]?\d+)\s*\|\s*([+-]?\d+)\s*\|`)

	// Also look for simple ability score pattern as fallback
	abilityPattern := regexp.MustCompile(`(?i)(FOR|DES|CON|INT|SAG|CAR)\s+(\d+)`)

	foundAbilities := make(map[string]int)

	for _, line := range lines {
		// Try table format first
		if match := tablePattern.FindStringSubmatch(line); len(match) >= 5 {
			ability := strings.ToUpper(match[1])
			if score, err := strconv.Atoi(match[2]); err == nil {
				foundAbilities[ability] = score
			}

			// Parse modifier (MOD column - third column)
			if modValue, err := strconv.Atoi(match[3]); err == nil {
				modifiers[ability] = modValue
			}

			// Parse saving throw (TS column - fourth column)
			if tsValue, err := strconv.Atoi(match[4]); err == nil {
				savingThrows[ability] = tsValue
			}
		} else {
			// Fallback to simple pattern
			matches := abilityPattern.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) >= 3 {
					ability := strings.ToUpper(match[1])
					if score, err := strconv.Atoi(match[2]); err == nil {
						foundAbilities[ability] = score
					}
				}
			}
		}
	}

	// Merge found abilities with defaults
	for abbr, defaultValue := range defaultAbilities {
		if value, found := foundAbilities[abbr]; found {
			defaultAbilities[abbr] = value
		} else {
			defaultAbilities[abbr] = defaultValue
		}
	}

	// Convert to domain format
	abilityMap := map[string]domain.TipoCaratteristica{
		"FOR": domain.CaratteristicaForza,
		"DES": domain.CaratteristicaDestrezza,
		"CON": domain.CaratteristicaCostituzione,
		"INT": domain.CaratteristicaIntelligenza,
		"SAG": domain.CaratteristicaSaggezza,
		"CAR": domain.CaratteristicaCarisma,
	}

	for abbr, tipo := range abilityMap {
		if valore, exists := defaultAbilities[abbr]; exists {
			// Use parsed modifier, or calculate as fallback
			modifier := (valore - 10) / 2 // fallback calculation
			if parsedMod, found := modifiers[abbr]; found {
				modifier = parsedMod
			}

			// Check if this ability has saving throw proficiency
			hasProficiency := false
			if tsValue, found := savingThrows[abbr]; found {
				// If TS value is different from modifier, it has proficiency
				hasProficiency = tsValue != modifier
			}

			caratteristica := map[string]interface{}{
				"tipo":          string(tipo),
				"valore":        valore,
				"ts_competenza": hasProficiency,
			}
			caratteristiche = append(caratteristiche, caratteristica)
		}
	}

	return caratteristiche
}

// parseAnimalTraits extracts special traits
func parseAnimalTraits(lines []string) []map[string]interface{} {
	var tratti []map[string]interface{}
	var inTraits bool

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for Traits section (case insensitive, markdown heading)
		if strings.HasPrefix(line, "### ") && strings.Contains(strings.ToLower(line), "tratti") {
			inTraits = true
			continue
		}

		// Stop when we hit another section
		if inTraits && strings.HasPrefix(line, "##") {
			break
		}

		if inTraits && line != "" {
			// Look for trait patterns with triple asterisk (***Nome.***)
			traitPattern := regexp.MustCompile(`^\*\*\*([^*]+)\.\*\*\*\s*(.*)$`)
			if match := traitPattern.FindStringSubmatch(line); len(match) >= 3 {
				nome := strings.TrimSpace(match[1])
				descrizione := strings.TrimSpace(match[2])

				if nome != "" {
					tratto := map[string]interface{}{
						"nome":        nome,
						"descrizione": descrizione,
					}
					tratti = append(tratti, tratto)
				}
			}
		}
	}

	return tratti
}

// parseAnimalProficiencyBonus extracts proficiency bonus as int
func parseAnimalProficiencyBonus(content string) int {
	valore := 2 // default proficiency bonus
	if match := AnimalPBRE.FindStringSubmatch(content); len(match) > 1 {
		if pb, err := strconv.Atoi(match[1]); err == nil {
			valore = pb
		}
	}

	return valore
}

// parseAnimalActions extracts actions from the text
func parseAnimalActions(lines []string) []map[string]interface{} {
	var azioni []map[string]interface{}
	var inActions bool

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for Actions section (case insensitive, markdown heading)
		if strings.HasPrefix(line, "### ") && strings.Contains(strings.ToLower(line), "azioni") {
			inActions = true
			continue
		}

		// Stop when we hit another section
		if inActions && strings.HasPrefix(line, "##") {
			break
		}

		if inActions && line != "" {
			// Look for action patterns with triple asterisk (***Nome.***)
			actionPattern := regexp.MustCompile(`^\*\*\*([^*]+)\.\*\*\*\s*(.*)$`)
			if match := actionPattern.FindStringSubmatch(line); len(match) >= 3 {
				nome := strings.TrimSpace(match[1])
				descrizione := strings.TrimSpace(match[2])

				if nome != "" {
					// Parse attack information and recharge
					attacco := parseAttaccoFromDescription(descrizione)
					ricarica := parseRicaricaFromName(nome)

					azione := map[string]interface{}{
						"nome":        nome,
						"descrizione": descrizione,
						"attacco":     attacco,
						"ricarica":    ricarica,
					}
					azioni = append(azioni, azione)
				}
			}
		}
	}

	return azioni
}

// parseRicaricaFromName extracts recharge information from action name
func parseRicaricaFromName(nome string) int {
	// Look for pattern like "Roccia (Ricarica 6)"
	ricaricaPattern := regexp.MustCompile(`\(Ricarica\s+(\d+)\)`)
	if match := ricaricaPattern.FindStringSubmatch(nome); len(match) > 1 {
		if ricarica, err := strconv.Atoi(match[1]); err == nil {
			return ricarica
		}
	}
	return 0 // No recharge
}

// parseAttaccoFromDescription extracts attack information from action description
func parseAttaccoFromDescription(descrizione string) map[string]interface{} {
	// Pattern for: "Tiro per colpire in mischia: +6, portata 1,5 m. 15 (2d10 + 4) danni Perforanti"
	// Or: "Tiro per colpire a distanza: +5, gittata 7,5/15 m. 10 (2d6 + 3) danni Contundenti"
	attackPattern := regexp.MustCompile(`\*Tiro per colpire (?:in mischia|a distanza):\*\s*([+-]?\d+),\s*(?:portata|gittata)\s+([\d,/\s]+)\s*m\.\s*(\d+)\s*\((\d+)d(\d+)(?:\s*[+-]\s*(\d+))?\)\s*danni\s+(\w+)`)

	// Default empty attack
	attacco := map[string]interface{}{
		"tiro_per_colpire": 0,
		"danno": map[string]interface{}{
			"valore": 0,
			"dadi": map[string]interface{}{
				"numero": 0,
				"facce":  0,
				"bonus":  0,
			},
			"tipo": "TODO",
		},
		"portata": map[string]interface{}{
			"valore": 0.0,
			"unita":  "m",
		},
	}

	if match := attackPattern.FindStringSubmatch(descrizione); len(match) >= 8 {
		// Parse attack bonus
		if bonus, err := strconv.Atoi(match[1]); err == nil {
			attacco["tiro_per_colpire"] = bonus
		}

		// Parse range/reach - take first number
		rangeStr := strings.Fields(strings.ReplaceAll(match[2], ",", "."))[0]
		if portata, err := strconv.ParseFloat(rangeStr, 64); err == nil {
			attacco["portata"] = map[string]interface{}{
				"valore": portata,
				"unita":  "m",
			}
		}

		// Parse damage value
		if dannoValore, err := strconv.Atoi(match[3]); err == nil {
			// Parse dice
			diceNum, _ := strconv.Atoi(match[4])
			diceFaces, _ := strconv.Atoi(match[5])
			diceBonus := 0
			if len(match) > 6 && match[6] != "" {
				diceBonus, _ = strconv.Atoi(match[6])
			}

			attacco["danno"] = map[string]interface{}{
				"valore": dannoValore,
				"dadi": map[string]interface{}{
					"numero": diceNum,
					"facce":  diceFaces,
					"bonus":  diceBonus,
				},
				"tipo": match[7], // Keep Italian damage type as-is
			}
		}
	}

	return attacco
}
