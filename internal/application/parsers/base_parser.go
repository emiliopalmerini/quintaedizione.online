package parsers

import (
	"strings"
)

// BaseParser provides common functionality for all parsing strategies
type BaseParser struct {
	contentType ContentType
	name        string
	description string
	logger      Logger
}

// NewBaseParser creates a new base parser
func NewBaseParser(contentType ContentType, name, description string) *BaseParser {
	return &BaseParser{
		contentType: contentType,
		name:        name,
		description: description,
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
