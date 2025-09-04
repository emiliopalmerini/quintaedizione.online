package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// Spell parsing constants
var (
	SpellMetaRE = regexp.MustCompile(`(?i)^\s*` +
		`(?:(?:Level\s+(?P<lvl_en>\d{1,2})\s+(?P<school_en>[^()]+))` + // EN: Level X School
		`|(?P<cantrip_school_en>[^()]+)\s+Cantrip` + // EN: School Cantrip
		`|(?:Livello\s+(?P<lvl_it>\d{1,2})\s+(?P<school_it>[^()]+))` + // IT: Livello X Scuola
		`|(?:(?:Trucchetto\s+di\s+(?P<cantrip_school_it1>[^()]+))|(?P<cantrip_school_it2>[^()]+)\s+Trucchetto))` + // IT: Trucchetto di Scuola / Scuola Trucchetto
		`\s*\((?P<classes>[^)]+)\)\s*$`)

	SectionH3RE = regexp.MustCompile(`^### (.+)$`)
)

// School mappings
var (
	SchoolEnToIt = map[string]string{
		"Abjuration":    "Abiurazione",
		"Conjuration":   "Evocazione",
		"Divination":    "Divinazione",
		"Enchantment":   "Ammaliament",
		"Evocation":     "Invocazione",
		"Illusion":      "Illusione",
		"Necromancy":    "Necromanzia",
		"Transmutation": "Trasmutazione",
	}

	SchoolItNormalize = map[string]string{
		"abiurazione":   "Abiurazione",
		"evocazione":    "Evocazione",
		"divinazione":   "Divinazione",
		"ammaliamento":  "Ammaliamento",
		"invocazione":   "Invocazione",
		"illusione":     "Illusione",
		"necromanzia":   "Necromanzia",
		"trasmutazione": "Trasmutazione",
	}

	ClassEnToIt = map[string]string{
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

	ClassItNormalize = map[string]string{
		"barbaro":   "Barbaro",
		"bardo":     "Bardo",
		"chierico":  "Chierico",
		"druido":    "Druido",
		"guerriero": "Guerriero",
		"monaco":    "Monaco",
		"paladino":  "Paladino",
		"ranger":    "Ranger",
		"ladro":     "Ladro",
		"stregone":  "Stregone",
		"warlock":   "Warlock",
		"mago":      "Mago",
	}
)

// SpellParser handles parsing of spell data
type SpellParser struct {
	context *domain.ParsingContext
}

// NewSpellParser creates a new spell parser
func NewSpellParser(context *domain.ParsingContext) *SpellParser {
	return &SpellParser{
		context: context,
	}
}

// ParseSpells parses spell data from markdown lines
func ParseSpells(lines []string) ([]map[string]interface{}, error) {
	context := &domain.ParsingContext{
		Language: domain.ExtractLanguageFromPath("spells"),
		Source:   "SRD",
	}
	
	parser := NewSpellParser(context)
	return parser.Parse(lines)
}

// Parse parses the spell lines
func (p *SpellParser) Parse(lines []string) ([]map[string]interface{}, error) {
	items := p.splitItems(lines)
	var results []map[string]interface{}

	for _, itemLines := range items {
		if len(itemLines) == 0 {
			continue
		}

		spell, err := p.parseSpellItem(itemLines)
		if err != nil {
			// Log error but continue with other spells
			continue
		}

		if spell != nil {
			results = append(results, spell)
		}
	}

	return results, nil
}

// splitItems splits lines into individual spell items
func (p *SpellParser) splitItems(lines []string) [][]string {
	var items [][]string
	var currentItem []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip empty lines at the start
		if len(currentItem) == 0 && line == "" {
			continue
		}

		// H3 header starts a new item
		if SectionH3RE.MatchString(line) {
			if len(currentItem) > 0 {
				items = append(items, currentItem)
			}
			currentItem = []string{line}
			continue
		}

		// Add line to current item
		if len(currentItem) > 0 || line != "" {
			currentItem = append(currentItem, line)
		}
	}

	// Add final item
	if len(currentItem) > 0 {
		items = append(items, currentItem)
	}

	return items
}

// parseSpellItem parses a single spell item
func (p *SpellParser) parseSpellItem(lines []string) (map[string]interface{}, error) {
	if len(lines) < 2 {
		return nil, fmt.Errorf("insufficient lines for spell")
	}

	// Extract name from H3 header
	nameMatch := SectionH3RE.FindStringSubmatch(lines[0])
	if len(nameMatch) < 2 {
		return nil, fmt.Errorf("invalid spell header: %s", lines[0])
	}
	name := strings.TrimSpace(nameMatch[1])

	// Parse meta line (second line)
	if len(lines) < 2 {
		return nil, fmt.Errorf("missing meta line for spell: %s", name)
	}

	meta := p.parseMeta(lines[1])
	if len(meta) == 0 {
		return nil, fmt.Errorf("failed to parse meta for spell: %s", name)
	}

	// Extract description and other fields
	contentLines := lines[2:]
	fields := p.collectLabeledFields(contentLines)

	// Build spell object
	spell := map[string]interface{}{
		"nome":                name,
		"slug":               domain.NormalizeID(name), // Use slug for MongoDB compatibility
		"livello":            meta["level"],
		"scuola":             meta["school"],
		"classi":             meta["classes"],
		"contenuto_markdown": strings.Join(lines, "\n"),
		"fonte":              p.context.Source,
		"versione":           "1.0",
	}

	// Add parsed fields
	for key, value := range fields {
		spell[key] = value
	}

	// Set casting information
	if tempo, ok := fields["tempo_lancio"]; ok {
		spell["tempo_lancio"] = tempo
	}
	if gittata, ok := fields["gittata"]; ok {
		spell["gittata"] = gittata
	}
	if durata, ok := fields["durata"]; ok {
		spell["durata"] = durata
	}
	if componenti, ok := fields["componenti"]; ok {
		spell["componenti"] = componenti
	}

	return spell, nil
}

// parseMeta parses the metadata line
func (p *SpellParser) parseMeta(line string) map[string]interface{} {
	result := make(map[string]interface{})

	matches := SpellMetaRE.FindStringSubmatch(strings.TrimSpace(line))
	if matches == nil {
		return result
	}

	// Create named group map
	names := SpellMetaRE.SubexpNames()
	groups := make(map[string]string)
	for i, match := range matches {
		if i > 0 && names[i] != "" {
			groups[names[i]] = match
		}
	}

	// Determine level and school
	var level int
	var school string

	if lvlStr := groups["lvl_en"]; lvlStr != "" {
		if l, err := strconv.Atoi(lvlStr); err == nil {
			level = l
		}
		school = groups["school_en"]
	} else if lvlStr := groups["lvl_it"]; lvlStr != "" {
		if l, err := strconv.Atoi(lvlStr); err == nil {
			level = l
		}
		school = groups["school_it"]
	} else {
		// Cantrip
		level = 0
		school = groups["cantrip_school_en"]
		if school == "" {
			school = groups["cantrip_school_it1"]
		}
		if school == "" {
			school = groups["cantrip_school_it2"]
		}
	}

	// Normalize school to Italian
	school = p.normalizeSchoolToIt(school)

	// Parse classes
	var classes []string
	if classesStr := groups["classes"]; classesStr != "" {
		classList := strings.Split(classesStr, ",")
		for _, class := range classList {
			class = strings.TrimSpace(class)
			if class != "" {
				classes = append(classes, class)
			}
		}
		classes = p.normalizeClassesToIt(classes)
	}

	result["level"] = level
	result["school"] = school
	result["classes"] = classes

	return result
}

// normalizeSchoolToIt normalizes school name to Italian
func (p *SpellParser) normalizeSchoolToIt(name string) string {
	if name == "" {
		return ""
	}

	name = strings.TrimSpace(name)

	// Try English mapping first
	if it, ok := SchoolEnToIt[name]; ok {
		return it
	}

	// Normalize Italian capitalization
	if it, ok := SchoolItNormalize[strings.ToLower(name)]; ok {
		return it
	}

	return name
}

// normalizeClassesToIt normalizes class names to Italian
func (p *SpellParser) normalizeClassesToIt(classes []string) []string {
	var result []string

	for _, class := range classes {
		class = strings.TrimSpace(class)
		if class == "" {
			continue
		}

		// Try English mapping first
		if it, ok := ClassEnToIt[class]; ok {
			result = append(result, it)
			continue
		}

		// Normalize Italian capitalization
		if it, ok := ClassItNormalize[strings.ToLower(class)]; ok {
			result = append(result, it)
		} else {
			result = append(result, class)
		}
	}

	return result
}

// collectLabeledFields extracts labeled fields from content lines
func (p *SpellParser) collectLabeledFields(lines []string) map[string]string {
	fields := make(map[string]string)
	
	var currentLabel string
	var currentContent []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if line is a label (bold text)
		if strings.HasPrefix(line, "**") && strings.HasSuffix(line, "**") {
			// Save previous field
			if currentLabel != "" {
				fields[p.normalizeLabel(currentLabel)] = strings.Join(currentContent, " ")
			}

			// Start new field
			currentLabel = strings.Trim(line, "* ")
			currentContent = []string{}
		} else if strings.Contains(line, ":") && currentLabel == "" {
			// Handle "Label: Content" format
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				label := strings.TrimSpace(parts[0])
				content := strings.TrimSpace(parts[1])
				fields[p.normalizeLabel(label)] = content
			}
		} else {
			// Add to current content
			currentContent = append(currentContent, line)
		}
	}

	// Save last field
	if currentLabel != "" {
		fields[p.normalizeLabel(currentLabel)] = strings.Join(currentContent, " ")
	}

	return fields
}

// normalizeLabel normalizes field labels
func (p *SpellParser) normalizeLabel(label string) string {
	label = strings.ToLower(strings.TrimSpace(label))
	
	// Map English to Italian labels
	labelMap := map[string]string{
		"casting time":   "tempo_lancio",
		"range":          "gittata",
		"components":     "componenti",
		"duration":       "durata",
		"at higher levels": "livelli_superiori",
		"tempo di lancio": "tempo_lancio",
		"ai livelli superiori": "livelli_superiori",
	}

	if mapped, ok := labelMap[label]; ok {
		return mapped
	}

	// Default normalization
	label = strings.ReplaceAll(label, " ", "_")
	label = strings.ReplaceAll(label, "-", "_")
	return label
}