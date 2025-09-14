package parsers

import (
	"fmt"
	"slices"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// ValidationError represents a parsing validation error with context
type ValidationError struct {
	EntityName string
	Field      string
	Line       int
	Message    string
}

func (e ValidationError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("validation error in entity '%s', field '%s' at line %d: %s", 
			e.EntityName, e.Field, e.Line, e.Message)
	}
	return fmt.Sprintf("validation error in entity '%s', field '%s': %s", 
		e.EntityName, e.Field, e.Message)
}

// Validator provides validation functionality for parsed entities
type Validator struct{}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateContent validates content structure before parsing
func (v *Validator) ValidateContent(content []string, contentType ContentType) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}

	hasTitle := false
	hasSections := false

	for _, line := range content {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			hasTitle = true
		}
		if strings.HasPrefix(line, "## ") {
			hasSections = true
		}
	}

	if !hasTitle {
		return fmt.Errorf("content missing main title (H1)")
	}

	if !hasSections {
		return fmt.Errorf("content missing entity sections (H2)")
	}

	return nil
}

// ValidateArma validates an Arma domain object
func (v *Validator) ValidateArma(arma *domain.Arma) error {
	if arma == nil {
		return fmt.Errorf("arma is nil")
	}

	if arma.Nome == "" {
		return ValidationError{
			EntityName: "Arma",
			Field:      "Nome",
			Message:    "name cannot be empty",
		}
	}

	if arma.Slug == "" {
		return ValidationError{
			EntityName: arma.Nome,
			Field:      "Slug",
			Message:    "slug cannot be empty",
		}
	}

	if arma.Danno == "" {
		return ValidationError{
			EntityName: arma.Nome,
			Field:      "Danno",
			Message:    "damage cannot be empty",
		}
	}

	// Validate category-property consistency
	if err := v.validateWeaponCategoryProperties(arma); err != nil {
		return err
	}

	return nil
}

// ValidateArmatura validates an Armatura domain object
func (v *Validator) ValidateArmatura(armatura *domain.Armatura) error {
	if armatura == nil {
		return fmt.Errorf("armatura is nil")
	}

	if armatura.Nome == "" {
		return ValidationError{
			EntityName: "Armatura",
			Field:      "Nome",
			Message:    "name cannot be empty",
		}
	}

	if armatura.Slug == "" {
		return ValidationError{
			EntityName: armatura.Nome,
			Field:      "Slug",
			Message:    "slug cannot be empty",
		}
	}

	// Validate AC makes sense for category
	if err := v.validateArmorACCategory(armatura); err != nil {
		return err
	}

	return nil
}

// validateWeaponCategoryProperties checks consistency between weapon category and properties
func (v *Validator) validateWeaponCategoryProperties(arma *domain.Arma) error {
	// Ranged weapons should not have certain melee properties
	if arma.Categoria == domain.CategoriaArmaSempliceDistanza || 
	   arma.Categoria == domain.CategoriaArmaMarzialeDistanza {
		
		invalidProps := []domain.ProprietaArma{
			domain.ProprietaVersatile,
		}
		
		for _, prop := range arma.Proprieta {
			if slices.Contains(invalidProps, prop) {
				return ValidationError{
					EntityName: arma.Nome,
					Field:      "Proprieta",
					Message:    fmt.Sprintf("ranged weapon cannot have property: %s", prop),
				}
			}
		}
	}

	return nil
}

// validateArmorACCategory checks if AC value makes sense for armor category
func (v *Validator) validateArmorACCategory(armatura *domain.Armatura) error {
	// ClasseArmatura is a struct, not a pointer, so it can't be nil

	ac := armatura.ClasseArmatura.Base
	
	switch armatura.Categoria {
	case domain.CategoriaArmaturaLeggera:
		if ac < 10 || ac > 13 {
			return ValidationError{
				EntityName: armatura.Nome,
				Field:      "ClasseArmatura",
				Message:    fmt.Sprintf("light armor AC %d outside expected range 10-13", ac),
			}
		}
	case domain.CategoriaArmaturaMedia:
		if ac < 12 || ac > 15 {
			return ValidationError{
				EntityName: armatura.Nome,
				Field:      "ClasseArmatura",
				Message:    fmt.Sprintf("medium armor AC %d outside expected range 12-15", ac),
			}
		}
	case domain.CategoriaArmaturaPesante:
		if ac < 14 || ac > 18 {
			return ValidationError{
				EntityName: armatura.Nome,
				Field:      "ClasseArmatura",
				Message:    fmt.Sprintf("heavy armor AC %d outside expected range 14-18", ac),
			}
		}
	}

	return nil
}

// ValidateSection validates a markdown section structure
func (v *Validator) ValidateSection(section []string, entityName string) error {
	if len(section) == 0 {
		return ValidationError{
			EntityName: entityName,
			Field:      "section",
			Message:    "section is empty",
		}
	}

	header := section[0]
	if !strings.HasPrefix(header, "## ") {
		return ValidationError{
			EntityName: entityName,
			Field:      "header",
			Line:       1,
			Message:    fmt.Sprintf("invalid header format: %s", header),
		}
	}

	// Check for required field format in remaining lines
	hasFields := false
	for i := 1; i < len(section); i++ {
		line := section[i]
		if strings.HasPrefix(line, "**") && strings.Contains(line, ":**") {
			hasFields = true
			break
		}
	}

	if !hasFields {
		return ValidationError{
			EntityName: entityName,
			Field:      "fields",
			Message:    "section contains no valid field definitions",
		}
	}

	return nil
}

// ValidateRequiredFields checks that required fields are present and non-empty
func (v *Validator) ValidateRequiredFields(fields map[string]string, required []string, entityName string) error {
	var missing []string
	var empty []string

	for _, field := range required {
		value, exists := fields[field]
		if !exists {
			missing = append(missing, field)
		} else if strings.TrimSpace(value) == "" {
			empty = append(empty, field)
		} else if field == "Peso" && value == "—" {
			// "—" is valid for Peso field (indicates no weight)
			continue
		} else if value == "—" {
			empty = append(empty, field)
		}
	}

	if len(missing) > 0 {
		return ValidationError{
			EntityName: entityName,
			Field:      "required_fields",
			Message:    fmt.Sprintf("missing required fields: %v", missing),
		}
	}

	if len(empty) > 0 {
		return ValidationError{
			EntityName: entityName,
			Field:      "required_fields",
			Message:    fmt.Sprintf("empty required fields: %v", empty),
		}
	}

	return nil
}