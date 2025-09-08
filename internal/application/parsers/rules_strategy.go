package parsers

import (
	"fmt"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
	"github.com/google/uuid"
)

// RulesStrategy implements the Strategy pattern for parsing rules using Template Method
type RulesStrategy struct {
	*BaseParser
}

// NewRulesStrategy creates a new rules parsing strategy
func NewRulesStrategy() ParsingStrategy {
	return &RulesStrategy{
		BaseParser: NewBaseParser(
			ContentTypeRules,
			"Rules Parser",
			"Parses D&D 5e rules from Italian SRD markdown content",
		),
	}
}

// parseSection implements the Template Method hook for rule-specific parsing
func (r *RulesStrategy) parseSection(section Section, context *ParsingContext) (domain.ParsedEntity, error) {
	return r.parseRuleSection(section)
}

func (r *RulesStrategy) parseRuleSection(section Section) (*domain.Regola, error) {
	if section.Title == "" {
		return nil, fmt.Errorf("rule section has no title")
	}

	content := section.GetCleanContent()
	if len(content) == 0 {
		return nil, fmt.Errorf("rule section has no content")
	}

	ruleContent := strings.Join(content, "\n")

	rule := domain.NewRegola(
		uuid.New(),
		section.Title,
		ruleContent,
	)

	return rule, nil
}

// postProcessEntity adds language-aware post-processing for rules
func (r *RulesStrategy) postProcessEntity(entity domain.ParsedEntity) domain.ParsedEntity {
	rule, ok := entity.(*domain.Regola)
	if !ok {
		return entity
	}

	// Add any rule-specific post-processing here
	// For example, generate slug, validate fields, etc.
	return rule
}
