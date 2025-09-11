package parsers

import (
	"fmt"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

type TalentiStrategy struct{}

func NewTalentiStrategy() *TalentiStrategy {
	return &TalentiStrategy{}
}

func (s *TalentiStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
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

		// Check for new talent section (H2)
		if strings.HasPrefix(line, "## ") {
			// Process previous section if exists
			if inSection && len(currentSection) > 0 {
				talent, err := s.parseTalentSection(currentSection)
				if err != nil {
					context.Logger.Error(fmt.Sprintf("Failed to parse talent section: %v", err))
					currentSection = []string{}
					inSection = false
					continue
				}
				entities = append(entities, talent)
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
		talent, err := s.parseTalentSection(currentSection)
		if err != nil {
			context.Logger.Error(fmt.Sprintf("Failed to parse last talent section: %v", err))
		} else {
			entities = append(entities, talent)
		}
	}

	if len(entities) == 0 {
		return nil, ErrEmptyContent
	}

	return entities, nil
}

func (s *TalentiStrategy) parseTalentSection(section []string) (*domain.Talento, error) {
	if len(section) == 0 {
		return nil, ErrEmptySectionContent
	}

	// Extract name from header
	header := section[0]
	if !strings.HasPrefix(header, "## ") {
		return nil, ErrMissingSectionTitle
	}
	nome := strings.TrimSpace(strings.TrimPrefix(header, "## "))

	// Find category line (italics line)
	var categoriaLine string
	var startIndex = 1
	
	for i := 1; i < len(section); i++ {
		line := strings.TrimSpace(section[i])
		if strings.HasPrefix(line, "*") && strings.HasSuffix(line, "*") {
			categoriaLine = line
			startIndex = i + 1
			break
		}
	}

	// Parse category and prerequisites
	categoria, prerequisiti := s.parseCategoriaEPrerequisiti(categoriaLine)

	// Parse benefits and content
	var benefici []string
	contenuto := strings.Builder{}
	
	for i := startIndex; i < len(section); i++ {
		line := section[i]
		contenuto.WriteString(line + "\n")
		
		// Parse benefits - lines starting with **
		if strings.HasPrefix(line, "**") && strings.Contains(line, ".**") {
			parts := strings.SplitN(line, ".**", 2)
			if len(parts) == 2 {
				beneficioTitolo := strings.TrimSpace(strings.Trim(parts[0], "*"))
				beneficioDescrizione := strings.TrimSpace(parts[1])
				beneficio := beneficioTitolo + ". " + beneficioDescrizione
				benefici = append(benefici, beneficio)
			}
		}
	}

	// If no benefits found, try to extract from general content
	if len(benefici) == 0 {
		contenutoText := contenuto.String()
		// Look for structured benefit text
		lines := strings.Split(contenutoText, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "*") && 
			   !strings.Contains(line, "Ottieni i seguenti benefici") &&
			   !strings.Contains(line, "Ripetibile") {
				// This might be a benefit description
				if len(line) > 10 { // Avoid very short lines
					benefici = append(benefici, line)
				}
			}
		}
	}

	talent := domain.NewTalento(
		nome,
		categoria,
		prerequisiti,
		benefici,
		strings.TrimSpace(contenuto.String()),
	)

	return talent, nil
}

func (s *TalentiStrategy) parseCategoriaEPrerequisiti(line string) (domain.CategoriaTalento, string) {
	if line == "" {
		return domain.CategoriaTalentoGenerale, ""
	}

	// Remove asterisks
	line = strings.Trim(line, "*")
	
	// Check for prerequisites in parentheses
	var categoria string
	var prerequisiti string
	
	if strings.Contains(line, "(") {
		// Format like "Talento di Categoria (Prerequisiti)"
		parenStart := strings.Index(line, "(")
		parenEnd := strings.LastIndex(line, ")")
		
		if parenStart != -1 && parenEnd != -1 && parenEnd > parenStart {
			categoria = strings.TrimSpace(line[:parenStart])
			prerequisiti = strings.TrimSpace(line[parenStart+1 : parenEnd])
		} else {
			categoria = line
		}
	} else {
		categoria = line
	}

	// Map to domain category
	var categoriaDomain domain.CategoriaTalento
	switch categoria {
	case "Talento di Origine":
		categoriaDomain = domain.CategoriaTalentoOrigine
	case "Combattimento":
		categoriaDomain = domain.CategoriaTalentoCombat
	case "Magia":
		categoriaDomain = domain.CategoriaTalentoMagia
	case "Abilità":
		categoriaDomain = domain.CategoriaTalentoAbilita
	case "Razziale":
		categoriaDomain = domain.CategoriaTalentoRazziale
	default:
		categoriaDomain = domain.CategoriaTalentoGenerale
	}

	return categoriaDomain, prerequisiti
}

func (s *TalentiStrategy) ContentType() ContentType {
	return ContentTypeTalenti
}

func (s *TalentiStrategy) Name() string {
	return "Talenti Strategy"
}

func (s *TalentiStrategy) Description() string {
	return "Parses Italian D&D 5e feats (talenti) from markdown content"
}

func (s *TalentiStrategy) Validate(content []string) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}
	return nil
}