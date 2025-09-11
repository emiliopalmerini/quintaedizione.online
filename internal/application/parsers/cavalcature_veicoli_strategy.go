package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

type CavalcatureVeicoliStrategy struct{}

func NewCavalcatureVeicoliStrategy() *CavalcatureVeicoliStrategy {
	return &CavalcatureVeicoliStrategy{}
}

func (s *CavalcatureVeicoliStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
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

		// Check for new mount/vehicle section (H2)
		if strings.HasPrefix(line, "## ") {
			// Process previous section if exists
			if inSection && len(currentSection) > 0 {
				mountVehicle, err := s.parseMountVehicleSection(currentSection)
				if err != nil {
					context.Logger.Error(fmt.Sprintf("Failed to parse mount/vehicle section: %v", err))
					currentSection = []string{}
					inSection = false
					continue
				}
				entities = append(entities, mountVehicle)
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
		mountVehicle, err := s.parseMountVehicleSection(currentSection)
		if err != nil {
			context.Logger.Error(fmt.Sprintf("Failed to parse last mount/vehicle section: %v", err))
		} else {
			entities = append(entities, mountVehicle)
		}
	}

	if len(entities) == 0 {
		return nil, ErrEmptyContent
	}

	return entities, nil
}

func (s *CavalcatureVeicoliStrategy) parseMountVehicleSection(section []string) (*domain.CavalcaturaVeicolo, error) {
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
	var descrizione []string
	
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
		} else if line != "" && !strings.HasPrefix(line, "**") {
			// This is description text
			descrizione = append(descrizione, line)
		}
	}

	// Parse fields
	tipo := s.parseTipo(fields["Tipo"])
	costo := s.parseCosto(fields["Costo"])
	velocita := s.parseVelocita(fields["Velocità"])
	capacitaCarico := s.parseCapacitaCarico(fields["Capacità Carico"])
	equipaggio := s.parseIntPtr(fields["Equipaggio"])
	passeggeri := s.parseIntPtr(fields["Passeggeri"])
	ca := s.parseIntPtr(fields["CA"])
	pf := s.parseIntPtr(fields["PF"])
	sogliaDanni := s.parseIntPtr(fields["Soglia Danni"])
	
	// Combine description
	descrizioneText := strings.Join(descrizione, " ")

	mountVehicle := domain.NewCavalcaturaVeicolo(
		nome,
		tipo,
		costo,
		velocita,
		capacitaCarico,
		equipaggio,
		passeggeri,
		ca,
		pf,
		sogliaDanni,
		descrizioneText,
		strings.TrimSpace(contenuto.String()),
	)

	return mountVehicle, nil
}

func (s *CavalcatureVeicoliStrategy) parseTipo(value string) domain.TipoCavalcaturaVeicolo {
	value = strings.ToLower(strings.TrimSpace(value))
	
	switch value {
	case "cavalcatura":
		return domain.TipoCavalcatura
	case "nave":
		return domain.TipoNave
	case "veicolo":
		return domain.TipoVeicolo
	default:
		return domain.TipoAltro
	}
}

func (s *CavalcatureVeicoliStrategy) parseCosto(value string) domain.Costo {
	if value == "" || value == "—" {
		return domain.NewCosto(0, domain.ValutaOro)
	}

	// Parse format like "75 mo", "400 mo", etc.
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*([a-z]{2})`)
	matches := re.FindStringSubmatch(value)
	if len(matches) != 3 {
		return domain.NewCosto(0, domain.ValutaOro)
	}

	valore, err := strconv.Atoi(matches[1])
	if err != nil {
		return domain.NewCosto(0, domain.ValutaOro)
	}

	var valuta domain.Valuta
	switch matches[2] {
	case "mr":
		valuta = domain.ValutaRame
	case "ma":
		valuta = domain.ValutaArgento
	case "me":
		valuta = domain.ValutaElettro
	case "mo":
		valuta = domain.ValutaOro
	case "mp":
		valuta = domain.ValutaPlatino
	default:
		valuta = domain.ValutaOro
	}

	return domain.NewCosto(valore, valuta)
}

func (s *CavalcatureVeicoliStrategy) parseVelocita(value string) domain.VelocitaVeicolo {
	if value == "" || value == "—" {
		return domain.VelocitaVeicolo{Valore: nil, Unita: domain.UnitaMetri}
	}

	// Parse formats like "12 m", "24 km/h", etc.
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*([a-zA-Z/]+)`)
	matches := re.FindStringSubmatch(value)
	if len(matches) != 3 {
		return domain.VelocitaVeicolo{Valore: nil, Unita: domain.UnitaMetri}
	}

	valore, err := strconv.Atoi(matches[1])
	if err != nil {
		return domain.VelocitaVeicolo{Valore: nil, Unita: domain.UnitaMetri}
	}

	var unita domain.UnitaVelocita
	unitaStr := strings.ToLower(matches[2])
	switch unitaStr {
	case "m", "metri":
		unita = domain.UnitaMetri
	case "km/h", "kmh":
		unita = domain.UnitaKmOra
	case "nodi":
		unita = domain.UnitaNodi
	default:
		unita = domain.UnitaMetri
	}

	return domain.VelocitaVeicolo{Valore: &valore, Unita: unita}
}

func (s *CavalcatureVeicoliStrategy) parseCapacitaCarico(value string) domain.Peso {
	if value == "" || value == "—" {
		return domain.NewPeso(0, domain.UnitaKg)
	}

	// Parse format like "225 kg", "1000 kg", etc.
	value = strings.ReplaceAll(value, ",", ".")
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*kg`)
	matches := re.FindStringSubmatch(value)
	if len(matches) != 2 {
		return domain.NewPeso(0, domain.UnitaKg)
	}

	valore, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return domain.NewPeso(0, domain.UnitaKg)
	}

	return domain.NewPeso(valore, domain.UnitaKg)
}

func (s *CavalcatureVeicoliStrategy) parseIntPtr(value string) *int {
	if value == "" || value == "—" {
		return nil
	}

	val, err := strconv.Atoi(value)
	if err != nil {
		return nil
	}

	return &val
}

func (s *CavalcatureVeicoliStrategy) ContentType() ContentType {
	return ContentTypeCavalcatureVeicoli
}

func (s *CavalcatureVeicoliStrategy) Name() string {
	return "Cavalcature Veicoli Strategy"
}

func (s *CavalcatureVeicoliStrategy) Description() string {
	return "Parses Italian D&D 5e mounts and vehicles from markdown content"
}

func (s *CavalcatureVeicoliStrategy) Validate(content []string) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}
	return nil
}