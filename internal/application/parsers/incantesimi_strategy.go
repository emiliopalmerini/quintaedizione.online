package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

type IncantesimiStrategy struct{}

func NewIncantesimiStrategy() *IncantesimiStrategy {
	return &IncantesimiStrategy{}
}

func (s *IncantesimiStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
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

		// Check for new spell section (H2)
		if strings.HasPrefix(line, "## ") {
			// Process previous section if exists
			if inSection && len(currentSection) > 0 {
				spell, err := s.parseSpellSection(currentSection)
				if err != nil {
					context.Logger.Error(fmt.Sprintf("Failed to parse spell section: %v", err))
					currentSection = []string{}
					inSection = false
					continue
				}
				entities = append(entities, spell)
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
		spell, err := s.parseSpellSection(currentSection)
		if err != nil {
			context.Logger.Error(fmt.Sprintf("Failed to parse last spell section: %v", err))
		} else {
			entities = append(entities, spell)
		}
	}

	if len(entities) == 0 {
		return nil, ErrEmptyContent
	}

	return entities, nil
}

func (s *IncantesimiStrategy) parseSpellSection(section []string) (*domain.Incantesimo, error) {
	if len(section) == 0 {
		return nil, ErrEmptySectionContent
	}

	// Extract name from header
	header := section[0]
	if !strings.HasPrefix(header, "## ") {
		return nil, ErrMissingSectionTitle
	}
	nome := strings.TrimSpace(strings.TrimPrefix(header, "## "))
	// Remove formatting: first remove trailing period, then ** markers
	nome = strings.TrimSuffix(nome, ".")
	nome = strings.Trim(nome, "*")

	// Find metadata line (italics line with level/school/classes)
	var metadataLine string
	var startIndex = 1
	
	for i := 1; i < len(section); i++ {
		line := strings.TrimSpace(section[i])
		// Match metadata line: starts with single * and ends with *.
		// Avoid matching bold text like ***Text***
		if strings.HasPrefix(line, "*") && strings.HasSuffix(line, "*.") && !strings.HasPrefix(line, "**") {
			metadataLine = line
			startIndex = i + 1
			break
		}
	}

	if metadataLine == "" {
		return nil, fmt.Errorf("missing spell metadata line")
	}

	// Parse metadata
	livello, scuola, classi, err := s.parseMetadata(metadataLine)
	if err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Parse fields and build contenuto
	fields := make(map[string]string)
	metadataLines := strings.Builder{}
	descriptionLines := strings.Builder{}
	contentStartIndex := startIndex

	for i := startIndex; i < len(section); i++ {
		line := section[i]

		// Parse field format: **Field:** value
		if strings.HasPrefix(line, "**") && strings.Contains(line, ":**") {
			parts := strings.SplitN(line, ":**", 2)
			if len(parts) == 2 {
				fieldName := strings.TrimSpace(strings.Trim(parts[0], "*"))
				fieldValue := strings.TrimSpace(parts[1])
				fields[fieldName] = fieldValue

				// Add to metadata section
				metadataLines.WriteString(line + "\n")

				// Track when we've found Durata - description starts after it
				if fieldName == "Durata" {
					contentStartIndex = i + 1
				}
			}
		}
	}

	// Build description starting after Durata
	for i := contentStartIndex; i < len(section); i++ {
		line := section[i]
		// Skip empty lines at the start of description
		if descriptionLines.Len() == 0 && strings.TrimSpace(line) == "" {
			continue
		}
		descriptionLines.WriteString(line + "\n")
	}

	// Combine metadata and description for contenuto
	contenuto := strings.Builder{}
	contenuto.WriteString(strings.TrimSpace(metadataLines.String()))
	if descriptionLines.Len() > 0 {
		contenuto.WriteString("\n\n")
		contenuto.WriteString(strings.TrimSpace(descriptionLines.String()))
	}

	// Parse lancio fields (handle both Italian and English field names)
	normalizedFields := s.normalizeFieldNames(fields)
	lancio, err := s.parseLancio(normalizedFields)
	if err != nil {
		return nil, fmt.Errorf("failed to parse lancio: %w", err)
	}

	spell := domain.NewIncantesimo(
		nome,
		livello,
		scuola,
		classi,
		lancio,
		contenuto.String(),
	)

	return spell, nil
}

func (s *IncantesimiStrategy) parseMetadata(line string) (uint8, string, []string, error) {
	// Remove asterisks
	line = strings.Trim(line, "*")
	
	// Parse patterns:
	// "Livello 2 Invocazione (Mago)"
	// "Trucchetto di Invocazione (Stregone, Mago)"
	
	var livello uint8
	var scuola string
	var classi []string
	
	// Find parentheses with classes
	parenStart := strings.LastIndex(line, "(")
	parenEnd := strings.LastIndex(line, ")")
	
	if parenStart == -1 || parenEnd == -1 || parenEnd < parenStart {
		return 0, "", nil, fmt.Errorf("invalid metadata format: %s", line)
	}
	
	// Extract classes
	classiStr := strings.TrimSpace(line[parenStart+1 : parenEnd])
	classiParts := strings.Split(classiStr, ",")
	for _, classe := range classiParts {
		classe = strings.TrimSpace(classe)
		if classe != "" {
			classi = append(classi, classe)
		}
	}
	
	// Extract level and school
	levelSchoolStr := strings.TrimSpace(line[:parenStart])
	
	if strings.HasPrefix(levelSchoolStr, "Trucchetto") {
		livello = 0
		// Extract school from "Trucchetto di Scuola"
		parts := strings.Split(levelSchoolStr, " di ")
		if len(parts) == 2 {
			scuola = strings.TrimSpace(parts[1])
		} else {
			return 0, "", nil, fmt.Errorf("invalid cantrip format: %s", levelSchoolStr)
		}
	} else if strings.HasPrefix(levelSchoolStr, "Livello") {
		// Extract level and school from "Livello X Scuola"
		re := regexp.MustCompile(`Livello (\d+) (.+)`)
		matches := re.FindStringSubmatch(levelSchoolStr)
		if len(matches) != 3 {
			return 0, "", nil, fmt.Errorf("invalid spell level format: %s", levelSchoolStr)
		}
		
		levelInt, err := strconv.Atoi(matches[1])
		if err != nil {
			return 0, "", nil, fmt.Errorf("invalid level number: %s", matches[1])
		}
		livello = uint8(levelInt)
		scuola = strings.TrimSpace(matches[2])
	} else if strings.Contains(levelSchoolStr, "° livello") {
		// Handle Italian format: "Scuola di X° livello"
		re := regexp.MustCompile(`(.+?)\s+di\s+(\d+)°\s+livello`)
		matches := re.FindStringSubmatch(levelSchoolStr)
		if len(matches) == 3 {
			scuola = strings.TrimSpace(matches[1])
			levelInt, err := strconv.Atoi(matches[2])
			if err != nil {
				return 0, "", nil, fmt.Errorf("invalid level number: %s", matches[2])
			}
			livello = uint8(levelInt)
		} else {
			return 0, "", nil, fmt.Errorf("invalid Italian spell level format: %s", levelSchoolStr)
		}
	} else {
		return 0, "", nil, fmt.Errorf("unknown metadata format: %s", levelSchoolStr)
	}
	
	return livello, scuola, classi, nil
}

// normalizeFieldNames is a placeholder - no normalization needed for Italian-only content
func (s *IncantesimiStrategy) normalizeFieldNames(fields map[string]string) map[string]string {
	return fields
}

func (s *IncantesimiStrategy) parseLancio(fields map[string]string) (domain.Lancio, error) {
	tempo, err := s.parseTempoLancio(fields["Tempo di Lancio"])
	if err != nil {
		return domain.Lancio{}, fmt.Errorf("failed to parse tempo di lancio: %w", err)
	}
	
	gittata, err := s.parseGittata(fields["Gittata"])
	if err != nil {
		return domain.Lancio{}, fmt.Errorf("failed to parse gittata: %w", err)
	}
	
	componenti, err := s.parseComponenti(fields["Componenti"])
	if err != nil {
		return domain.Lancio{}, fmt.Errorf("failed to parse componenti: %w", err)
	}
	
	durata, err := s.parseDurata(fields["Durata"])
	if err != nil {
		return domain.Lancio{}, fmt.Errorf("failed to parse durata: %w", err)
	}
	
	return domain.Lancio{
		Tempo:      tempo,
		Gittata:    gittata,
		Componenti: componenti,
		Durata:     durata,
	}, nil
}

func (s *IncantesimiStrategy) parseTempoLancio(value string) (domain.TempoLancio, error) {
	if value == "" {
		return domain.TempoLancio{}, fmt.Errorf("empty tempo di lancio")
	}
	
	value = strings.TrimSpace(value)
	
	// Handle special cases
	if value == "Azione" {
		return domain.TempoLancio{Tipo: domain.TempoAzione, Valore: 0}, nil
	}
	if value == "Azione bonus" {
		return domain.TempoLancio{Tipo: domain.TempoAzioneBonus, Valore: 0}, nil
	}
	if strings.Contains(value, "minuto") {
		// Extract number of minutes
		re := regexp.MustCompile(`(\d+)\s*minut`)
		matches := re.FindStringSubmatch(value)
		if len(matches) >= 2 {
			minutes, err := strconv.Atoi(matches[1])
			if err == nil {
				nota := ""
				if strings.Contains(value, "Rituale") {
					nota = "Rituale"
				}
				return domain.TempoLancio{Tipo: domain.TempoMinuti, Valore: minutes, Nota: nota}, nil
			}
		}
		return domain.TempoLancio{Tipo: domain.TempoMinuti, Valore: 1}, nil
	}
	
	// Default to special with the full text as nota
	return domain.TempoLancio{Tipo: domain.TempoSpeciale, Valore: 0, Nota: value}, nil
}

func (s *IncantesimiStrategy) parseGittata(value string) (domain.GittataIncantesimo, error) {
	if value == "" {
		return domain.GittataIncantesimo{}, fmt.Errorf("empty gittata")
	}
	
	value = strings.TrimSpace(value)
	
	if value == "Contatto" || value == "Tocco" {
		return domain.GittataIncantesimo{Tipo: domain.GittataContatto}, nil
	}
	if value == "Se" {
		return domain.GittataIncantesimo{Tipo: domain.GittataSe}, nil
	}
	
	// Parse distance like "27 m", "18 m", "9 m"
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*([a-zA-Z]+)`)
	matches := re.FindStringSubmatch(value)
	if len(matches) >= 3 {
		distValue, err := strconv.ParseFloat(matches[1], 64)
		if err == nil {
			unita := matches[2]
			distanza := &domain.Distanza{
				Valore: distValue,
				Unita:  unita,
			}
			return domain.GittataIncantesimo{
				Tipo:     domain.GittataDistanza,
				Distanza: distanza,
			}, nil
		}
	}
	
	// Default to special
	return domain.GittataIncantesimo{Tipo: domain.GittataSpeciale, Nota: value}, nil
}

func (s *IncantesimiStrategy) parseComponenti(value string) (domain.Componenti, error) {
	if value == "" {
		return domain.Componenti{}, fmt.Errorf("empty componenti")
	}
	
	comp := domain.Componenti{}
	
	// Check for V, S, M components
	if strings.Contains(value, "V") {
		comp.V = true
	}
	if strings.Contains(value, "S") {
		comp.S = true
	}
	if strings.Contains(value, "M") {
		comp.M = true
		
		// Extract material components from parentheses
		parenStart := strings.Index(value, "(")
		parenEnd := strings.LastIndex(value, ")")
		if parenStart != -1 && parenEnd != -1 && parenEnd > parenStart {
			comp.Materiali = strings.TrimSpace(value[parenStart+1 : parenEnd])
		}
	}
	
	return comp, nil
}

func (s *IncantesimiStrategy) parseDurata(value string) (domain.Durata, error) {
	if value == "" {
		return domain.Durata{}, fmt.Errorf("empty durata")
	}
	
	value = strings.TrimSpace(value)
	
	if value == "Istantanea" {
		return domain.Durata{Tipo: domain.DurataIstantanea}, nil
	}
	
	// Check for concentration
	isConcentration := strings.Contains(strings.ToLower(value), "concentrazione")
	
	// Parse time durations like "8 ore", "1 minuto", "fino a 10 minuti"
	re := regexp.MustCompile(`(\d+)\s*(minut|or)`)
	matches := re.FindStringSubmatch(value)
	if len(matches) >= 3 {
		timeValue, err := strconv.Atoi(matches[1])
		if err == nil {
			unit := matches[2]
			if strings.HasPrefix(unit, "minut") {
				return domain.Durata{
					Tipo:           domain.DurataTempo,
					Minuti:         timeValue,
					Concentrazione: isConcentration,
				}, nil
			} else if strings.HasPrefix(unit, "or") {
				return domain.Durata{
					Tipo:           domain.DurataTempo,
					Minuti:         timeValue * 60, // Convert hours to minutes
					Concentrazione: isConcentration,
				}, nil
			}
		}
	}
	
	// Default to special
	return domain.Durata{Tipo: domain.DurataSpeciale, Nota: value}, nil
}

func (s *IncantesimiStrategy) ContentType() ContentType {
	return ContentTypeIncantesimi
}

func (s *IncantesimiStrategy) Name() string {
	return "Incantesimi Strategy"
}

func (s *IncantesimiStrategy) Description() string {
	return "Parses Italian D&D 5e spells (incantesimi) from markdown content"
}

func (s *IncantesimiStrategy) Validate(content []string) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}
	return nil
}