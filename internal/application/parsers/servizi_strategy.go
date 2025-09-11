package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

type ServiziStrategy struct{}

func NewServiziStrategy() *ServiziStrategy {
	return &ServiziStrategy{}
}

func (s *ServiziStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
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

		// Check for new service section (H2)
		if strings.HasPrefix(line, "## ") {
			// Process previous section if exists
			if inSection && len(currentSection) > 0 {
				service, err := s.parseServiceSection(currentSection)
				if err != nil {
					context.Logger.Error(fmt.Sprintf("Failed to parse service section: %v", err))
					currentSection = []string{}
					inSection = false
					continue
				}
				entities = append(entities, service)
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
		service, err := s.parseServiceSection(currentSection)
		if err != nil {
			context.Logger.Error(fmt.Sprintf("Failed to parse last service section: %v", err))
		} else {
			entities = append(entities, service)
		}
	}

	if len(entities) == 0 {
		return nil, ErrEmptyContent
	}

	return entities, nil
}

func (s *ServiziStrategy) parseServiceSection(section []string) (*domain.Servizio, error) {
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

	// Parse cost
	costo := s.parseCosto(fields["Costo"])
	
	// Parse category
	categoria := s.parseCategoria(fields["Categoria"])
	
	// Get description
	descrizione := fields["Descrizione"]

	service := domain.NewServizio(
		nome,
		costo,
		categoria,
		descrizione,
		strings.TrimSpace(contenuto.String()),
	)

	return service, nil
}

func (s *ServiziStrategy) parseCosto(value string) domain.CostoServizio {
	if value == "" || strings.ToLower(value) == "gratuito" {
		return domain.CostoServizio{Valore: 0, Valuta: "gratuito"}
	}

	// Parse formats like "1 AP", "2 MO", "5 mo", etc.
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*([A-Za-z]{1,2})`)
	matches := re.FindStringSubmatch(value)
	if len(matches) != 3 {
		return domain.CostoServizio{Valore: 0, Valuta: "gratuito"}
	}

	valore, err := strconv.Atoi(matches[1])
	if err != nil {
		return domain.CostoServizio{Valore: 0, Valuta: "gratuito"}
	}

	valuta := strings.ToLower(matches[2])
	// Map different currency representations
	switch valuta {
	case "ap":
		valuta = "ma" // AP = Argento
	case "mo":
		valuta = "mo" // MO = Oro
	default:
		// Keep as is for other currencies (mr, me, mp)
	}

	return domain.CostoServizio{Valore: valore, Valuta: valuta}
}

func (s *ServiziStrategy) parseCategoria(value string) domain.CategoriaServizio {
	value = strings.ToLower(strings.TrimSpace(value))
	
	switch value {
	case "servizio", "tenore di vita":
		return domain.CategoriaTenorevita
	case "alloggio":
		return domain.CategoriaAlloggio
	case "trasporto":
		return domain.CategoriaTrasporto
	case "servizio magico", "magia":
		return domain.CategoriaServizioMagico
	default:
		// Default to tenore di vita for services
		return domain.CategoriaTenorevita
	}
}

func (s *ServiziStrategy) ContentType() ContentType {
	return ContentTypeServizi
}

func (s *ServiziStrategy) Name() string {
	return "Servizi Strategy"
}

func (s *ServiziStrategy) Description() string {
	return "Parses Italian D&D 5e services from markdown content"
}

func (s *ServiziStrategy) Validate(content []string) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}
	return nil
}