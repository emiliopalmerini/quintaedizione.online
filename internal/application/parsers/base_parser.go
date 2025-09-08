package parsers

import (
	"fmt"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// BaseParser provides common functionality and Template Method for all parsing strategies
type BaseParser struct {
	contentType ContentType
	name        string
	description string
	language    LanguageCode
	config      *LanguageConfig
	logger      Logger
}

// NewBaseParser creates a new base parser
func NewBaseParser(contentType ContentType, name, description string) *BaseParser {
	return &BaseParser{
		contentType: contentType,
		name:        name,
		description: description,
		language:    Italian, // default to Italian for backward compatibility
		logger:      &NoOpLogger{},
	}
}

// NewBaseParserWithLanguage creates a new base parser with language support
func NewBaseParserWithLanguage(contentType ContentType, name, description string, language LanguageCode, config *LanguageConfig) *BaseParser {
	return &BaseParser{
		contentType: contentType,
		name:        name,
		description: description,
		language:    language,
		config:      config,
		logger:      &NoOpLogger{},
	}
}

// WithLogger sets the logger for this parser (allows chaining)
func (bp *BaseParser) WithLogger(logger Logger) *BaseParser {
	bp.logger = logger
	return bp
}

// ContentType returns the content type this parser handles
func (bp *BaseParser) ContentType() ContentType {
	return bp.contentType
}

// Name returns a human-readable name for this parser
func (bp *BaseParser) Name() string {
	return bp.name
}

// Description returns a description of what this parser does
func (bp *BaseParser) Description() string {
	return bp.description
}

// Validate performs common validation on content
func (bp *BaseParser) Validate(content []string) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}

	hasContent := false
	for _, line := range content {
		if strings.TrimSpace(line) != "" {
			hasContent = true
			break
		}
	}

	if !hasContent {
		return ErrEmptyContent
	}

	return nil
}

// ExtractSections splits content into sections based on markdown headers
func (bp *BaseParser) ExtractSections(content []string, headerLevel int) []Section {
	var sections []Section
	var currentSection Section
	headerPrefix := strings.Repeat("#", headerLevel) + " "

	for i, line := range content {
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, headerPrefix) && !strings.HasPrefix(trimmedLine, headerPrefix+"#") {
			if currentSection.Title != "" || len(currentSection.Content) > 0 {
				sections = append(sections, currentSection)
			}

			currentSection = Section{
				Title:     strings.TrimSpace(trimmedLine[len(headerPrefix):]),
				StartLine: i,
				Content:   []string{},
			}
		} else {
			currentSection.Content = append(currentSection.Content, line)
		}
	}

	if currentSection.Title != "" || len(currentSection.Content) > 0 {
		sections = append(sections, currentSection)
	}

	return sections
}

// NewDocumentBuilder creates a DocumentBuilder with base document data
func (bp *BaseParser) NewDocumentBuilder(title, content string, context *ParsingContext) (*DocumentBuilder, error) {
	return NewDocumentBuilder(title, content, context)
}

// ===== TEMPLATE METHOD IMPLEMENTATION =====

// Parse implements the Template Method pattern - defines the parsing algorithm skeleton
func (bp *BaseParser) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
	// Step 1: Pre-processing validation (common across all parsers)
	if err := bp.Validate(content); err != nil {
		return nil, err
	}

	// Step 2: Extract sections using language-aware delimiter
	sections, err := bp.extractSectionsWithLanguage(content)
	if err != nil {
		return nil, fmt.Errorf("failed to extract sections: %w", err)
	}

	// Step 3: Parse each section (hook method - implemented by concrete parsers)
	var entities []domain.ParsedEntity
	for _, section := range sections {
		if !section.HasContent() {
			continue
		}

		entity, err := bp.parseSection(section, context)
		if err != nil {
			bp.LogParsingProgress("Section parsing failed for %s: %v", section.Title, err)
			continue // or handle based on error strategy
		}

		// Step 4: Post-process entity (hook method - can be overridden)
		if processed := bp.postProcessEntity(entity); processed != nil {
			entities = append(entities, processed)
		}
	}

	// Step 5: Final validation and cleanup (common)
	return bp.finalizeEntities(entities)
}

// ===== HOOK METHODS (to be implemented by concrete parsers) =====

// parseSection is the main hook method that concrete parsers must implement
func (bp *BaseParser) parseSection(section Section, context *ParsingContext) (domain.ParsedEntity, error) {
	return nil, fmt.Errorf("parseSection must be implemented by concrete parser")
}

// postProcessEntity is an optional hook method for entity post-processing
func (bp *BaseParser) postProcessEntity(entity domain.ParsedEntity) domain.ParsedEntity {
	// Default implementation - can be overridden by concrete parsers
	return entity
}

// ===== TEMPLATE METHOD HELPER METHODS (common implementation) =====

// extractSectionsWithLanguage extracts sections using language-specific configuration
func (bp *BaseParser) extractSectionsWithLanguage(content []string) ([]Section, error) {
	headerLevel := 2 // default header level

	// Use language-specific section delimiter if available
	if bp.config != nil && bp.config.SectionDelimiter != "" {
		// Count the # characters in the delimiter to determine header level
		headerLevel = strings.Count(bp.config.SectionDelimiter, "#")
		if headerLevel == 0 {
			headerLevel = 2 // fallback
		}
	}

	return bp.ExtractSections(content, headerLevel), nil
}

// finalizeEntities performs final processing on parsed entities
func (bp *BaseParser) finalizeEntities(entities []domain.ParsedEntity) ([]domain.ParsedEntity, error) {
	// Remove nil entities
	var cleaned []domain.ParsedEntity
	for _, entity := range entities {
		if entity != nil {
			cleaned = append(cleaned, entity)
		}
	}

	bp.LogParsingProgress("Successfully parsed %d entities of type %s", len(cleaned), bp.contentType)
	return cleaned, nil
}

// ===== LANGUAGE-AWARE HELPER METHODS =====

// getFieldName resolves language-specific field names using configuration
func (bp *BaseParser) getFieldName(englishField string) string {
	if bp.config != nil && bp.config.FieldMappings != nil {
		if mapped, exists := bp.config.FieldMappings[englishField]; exists {
			return mapped
		}
	}
	return englishField // fallback to English field name
}

// matchPattern matches content against language-specific patterns
func (bp *BaseParser) matchPattern(patternName, content string) []string {
	if bp.config != nil && bp.config.Patterns != nil {
		if pattern, exists := bp.config.Patterns[patternName]; exists {
			return pattern.FindStringSubmatch(content)
		}
	}
	return nil
}

// extractFieldFromLines extracts a field value from content lines
func (bp *BaseParser) extractFieldFromLines(lines []string, fieldName string) string {
	fieldKey := bp.getFieldName(fieldName)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), strings.ToLower("**"+fieldKey+":")) {
			// Extract value after the field name
			parts := strings.SplitN(line, ":", 2)
			if len(parts) > 1 {
				return strings.TrimSpace(strings.Trim(parts[1], "* "))
			}
		}
	}
	return ""
}

// CleanContent removes empty lines and trims whitespace
func (bp *BaseParser) CleanContent(lines []string) []string {
	var cleaned []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}

	return cleaned
}

// JoinContent joins lines into a single string with proper spacing
func (bp *BaseParser) JoinContent(lines []string) string {
	cleaned := bp.CleanContent(lines)
	return strings.Join(cleaned, "\n\n")
}

// Section represents a parsed section of content
type Section struct {
	Title     string
	Content   []string
	StartLine int
}

// GetContentAsString returns the section content as a single string
func (s *Section) GetContentAsString() string {
	return strings.Join(s.Content, "\n")
}

// GetCleanContent returns cleaned content lines
func (s *Section) GetCleanContent() []string {
	var cleaned []string
	for _, line := range s.Content {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	return cleaned
}

// HasContent returns true if the section has non-empty content
func (s *Section) HasContent() bool {
	for _, line := range s.Content {
		if strings.TrimSpace(line) != "" {
			return true
		}
	}
	return false
}

func (bp *BaseParser) LogParsingProgress(message string, args ...interface{}) {
	bp.logger.Info(message, args...)
}

// Validate checks if the section is valid for parsing
func (s *Section) Validate() error {
	if s.Title == "" {
		return ErrMissingSectionTitle
	}
	if !s.HasContent() {
		return ErrEmptySectionContent
	}
	return nil
}

// GetMetadata retrieves metadata for the section (extensible for future use)
func (s *Section) GetMetadata(key string) (string, bool) {
	// Add metadata support if needed in the future
	return "", false
}
