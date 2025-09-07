package parsers

import (
	"fmt"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
	"github.com/google/uuid"
)

// ClassesStrategy implements the Strategy pattern for parsing classes
type ClassesStrategy struct {
	*BaseParser
}

// NewClassesStrategy creates a new classes parsing strategy
func NewClassesStrategy() ParsingStrategy {
	return &ClassesStrategy{
		BaseParser: NewBaseParser(
			ContentTypeClasses,
			"Classes Parser",
			"Parses D&D 5e classes from Italian SRD markdown content",
		),
	}
}

// Parse processes class content and returns domain Classe objects
func (c *ClassesStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
	if err := c.Validate(content); err != nil {
		return nil, err
	}

	sections := c.ExtractSections(content, 2) // H2 level for classes
	var classes []domain.ParsedEntity

	for _, section := range sections {
		if !section.HasContent() {
			continue
		}

		class, err := c.parseClassSection(section)
		if err != nil {
			c.LogParsingProgress("Error parsing class %s: %v", section.Title, err)
			continue
		}

		if class != nil {
			classes = append(classes, class)
		}
	}

	return classes, nil
}

func (c *ClassesStrategy) parseClassSection(section Section) (*domain.Classe, error) {
	if section.Title == "" {
		return nil, fmt.Errorf("class section has no title")
	}

	content := section.GetCleanContent()
	if len(content) == 0 {
		return nil, fmt.Errorf("class section has no content")
	}

	// Parse class information from content
	classContent := strings.Join(content, "\n")
	description := c.extractClassDescription(content)

	// Create domain object - using placeholder values for now
	// These should be properly parsed from the content
	class := domain.NewClasse(
		uuid.New(),
		section.Title,
		description, // sottotitolo - could be extracted better
		classContent, // markdown content
		domain.Dadi{Numero: 1, Facce: 8, Bonus: 0}, // TODO: parse from content - default d8
		[]domain.Caratteristica{}, // caratteristica primaria - TODO: parse
		[]domain.NomeCaratteristica{}, // salvezze competenze - TODO: parse
		domain.Scelta{}, // abilita competenze opzioni - TODO: parse
		[]string{}, // armi competenze - TODO: parse
		[]domain.CompetenzaArmatura{}, // armature competenze - TODO: parse
		[]domain.StrumentoID{}, // strumenti competenze - TODO: parse
		[]domain.EquipaggiamentoOpzione{}, // equipaggiamento iniziale - TODO: parse
		domain.Multiclasse{}, // multiclasse - TODO: parse
		domain.Progressioni{}, // progressioni - TODO: parse
		domain.Magia{}, // magia - TODO: parse
		[]domain.Privilegio{}, // privilegi - TODO: parse
		[]domain.Sottoclasse{}, // sottoclassi - TODO: parse
		domain.ListaIncantesimi{}, // lista incantesimi - TODO: parse
		domain.Raccomandazioni{}, // raccomandazioni - TODO: parse
		classContent,
	)

	return class, nil
}

// extractClassDescription extracts the class description from content lines
func (c *ClassesStrategy) extractClassDescription(lines []string) string {
	var descriptionLines []string
	
	// Look for text before the first table or structured content
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Stop at tables or structured content
		if strings.HasPrefix(trimmed, "|") || 
		   strings.Contains(trimmed, "Tratti base") ||
		   strings.HasPrefix(trimmed, "####") {
			break
		}
		
		// Skip empty lines and headers
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			descriptionLines = append(descriptionLines, trimmed)
		}
	}
	
	if len(descriptionLines) == 0 {
		return "Classe " + strings.ToLower(c.Name()) // fallback description
	}
	
	// Take first paragraph as description
	description := strings.Join(descriptionLines, " ")
	if len(description) > 200 {
		// Truncate at sentence boundary if too long
		if idx := strings.Index(description[150:], "."); idx >= 0 {
			description = description[:150+idx+1]
		} else {
			description = description[:200] + "..."
		}
	}
	
	return description
}