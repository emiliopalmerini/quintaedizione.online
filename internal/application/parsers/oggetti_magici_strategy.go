package parsers

import (
	"fmt"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

type OggettiMagiciStrategy struct{}

func NewOggettiMagiciStrategy() *OggettiMagiciStrategy {
	return &OggettiMagiciStrategy{}
}

func (s *OggettiMagiciStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
	if len(content) == 0 {
		return nil, ErrEmptyContent
	}

	if err := context.Validate(); err != nil {
		return nil, err
	}

	var entities []domain.ParsedEntity
	currentSection := []string{}
	inSection := false

	for _, line := range content {
		line = strings.TrimSpace(line)
		
		// Skip empty lines, main title, and separator lines
		if line == "" || strings.HasPrefix(line, "# ") || line == "---" {
			continue
		}

		// Check for new magic item section (H2)
		if strings.HasPrefix(line, "## ") {
			// Process previous section if exists
			if inSection && len(currentSection) > 0 {
				magicItem, err := s.parseMagicItemSection(currentSection)
				if err != nil {
					context.Logger.Error(fmt.Sprintf("Failed to parse magic item section: %v", err))
					currentSection = []string{}
					inSection = false
					continue
				}
				entities = append(entities, magicItem)
			}

			// Start new section
			currentSection = []string{line}
			inSection = true
		} else if inSection {
			// Add line to current section
			currentSection = append(currentSection, line)
		}
	}

	// Process last section
	if inSection && len(currentSection) > 0 {
		magicItem, err := s.parseMagicItemSection(currentSection)
		if err != nil {
			context.Logger.Error(fmt.Sprintf("Failed to parse last magic item section: %v", err))
		} else {
			entities = append(entities, magicItem)
		}
	}

	if len(entities) == 0 {
		return nil, ErrEmptyContent
	}

	return entities, nil
}

func (s *OggettiMagiciStrategy) parseMagicItemSection(section []string) (*domain.OggettoMagico, error) {
	if len(section) == 0 {
		return nil, ErrEmptySectionContent
	}

	// Extract name from header
	header := section[0]
	if !strings.HasPrefix(header, "## ") {
		return nil, ErrMissingSectionTitle
	}
	nome := strings.TrimSpace(strings.TrimPrefix(header, "## "))

	// Find metadata line (italics line with type/rarity/attunement)
	var metadataLine string
	var startIndex = 1
	
	for i := 1; i < len(section); i++ {
		line := strings.TrimSpace(section[i])
		if strings.HasPrefix(line, "*") && strings.HasSuffix(line, "*") {
			metadataLine = line
			startIndex = i + 1
			break
		}
	}

	if metadataLine == "" {
		return nil, fmt.Errorf("missing magic item metadata line")
	}

	// Parse metadata
	tipo, rarita, sintonizzazione, err := s.parseMetadata(metadataLine)
	if err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Collect all content
	contenuto := strings.Builder{}
	for i := startIndex; i < len(section); i++ {
		contenuto.WriteString(section[i] + "\n")
	}

	magicItem := domain.NewOggettoMagico(
		nome,
		tipo,
		rarita,
		sintonizzazione,
		strings.TrimSpace(contenuto.String()),
	)

	return magicItem, nil
}

func (s *OggettiMagiciStrategy) parseMetadata(line string) (string, domain.Rarita, bool, error) {
	// Remove asterisks
	line = strings.Trim(line, "*")
	
	// Parse patterns:
	// "Oggetto meraviglioso, molto raro (richiede sintonia)"
	// "Armatura (corazza a scaglie), raro"
	// "Arma (qualsiasi spada), leggendario (richiede sintonia)"
	
	var tipo string
	var rarita domain.Rarita
	var sintonizzazione = false
	
	// Check for attunement
	if strings.Contains(strings.ToLower(line), "richiede sintonia") {
		sintonizzazione = true
		// Remove attunement part for parsing
		line = strings.ReplaceAll(line, "(richiede sintonia)", "")
		line = strings.ReplaceAll(line, " richiede sintonia", "")
		line = strings.TrimSpace(line)
	}
	
	// Split by comma to separate type and rarity
	parts := strings.Split(line, ",")
	if len(parts) >= 2 {
		tipo = strings.TrimSpace(parts[0])
		rarityStr := strings.TrimSpace(parts[1])
		
		// Parse rarity
		rarita = s.parseRarita(rarityStr)
	} else {
		// Only one part, try to determine if it's type or rarity
		single := strings.TrimSpace(parts[0])
		if s.isRarityString(single) {
			rarita = s.parseRarita(single)
			tipo = "Oggetto meraviglioso" // default type
		} else {
			tipo = single
			rarita = domain.RaritaComune // default rarity
		}
	}
	
	return tipo, rarita, sintonizzazione, nil
}

func (s *OggettiMagiciStrategy) isRarityString(str string) bool {
	rarities := []string{
		"comune", "non comune", "raro", "molto raro", "leggendario", "artefatto",
	}
	str = strings.ToLower(strings.TrimSpace(str))
	for _, rarity := range rarities {
		if str == rarity {
			return true
		}
	}
	return false
}

func (s *OggettiMagiciStrategy) parseRarita(value string) domain.Rarita {
	value = strings.ToLower(strings.TrimSpace(value))
	
	switch value {
	case "comune":
		return domain.RaritaComune
	case "non comune":
		return domain.RaritaNonComune
	case "raro":
		return domain.RaritaRara
	case "molto raro":
		return domain.RaritaMoltoRara
	case "leggendario":
		return domain.RaritaLeggendaria
	case "artefatto":
		return domain.RaritaArtefatto
	default:
		return domain.RaritaComune
	}
}

func (s *OggettiMagiciStrategy) ContentType() ContentType {
	return ContentTypeOggettiMagici
}

func (s *OggettiMagiciStrategy) Name() string {
	return "Oggetti Magici Strategy"
}

func (s *OggettiMagiciStrategy) Description() string {
	return "Parses Italian D&D 5e magic items from markdown content"
}

func (s *OggettiMagiciStrategy) Validate(content []string) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}
	return nil
}