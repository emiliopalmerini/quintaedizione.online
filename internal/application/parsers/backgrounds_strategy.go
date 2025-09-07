package parsers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
	"github.com/google/uuid"
)

var (
	// Bold field regex for backgrounds
	BackgroundBoldFieldRE = regexp.MustCompile(`\*\*([^*]+)\*\*`)
)

// BackgroundsStrategy implements the Strategy pattern for parsing backgrounds
type BackgroundsStrategy struct {
	*BaseParser
}

// NewBackgroundsStrategy creates a new backgrounds parsing strategy
func NewBackgroundsStrategy() ParsingStrategy {
	return &BackgroundsStrategy{
		BaseParser: NewBaseParser(
			ContentTypeBackgrounds,
			"Backgrounds Parser",
			"Parses D&D 5e backgrounds from Italian SRD markdown content",
		),
	}
}

// Parse processes background content and returns domain Background objects
func (b *BackgroundsStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
	if err := b.Validate(content); err != nil {
		return nil, err
	}

	sections := b.ExtractSections(content, 2) // H2 level for backgrounds
	var backgrounds []domain.ParsedEntity

	for _, section := range sections {
		if !section.HasContent() {
			continue
		}

		background, err := b.parseBackgroundSection(section)
		if err != nil {
			b.LogParsingProgress("Error parsing background %s: %v", section.Title, err)
			continue
		}

		if background != nil {
			backgrounds = append(backgrounds, background)
		}
	}

	return backgrounds, nil
}

func (b *BackgroundsStrategy) parseBackgroundSection(section Section) (*domain.Background, error) {
	if section.Title == "" {
		return nil, fmt.Errorf("background section has no title")
	}

	content := section.GetCleanContent()
	if len(content) == 0 {
		return nil, fmt.Errorf("background section has no content")
	}

	// Parse background information from content
	backgroundContent := strings.Join(content, "\n")
	
	// TODO: Extract skills, tools, languages from the content
	// This is simplified - in reality we'd parse the structured content

	// Create domain object - using placeholder values for now
	// These should be properly parsed from the content
	background := domain.NewBackground(
		uuid.New(),
		section.Title,
		[]uuid.UUID{}, // caratteristiche - TODO: parse from content
		[]uuid.UUID{}, // abilita - TODO: parse from content
		[]uuid.UUID{}, // strumenti - TODO: parse from content
		uuid.Nil,      // talento - TODO: parse from content
		domain.Scelta{}, // equipaggiamento iniziale - TODO: parse
		backgroundContent,
	)

	return background, nil
}

// TODO: Add helper methods to extract:
// - extractSkills: skill competencies from background content  
// - extractTools: tool proficiencies from background content
// - extractLanguages: language proficiencies from background content
// These would parse structured content looking for specific patterns