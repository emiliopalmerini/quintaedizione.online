package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
	"github.com/google/uuid"
)

// Monster stat block patterns
var (
	MonsterACRE    = regexp.MustCompile(`(?i)classe\s+armatura\s+(\d+)`)
	MonsterHPRE    = regexp.MustCompile(`(?i)punti\s+ferita\s+(\d+)`)
	MonsterSpeedRE = regexp.MustCompile(`(?i)velocità\s+(.+)`)
	MonsterCRRE    = regexp.MustCompile(`(?i)grado\s+di\s+sfida\s+([\d/]+)`)
)

// MonstersStrategy implements the Strategy pattern for parsing monsters
type MonstersStrategy struct {
	*BaseParser
}

// NewMonstersStrategy creates a new monsters parsing strategy
func NewMonstersStrategy() ParsingStrategy {
	return &MonstersStrategy{
		BaseParser: NewBaseParser(
			ContentTypeMonsters,
			"Monsters Parser",
			"Parses D&D 5e monsters from Italian SRD markdown content",
		),
	}
}

// Parse processes monster content and returns domain Mostro objects
func (m *MonstersStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
	if err := m.Validate(content); err != nil {
		return nil, err
	}

	sections := m.ExtractSections(content, 2) // H2 level for monsters
	var monsters []domain.ParsedEntity

	for _, section := range sections {
		if !section.HasContent() {
			continue
		}

		monster, err := m.parseMonsterSection(section)
		if err != nil {
			m.LogParsingProgress("Error parsing monster %s: %v", section.Title, err)
			continue
		}

		if monster != nil {
			monsters = append(monsters, monster)
		}
	}

	return monsters, nil
}

func (m *MonstersStrategy) parseMonsterSection(section Section) (*domain.Mostro, error) {
	if section.Title == "" {
		return nil, fmt.Errorf("monster section has no title")
	}

	content := section.GetCleanContent()
	if len(content) == 0 {
		return nil, fmt.Errorf("monster section has no content")
	}

	// Parse monster stats from content
	stats := m.parseMonsterStats(content)
	
	// Create monster content
	monsterContent := strings.Join(content, "\n")

	// Create domain object - using placeholder values for now
	// These should be properly parsed from the content
	monster := domain.NewMostro(
		uuid.New(),
		section.Title,
		domain.TagliaMedia, // TODO: parse from content
		domain.TipoBestia,  // TODO: parse from content  
		domain.AllineamentoNeutrale, // TODO: parse from content
		1, // grado sfida - TODO: parse from content
		domain.PuntiEsperienza{Base: 200}, // TODO: parse from content
		domain.ClasseArmatura(stats.classeArmatura), // TODO: improve parsing
		domain.PuntiFerita{Valore: stats.puntiFerita}, // TODO: improve parsing
		stats.velocita, // TODO: improve parsing
		[]domain.Caratteristica{}, // TODO: parse from content
		domain.Sensibilita{}, // TODO: parse from content
		domain.TiriSalvezza{}, // TODO: parse from content
		domain.AbilitaMostro{}, // TODO: parse from content
		domain.Immunita{}, // TODO: parse from content
		[]domain.Azione{}, // azioni - TODO: parse from content
		[]domain.Tratto{}, // tratti - TODO: parse from content
		[]domain.ReazioneMostro{}, // reazioni - TODO: parse from content
		[]domain.AzioneLeggendaria{}, // azioni leggendarie - TODO: parse from content
		domain.IncantesimiMostro{}, // incantesimi - TODO: parse from content
		monsterContent,
	)

	return monster, nil
}

// MonsterStats holds parsed monster statistics
type MonsterStats struct {
	classeArmatura int
	puntiFerita    int
	velocita       domain.Velocita
}

// parseMonsterStats parses monster statistics from content lines
func (m *MonstersStrategy) parseMonsterStats(lines []string) MonsterStats {
	stats := MonsterStats{
		velocita: domain.Velocita{},
	}
	
	content := strings.Join(lines, " ")

	// Parse AC
	if match := MonsterACRE.FindStringSubmatch(content); len(match) > 1 {
		if ac, err := strconv.Atoi(match[1]); err == nil {
			stats.classeArmatura = ac
		}
	}

	// Parse HP
	if match := MonsterHPRE.FindStringSubmatch(content); len(match) > 1 {
		if hp, err := strconv.Atoi(match[1]); err == nil {
			stats.puntiFerita = hp
		}
	}

	// Parse Speed - simplified for now, using default walking speed
	if match := MonsterSpeedRE.FindStringSubmatch(content); len(match) > 1 {
		// TODO: Properly parse different speed types from the content
		stats.velocita = domain.Velocita{
			Valore: 9, // default speed in meters
			Unita:  domain.UnitaMetri, // TODO: determine from content
		}
	}

	return stats
}