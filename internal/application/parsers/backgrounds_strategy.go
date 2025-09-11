package parsers

import (
	"fmt"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

type BackgroundsStrategy struct{}

func NewBackgroundsStrategy() *BackgroundsStrategy {
	return &BackgroundsStrategy{}
}

func (s *BackgroundsStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
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
		
		// Skip empty lines and main title
		if line == "" || strings.HasPrefix(line, "# ") {
			continue
		}

		// Check for new background section (H2)
		if strings.HasPrefix(line, "## ") {
			// Process previous section if exists
			if inSection && len(currentSection) > 0 {
				// Skip non-background sections like "Specie del Personaggio"
				if !s.isBackgroundSection(currentSection[0]) {
					currentSection = []string{}
					inSection = false
					continue
				}
				
				background, err := s.parseBackgroundSection(currentSection)
				if err != nil {
					context.Logger.Error(fmt.Sprintf("Failed to parse background section: %v", err))
					currentSection = []string{}
					inSection = false
					continue
				}
				entities = append(entities, background)
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
	if inSection && len(currentSection) > 0 && s.isBackgroundSection(currentSection[0]) {
		background, err := s.parseBackgroundSection(currentSection)
		if err != nil {
			context.Logger.Error(fmt.Sprintf("Failed to parse last background section: %v", err))
		} else {
			entities = append(entities, background)
		}
	}

	if len(entities) == 0 {
		return nil, ErrEmptyContent
	}

	return entities, nil
}

func (s *BackgroundsStrategy) isBackgroundSection(header string) bool {
	// Skip sections that are not backgrounds
	sectionName := strings.TrimSpace(strings.TrimPrefix(header, "## "))
	skipSections := []string{
		"Specie del Personaggio",
		"Parti di una Specie",
		"Tipo di Creatura",
		"Taglia",
		"Velocità",
		"Tratti Speciali",
	}
	
	for _, skip := range skipSections {
		if sectionName == skip {
			return false
		}
	}
	return true
}

func (s *BackgroundsStrategy) parseBackgroundSection(section []string) (*domain.Background, error) {
	if len(section) == 0 {
		return nil, ErrEmptySectionContent
	}

	// Extract name from header
	header := section[0]
	if !strings.HasPrefix(header, "## ") {
		return nil, ErrMissingSectionTitle
	}
	nome := strings.TrimSpace(strings.TrimPrefix(header, "## "))

	// Parse fields
	fields := make(map[string]string)
	contenuto := strings.Builder{}
	
	for i := 1; i < len(section); i++ {
		line := section[i]
		contenuto.WriteString(line + "\n")
		
		// Parse field format: **Field:** value
		if strings.HasPrefix(line, "**") && strings.Contains(line, ":**") {
			parts := strings.SplitN(line, ":**", 2)
			if len(parts) == 2 {
				fieldName := strings.TrimSpace(strings.Trim(parts[0], "*"))
				fieldValue := strings.TrimSpace(parts[1])
				fields[fieldName] = fieldValue
			}
		}
	}

	// Parse required fields
	caratteristiche := s.parseCaratteristiche(fields["Punteggi di Caratteristica"])
	competenzeAbilita := s.parseCompetenzeAbilita(fields["Competenze in Abilità"])
	competenzeStrumenti := s.parseCompetenzeStrumenti(fields["Competenza negli Strumenti"])
	talento := s.parseTalento(fields["Talento"])
	equipaggiamento := s.parseEquipaggiamento(fields["Equipaggiamento"])

	background := domain.NewBackground(
		nome,
		caratteristiche,
		competenzeAbilita,
		competenzeStrumenti,
		talento,
		equipaggiamento,
		strings.TrimSpace(contenuto.String()),
	)

	return background, nil
}

func (s *BackgroundsStrategy) parseCaratteristiche(value string) []string {
	if value == "" {
		return []string{}
	}
	
	// Split by comma and clean up
	parts := strings.Split(value, ",")
	var caratteristiche []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			caratteristiche = append(caratteristiche, part)
		}
	}
	return caratteristiche
}

func (s *BackgroundsStrategy) parseCompetenzeAbilita(value string) []string {
	if value == "" {
		return []string{}
	}
	
	// Split by " e " or comma
	value = strings.ReplaceAll(value, " e ", ",")
	parts := strings.Split(value, ",")
	var competenze []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			competenze = append(competenze, part)
		}
	}
	return competenze
}

func (s *BackgroundsStrategy) parseCompetenzeStrumenti(value string) []string {
	if value == "" {
		return []string{}
	}
	
	// Handle complex cases like "*Scegli un tipo di Set da Gioco* (vedi "Equipaggiamento")"
	if strings.Contains(value, "*") {
		// Extract text between asterisks
		parts := strings.Split(value, "*")
		if len(parts) >= 3 {
			value = strings.TrimSpace(parts[1])
		}
	}
	
	// Split by comma and clean up
	parts := strings.Split(value, ",")
	var strumenti []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			strumenti = append(strumenti, part)
		}
	}
	return strumenti
}

func (s *BackgroundsStrategy) parseTalento(value string) string {
	if value == "" {
		return ""
	}
	
	// Extract talent name, removing parentheses and references
	// Example: "Iniziato alla Magia (Chierico) (vedi "Talenti")" -> "Iniziato alla Magia (Chierico)"
	if strings.Contains(value, "(vedi") {
		parts := strings.Split(value, "(vedi")
		if len(parts) > 0 {
			value = strings.TrimSpace(parts[0])
		}
	}
	
	return value
}

func (s *BackgroundsStrategy) parseEquipaggiamento(value string) domain.Scelta {
	if value == "" {
		return domain.NewScelta(1, []any{})
	}
	
	// Parse equipment choices - format like "*Scegli A o B:* (A) items; oppure (B) items"
	if !strings.Contains(value, "Scegli A o B") {
		// Single option
		return domain.NewScelta(1, []any{value})
	}
	
	// Extract options A and B
	var opzioni []any
	
	// Find (A) section
	if aStart := strings.Index(value, "(A)"); aStart != -1 {
		remaining := value[aStart+3:]
		var aEnd int
		if strings.Contains(remaining, "oppure (B)") {
			aEnd = strings.Index(remaining, "oppure (B)")
		} else if strings.Contains(remaining, "; oppure") {
			aEnd = strings.Index(remaining, "; oppure")
		} else {
			aEnd = len(remaining)
		}
		
		if aEnd > 0 {
			optionA := strings.TrimSpace(remaining[:aEnd])
			optionA = strings.TrimSuffix(optionA, ";")
			opzioni = append(opzioni, strings.TrimSpace(optionA))
		}
	}
	
	// Find (B) section
	if bStart := strings.Index(value, "(B)"); bStart != -1 {
		remaining := value[bStart+3:]
		optionB := strings.TrimSpace(remaining)
		opzioni = append(opzioni, optionB)
	}
	
	return domain.NewScelta(1, opzioni)
}

func (s *BackgroundsStrategy) ContentType() ContentType {
	return ContentTypeBackgrounds
}

func (s *BackgroundsStrategy) Name() string {
	return "Backgrounds Strategy"
}

func (s *BackgroundsStrategy) Description() string {
	return "Parses Italian D&D 5e character backgrounds from markdown content"
}

func (s *BackgroundsStrategy) Validate(content []string) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}
	return nil
}