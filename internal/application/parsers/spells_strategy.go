package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
	"github.com/google/uuid"
)

// SpellsStrategy implements the Strategy pattern for parsing spells
type SpellsStrategy struct {
	*BaseParser
}

// NewSpellsStrategy creates a new spells parsing strategy
func NewSpellsStrategy() ParsingStrategy {
	return &SpellsStrategy{
		BaseParser: NewBaseParser(
			ContentTypeSpells,
			"Spells Parser",
			"Parses D&D 5e spells from Italian SRD markdown content",
		),
	}
}

// Parse processes spell content and returns domain Incantesimo objects
func (s *SpellsStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
	if err := s.Validate(content); err != nil {
		return nil, err
	}

	sections := s.ExtractSections(content, 3)
	var spells []domain.ParsedEntity

	for _, section := range sections {
		if !section.HasContent() {
			continue
		}

		spell, err := s.parseSpellSection(section)
		if err != nil {
			s.LogParsingProgress("Error parsing spell %s: %v", section.Title, err)
			continue
		}

		if spell != nil {
			spells = append(spells, spell)
		}
	}

	return spells, nil
}

func (s *SpellsStrategy) parseSpellSection(section Section) (*domain.Incantesimo, error) {
	if section.Title == "" {
		return nil, fmt.Errorf("spell section has no title")
	}

	content := section.GetCleanContent()
	if len(content) == 0 {
		return nil, fmt.Errorf("spell section has no content")
	}

	metaLine := content[0]
	level, school, classes, err := s.parseSpellMeta(metaLine)
	if err != nil {
		return nil, fmt.Errorf("failed to parse spell metadata: %w", err)
	}

	descriptionLines := content[1:]
	lancio := s.parseSpellDetails(descriptionLines)

	spellContent := strings.Join(descriptionLines, "\n")

	spell := domain.NewIncantesimo(
		uuid.New(),
		section.Title,
		uint8(level),
		s.resolveSchoolUUID(school),
		s.resolveClassUUIDs(classes),
		lancio,
		spellContent,
	)

	return spell, nil
}

func (s *SpellsStrategy) parseSpellMeta(metaLine string) (int, string, []string, error) {
	re := regexp.MustCompile(`(?i)^\s*` +
		`(?:(?:Level\s+(?P<lvl_en>\d{1,2})\s+(?P<school_en>[^()]+))` +
		`|(?P<cantrip_school_en>[^()]+)\s+Cantrip` +
		`|(?:Livello\s+(?P<lvl_it>\d{1,2})\s+(?P<school_it>[^()]+))` +
		`|(?:(?:Trucchetto\s+di\s+(?P<cantrip_school_it1>[^()]+))|(?P<cantrip_school_it2>[^()]+)\s+Trucchetto))` +
		`\s*\((?P<classes>[^)]+)\)\s*$`)

	matches := re.FindStringSubmatch(metaLine)
	if matches == nil {
		return 0, "", nil, fmt.Errorf("could not parse spell metadata from: %s", metaLine)
	}

	names := re.SubexpNames()
	result := make(map[string]string)
	for i, match := range matches {
		if i > 0 && match != "" {
			result[names[i]] = match
		}
	}

	level := 0
	school := ""

	if lvlStr := result["lvl_it"]; lvlStr != "" {
		var err error
		level, err = strconv.Atoi(lvlStr)
		if err != nil {
			return 0, "", nil, fmt.Errorf("invalid Italian level: %s", lvlStr)
		}
		school = strings.TrimSpace(result["school_it"])
	} else if lvlStr := result["lvl_en"]; lvlStr != "" {
		var err error
		level, err = strconv.Atoi(lvlStr)
		if err != nil {
			return 0, "", nil, fmt.Errorf("invalid English level: %s", lvlStr)
		}
		school = s.normalizeSchool(result["school_en"])
	} else if school = result["cantrip_school_it1"]; school != "" {
		level = 0
	} else if school = result["cantrip_school_it2"]; school != "" {
		level = 0
	} else if school = result["cantrip_school_en"]; school != "" {
		level = 0
		school = s.normalizeSchool(school)
	}

	school = s.normalizeSchool(school)

	classesStr := result["classes"]
	if classesStr == "" {
		return 0, "", nil, fmt.Errorf("no classes found in metadata")
	}

	classes := s.parseClasses(classesStr)

	return level, school, classes, nil
}

func (s *SpellsStrategy) normalizeSchool(school string) string {
	school = strings.TrimSpace(school)

	schoolEnToIt := map[string]string{
		"Abjuration":    "Abiurazione",
		"Conjuration":   "Evocazione",
		"Divination":    "Divinazione",
		"Enchantment":   "Ammaliamento",
		"Evocation":     "Invocazione",
		"Illusion":      "Illusione",
		"Necromancy":    "Necromanzia",
		"Transmutation": "Trasmutazione",
	}

	if italian, exists := schoolEnToIt[school]; exists {
		return italian
	}

	schoolItNormalize := map[string]string{
		"abiurazione":   "Abiurazione",
		"evocazione":    "Evocazione",
		"divinazione":   "Divinazione",
		"ammaliamento":  "Ammaliamento",
		"invocazione":   "Invocazione",
		"illusione":     "Illusione",
		"necromanzia":   "Necromanzia",
		"trasmutazione": "Trasmutazione",
	}

	if normalized, exists := schoolItNormalize[strings.ToLower(school)]; exists {
		return normalized
	}

	return school
}

func (s *SpellsStrategy) parseClasses(classesStr string) []string {
	classEnToIt := map[string]string{
		"Barbarian": "Barbaro",
		"Bard":      "Bardo",
		"Cleric":    "Chierico",
		"Druid":     "Druido",
		"Fighter":   "Guerriero",
		"Monk":      "Monaco",
		"Paladin":   "Paladino",
		"Ranger":    "Ranger",
		"Rogue":     "Ladro",
		"Sorcerer":  "Stregone",
		"Warlock":   "Warlock",
		"Wizard":    "Mago",
	}

	var classes []string
	parts := strings.Split(classesStr, ",")

	for _, part := range parts {
		className := strings.TrimSpace(part)

		if italian, exists := classEnToIt[className]; exists {
			classes = append(classes, italian)
		} else {
			classes = append(classes, className)
		}
	}

	return classes
}

func (s *SpellsStrategy) parseSpellDetails(lines []string) domain.Lancio {
	var tempoLancio domain.TempoLancio
	var gittata domain.GittataIncantesimo
	var componenti domain.Componenti
	var durata domain.Durata

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "**Tempo di Lancio:**") || strings.HasPrefix(line, "**Casting Time:**") {
			castingTimeStr := strings.TrimSpace(strings.TrimPrefix(line, strings.Split(line, ":")[0]+":"))
			tempoLancio = s.parseCastingTime(castingTimeStr)
		} else if strings.HasPrefix(line, "**Gittata:**") || strings.HasPrefix(line, "**Range:**") {
			rangeStr := strings.TrimSpace(strings.TrimPrefix(line, strings.Split(line, ":")[0]+":"))
			gittata = s.parseRange(rangeStr)
		} else if strings.HasPrefix(line, "**Componenti:**") || strings.HasPrefix(line, "**Components:**") {
			componentsStr := strings.TrimSpace(strings.TrimPrefix(line, strings.Split(line, ":")[0]+":"))
			componenti = s.parseComponents(componentsStr)
		} else if strings.HasPrefix(line, "**Durata:**") || strings.HasPrefix(line, "**Duration:**") {
			durationStr := strings.TrimSpace(strings.TrimPrefix(line, strings.Split(line, ":")[0]+":"))
			durata = s.parseDuration(durationStr)
		}
	}

	return domain.Lancio{
		Tempo:      tempoLancio,
		Gittata:    gittata,
		Componenti: componenti,
		Durata:     durata,
	}
}

func (s *SpellsStrategy) parseCastingTime(timeStr string) domain.TempoLancio {
	lower := strings.ToLower(timeStr)

	if strings.Contains(lower, "azione") {
		if strings.Contains(lower, "bonus") {
			return domain.TempoLancio{Tipo: domain.TempoAzioneBonus, Valore: 0}
		}
		return domain.TempoLancio{Tipo: domain.TempoAzione, Valore: 0}
	}

	if strings.Contains(lower, "reazione") {
		return domain.TempoLancio{Tipo: domain.TempoReazione, Valore: 0, Nota: timeStr}
	}

	return domain.TempoLancio{Tipo: domain.TempoSpeciale, Valore: 0, Nota: timeStr}
}

func (s *SpellsStrategy) parseRange(rangeStr string) domain.GittataIncantesimo {
	lower := strings.ToLower(rangeStr)

	if strings.Contains(lower, "contatto") || strings.Contains(lower, "touch") {
		return domain.GittataIncantesimo{Tipo: domain.GittataContatto}
	}

	if strings.Contains(lower, "se stesso") || strings.Contains(lower, "self") {
		return domain.GittataIncantesimo{Tipo: domain.GittataSe}
	}

	return domain.GittataIncantesimo{Tipo: domain.GittataSpeciale, Nota: rangeStr}
}

func (s *SpellsStrategy) parseComponents(compStr string) domain.Componenti {
	var comp domain.Componenti

	comp.V = strings.Contains(compStr, "V")
	comp.S = strings.Contains(compStr, "S")
	comp.M = strings.Contains(compStr, "M")

	if comp.M {
		comp.Materiali = compStr
	}

	return comp
}

func (s *SpellsStrategy) parseDuration(durStr string) domain.Durata {
	lower := strings.ToLower(durStr)

	if strings.Contains(lower, "istantaneo") || strings.Contains(lower, "instantaneous") {
		return domain.Durata{Tipo: domain.DurataIstantanea}
	}

	if strings.Contains(lower, "concentrazione") || strings.Contains(lower, "concentration") {
		return domain.Durata{Tipo: domain.DurataConcentrazione, Concentrazione: true}
	}

	return domain.Durata{Tipo: domain.DurataSpeciale, Nota: durStr}
}

func (s *SpellsStrategy) resolveSchoolUUID(school string) uuid.UUID {
	// TODO: resolve actual school UUID from school name
	return uuid.New()
}

func (s *SpellsStrategy) resolveClassUUIDs(classes []string) []uuid.UUID {
	var uuids []uuid.UUID
	for range classes {
		uuids = append(uuids, uuid.New()) // TODO: resolve actual class UUIDs
	}
	return uuids
}
