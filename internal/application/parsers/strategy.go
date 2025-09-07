package parsers

import "github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"

// ParsingStrategy defines the interface for all content parsing strategies
type ParsingStrategy interface {
	Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error)
	ContentType() ContentType
	Name() string
	Description() string
	Validate(content []string) error
}

// TemplateParsingStrategy provides template method support for parsing strategies
type TemplateParsingStrategy interface {
	ParsingStrategy
	ParseSection(section Section, context *ParsingContext) (domain.ParsedEntity, error)
}
