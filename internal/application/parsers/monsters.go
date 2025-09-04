package parsers

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	// Monster stat block patterns
	MonsterACRE     = regexp.MustCompile(`(?i)classe\s+armatura\s+(\d+)`)
	MonsterHPRE     = regexp.MustCompile(`(?i)punti\s+ferita\s+(\d+)`)
	MonsterSpeedRE  = regexp.MustCompile(`(?i)velocità\s+(.+)`)
	MonsterCRRE     = regexp.MustCompile(`(?i)grado\s+di\s+sfida\s+([\d/]+)`)
)

// ParseMonstersMonster parses Italian D&D 5e monster data from markdown
func ParseMonstersMonster(lines []string) ([]map[string]interface{}, error) {
	return parseMonsters(lines, "monster")
}

// ParseMonstersAnimal parses Italian D&D 5e animal data from markdown  
func ParseMonstersAnimal(lines []string) ([]map[string]interface{}, error) {
	return parseMonsters(lines, "animal")
}

// parseMonsters parses monsters with a given namespace
func parseMonsters(lines []string, namespace string) ([]map[string]interface{}, error) {
	items := splitItemsByH2(lines)
	var monsters []map[string]interface{}

	for _, item := range items {
		if len(item.lines) == 0 {
			continue
		}

		monster := parseMonsterItem(item.title, item.lines, namespace)
		if monster != nil {
			monsters = append(monsters, monster)
		}
	}

	return monsters, nil
}

// parseMonsterItem parses a single monster
func parseMonsterItem(title string, lines []string, namespace string) map[string]interface{} {
	name := strings.TrimSpace(title)
	if name == "" {
		return nil
	}

	// Extract monster stats
	stats := parseMonsterStats(lines)

	// Build monster object
	monster := map[string]interface{}{
		"slug":                name,
		"nome":                name,
		"namespace":           namespace,
		"contenuto_markdown":  strings.Join(append([]string{"## " + title}, lines...), "\n"),
		"fonte":               "SRD",
		"versione":            "1.0",
	}

	// Add parsed stats
	for key, value := range stats {
		monster[key] = value
	}

	// Extract description
	description := extractMonsterDescription(lines)
	if description != "" {
		monster["descrizione"] = description
	}

	return monster
}

// parseMonsterStats parses monster statistics from text
func parseMonsterStats(lines []string) map[string]interface{} {
	stats := make(map[string]interface{})
	content := strings.Join(lines, " ")

	// Parse AC
	if match := MonsterACRE.FindStringSubmatch(content); len(match) > 1 {
		if ac, err := strconv.Atoi(match[1]); err == nil {
			stats["classe_armatura"] = ac
		}
	}

	// Parse HP  
	if match := MonsterHPRE.FindStringSubmatch(content); len(match) > 1 {
		if hp, err := strconv.Atoi(match[1]); err == nil {
			stats["punti_ferita"] = hp
		}
	}

	// Parse Speed
	if match := MonsterSpeedRE.FindStringSubmatch(content); len(match) > 1 {
		stats["velocita"] = strings.TrimSpace(match[1])
	}

	// Parse Challenge Rating
	if match := MonsterCRRE.FindStringSubmatch(content); len(match) > 1 {
		stats["grado_sfida"] = strings.TrimSpace(match[1])
	}

	// Parse abilities (STR, DEX, CON, INT, WIS, CHA)
	abilities := parseAbilityScores(lines)
	if len(abilities) > 0 {
		stats["punteggi_caratteristica"] = abilities
	}

	return stats
}

// parseAbilityScores parses ability scores from monster text
func parseAbilityScores(lines []string) map[string]int {
	abilities := make(map[string]int)
	
	// Look for ability score pattern
	abilityPattern := regexp.MustCompile(`(?i)(FOR|DES|CON|INT|SAG|CAR)\s+(\d+)`)
	
	for _, line := range lines {
		matches := abilityPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				ability := strings.ToUpper(match[1])
				if score, err := strconv.Atoi(match[2]); err == nil {
					abilities[ability] = score
				}
			}
		}
	}

	return abilities
}

// extractMonsterDescription extracts monster description
func extractMonsterDescription(lines []string) string {
	var descLines []string
	inStatBlock := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip stat block lines
		if strings.Contains(line, "Classe Armatura") ||
		   strings.Contains(line, "Punti Ferita") ||
		   strings.Contains(line, "Velocità") ||
		   strings.Contains(line, "FOR") ||
		   strings.Contains(line, "Grado di Sfida") {
			inStatBlock = true
			continue
		}

		// Stop at stat block
		if inStatBlock {
			continue
		}

		if line != "" {
			descLines = append(descLines, line)
		}
	}

	return strings.Join(descLines, "\n")
}