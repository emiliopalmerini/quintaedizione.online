package parsers

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

type ClassiStrategy struct{}

func NewClassiStrategy() *ClassiStrategy {
	return &ClassiStrategy{}
}

func (s *ClassiStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
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

		// Check for new classe section (H2)
		if strings.HasPrefix(line, "## ") {
			// Process previous section if exists
			if inSection && len(currentSection) > 0 {
				classe, err := s.parseClasseSection(currentSection)
				if err != nil {
					context.Logger.Error(fmt.Sprintf("Failed to parse classe section: %v", err))
					continue
				}
				entities = append(entities, classe)
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
		classe, err := s.parseClasseSection(currentSection)
		if err != nil {
			context.Logger.Error(fmt.Sprintf("Failed to parse last classe section: %v", err))
		} else {
			entities = append(entities, classe)
		}
	}

	if len(entities) == 0 {
		return nil, ErrEmptyContent
	}

	return entities, nil
}

func (s *ClassiStrategy) parseClasseSection(section []string) (*domain.Classe, error) {
	if len(section) == 0 {
		return nil, ErrEmptySectionContent
	}

	// Extract name from header
	header := section[0]
	if !strings.HasPrefix(header, "## ") {
		return nil, ErrMissingSectionTitle
	}
	nome := strings.TrimSpace(strings.TrimPrefix(header, "## "))

	// Build content for audit
	contenuto := strings.Join(section, "\n")

	// Parse traits table
	traitTable, err := s.parseTraitTable(section)
	if err != nil {
		return nil, fmt.Errorf("failed to parse trait table: %w", err)
	}

	// Parse progression table
	progressionTable := s.parseProgressionTable(section)

	// Parse privileges
	privileges := s.parsePrivileges(section)

	// Parse subclasses (if any in the section)
	subclasses := s.parseSubclasses(section)

	// Extract basic fields from trait table
	dadoVita := s.parseDadoVita(traitTable["Dado Punti Ferita"])
	caratteristichePrimarie := s.parseCaratteristichePrimarie(traitTable["Caratteristica primaria"])
	salvezzeCompetenze := s.parseSalvezzeCompetenze(traitTable["Tiri salvezza competenti"])
	abilitaCompetenzeOpzioni := s.parseAbilitaCompetenze(traitTable["Abilità competenti"])
	armiCompetenze := s.parseArmiCompetenze(traitTable["Armi competenti"])
	armatureCompetenze := s.parseArmatureCompetenze(traitTable["Armature addestramento"])
	equipaggiamentoOpzioni := s.parseEquipaggiamentoInizialeOpzioni(traitTable["Equipaggiamento iniziale"])

	// Create default values for complex structures
	multiclasse := domain.Multiclasse{
		Prerequisiti:    []string{},
		TrattiAcquisiti: []string{},
		Note:            "",
	}

	progressioni := domain.Progressioni{
		MaestriaArmi:          make(domain.ProgressioneLivelli),
		AttacchiExtra:         make(domain.ProgressioneLivelli),
		Risorse:               []domain.RisorsaClasse{},
		AumentiCaratteristica: []int{4, 8, 12, 16, 19},
		DonoEpico:             19,
	}

	// Fill progression data if available
	if progressionTable != nil {
		s.fillProgressionData(&progressioni, progressionTable)
	}

	magia := domain.Magia{
		HaIncantesimi:    false,
		ListaRiferimento: make(domain.ListaIncantesimi),
		Preparazione:     domain.PreparazioneNone,
		Focus:            "",
		Trucchetti:       make(domain.ListaIncantesimi),
		Incantesimi:      make(domain.ListaIncantesimi),
	}

	raccomandazioni := domain.Raccomandazioni{
		TruccanettiCons:         []domain.IncantesimoSlug{},
		IncantesimiInizialiCons: []domain.IncantesimoSlug{},
		EquipInizialecons:       []domain.EquipaggiamentoSlug{},
		TalentiCons:             []domain.TalentoSlug{},
		DonoEpicoCons:           []domain.TalentoSlug{},
	}

	classe := domain.NewClasse(
		nome,
		"", // sottotitolo - not present in standard format
		"", // markdown - not used here
		dadoVita,
		caratteristichePrimarie,
		salvezzeCompetenze,
		abilitaCompetenzeOpzioni,
		armiCompetenze,
		armatureCompetenze,
		[]domain.StrumentoSlug{}, // strumenti - not in basic table
		equipaggiamentoOpzioni,
		multiclasse,
		progressioni,
		magia,
		privileges,
		subclasses,
		make(domain.ListaIncantesimi), // listeIncantesimi - parameter not used in constructor
		raccomandazioni,
		contenuto,
	)

	return classe, nil
}

func (s *ClassiStrategy) parseTraitTable(section []string) (map[string]string, error) {
	traits := make(map[string]string)
	inTable := false

	for i, line := range section {
		// Look for trait table markers
		if strings.Contains(line, "Tratti base del") {
			inTable = true
			continue
		}

		if inTable && strings.HasPrefix(line, "|") {
			// Skip table separators
			if strings.Contains(line, "---") {
				continue
			}

			// Parse table row
			parts := strings.Split(line, "|")
			if len(parts) >= 3 {
				key := strings.TrimSpace(parts[1])
				value := strings.TrimSpace(parts[2])
				if key != "" && value != "" {
					traits[key] = value
				}
			}
		}

		// Stop at next section
		if inTable && (strings.HasPrefix(line, "### ") || strings.HasPrefix(line, "## ")) && i > 0 {
			break
		}
	}

	if len(traits) == 0 {
		return nil, fmt.Errorf("no traits found in section")
	}

	return traits, nil
}

func (s *ClassiStrategy) parseProgressionTable(section []string) []map[string]string {
	var progressionTable []map[string]string
	var headers []string
	inTable := false

	for _, line := range section {
		// Look for progression table
		if strings.Contains(line, "Privilegi del") && strings.Contains(line, "Tabella:") {
			inTable = true
			continue
		}

		if inTable && strings.HasPrefix(line, "|") {
			parts := strings.Split(line, "|")

			// Parse headers
			if len(headers) == 0 && !strings.Contains(line, "---") {
				for _, part := range parts {
					header := strings.TrimSpace(part)
					if header != "" {
						headers = append(headers, header)
					}
				}
				continue
			}

			// Skip separator
			if strings.Contains(line, "---") {
				continue
			}

			// Parse data rows
			if len(headers) > 0 {
				row := make(map[string]string)
				for i, part := range parts {
					value := strings.TrimSpace(part)
					if i-1 < len(headers) && i > 0 && value != "" {
						row[headers[i-1]] = value
					}
				}
				if len(row) > 0 {
					progressionTable = append(progressionTable, row)
				}
			}
		}

		// Stop at next major section
		if inTable && strings.HasPrefix(line, "#### ") {
			break
		}
	}

	return progressionTable
}

func (s *ClassiStrategy) parsePrivileges(section []string) []domain.Privilegio {
	var privileges []domain.Privilegio

	for i, line := range section {
		// Look for privilege headers (#### level: name)
		if strings.HasPrefix(line, "#### ") {
			privilegeText := strings.TrimPrefix(line, "#### ")

			// Extract level and name
			levelMatch := regexp.MustCompile(`^(\d+)° livello: (.+)$`).FindStringSubmatch(privilegeText)
			if len(levelMatch) == 3 {
				level, err := strconv.Atoi(levelMatch[1])
				if err != nil {
					continue
				}

				name := levelMatch[2]

				// Collect description until next section
				var description strings.Builder
				for j := i + 1; j < len(section); j++ {
					nextLine := section[j]
					if strings.HasPrefix(nextLine, "#### ") || strings.HasPrefix(nextLine, "### ") || strings.HasPrefix(nextLine, "## ") {
						break
					}
					description.WriteString(nextLine + "\n")
				}

				privilege := domain.Privilegio{
					Nome:        name,
					Livello:     level,
					Descrizione: strings.TrimSpace(description.String()),
				}
				privileges = append(privileges, privilege)
			}
		}
	}

	return privileges
}

func (s *ClassiStrategy) parseSubclasses(_ []string) []domain.Sottoclasse {
	// For now, return empty slice as subclasses are usually in separate sections
	return []domain.Sottoclasse{}
}

func (s *ClassiStrategy) parseDadoVita(value string) domain.Dadi {
	if strings.Contains(value, "D4") {
		return domain.NewDado(1, 4, 0)
	}
	if strings.Contains(value, "D6") {
		return domain.NewDado(1, 6, 0)
	}
	if strings.Contains(value, "D8") {
		return domain.NewDado(1, 8, 0)
	}
	if strings.Contains(value, "D10") {
		return domain.NewDado(1, 10, 0)
	}
	if strings.Contains(value, "D12") {
		return domain.NewDado(1, 12, 0)
	}
	if strings.Contains(value, "D20") {
		return domain.NewDado(1, 20, 0)
	}
	return domain.NewDado(1, 6, 0) // default
}

func (s *ClassiStrategy) parseCaratteristichePrimarie(value string) []domain.Caratteristica {
	var caratteristiche []domain.Caratteristica

	// Split by common separators
	parts := regexp.MustCompile(`\s*[,/e]\s*|\s+e\s+`).Split(value, -1)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		switch strings.ToLower(part) {
		case "forza":
			caratteristiche = append(caratteristiche, domain.NewCaratteristica(domain.CaratteristicaForza, 10))
		case "destrezza":
			caratteristiche = append(caratteristiche, domain.NewCaratteristica(domain.CaratteristicaDestrezza, 10))
		case "costituzione":
			caratteristiche = append(caratteristiche, domain.NewCaratteristica(domain.CaratteristicaCostituzione, 10))
		case "intelligenza":
			caratteristiche = append(caratteristiche, domain.NewCaratteristica(domain.CaratteristicaIntelligenza, 10))
		case "saggezza":
			caratteristiche = append(caratteristiche, domain.NewCaratteristica(domain.CaratteristicaSaggezza, 10))
		case "carisma":
			caratteristiche = append(caratteristiche, domain.NewCaratteristica(domain.CaratteristicaCarisma, 10))
		}
	}

	return caratteristiche
}

func (s *ClassiStrategy) parseSalvezzeCompetenze(value string) []domain.NomeCaratteristica {
	var salvezze []domain.NomeCaratteristica

	parts := regexp.MustCompile(`\s*[,/e]\s*|\s+e\s+`).Split(value, -1)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		switch strings.ToLower(part) {
		case "forza":
			salvezze = append(salvezze, domain.Forza)
		case "destrezza":
			salvezze = append(salvezze, domain.Destrezza)
		case "costituzione":
			salvezze = append(salvezze, domain.Costituzione)
		case "intelligenza":
			salvezze = append(salvezze, domain.Intelligenza)
		case "saggezza":
			salvezze = append(salvezze, domain.Saggezza)
		case "carisma":
			salvezze = append(salvezze, domain.Carisma)
		}
	}

	return salvezze
}

func (s *ClassiStrategy) parseAbilitaCompetenze(value string) domain.Scelta {
	// Extract number of choices if present
	numeroScelte := 2 // default
	if strings.Contains(value, "Scegli ") {
		re := regexp.MustCompile(`Scegli (\d+)`)
		matches := re.FindStringSubmatch(value)
		if len(matches) > 1 {
			if num, err := strconv.Atoi(matches[1]); err == nil {
				numeroScelte = num
			}
		}
	}

	// Parse available abilities
	var opzioni []any
	abilities := []string{
		"Addestrare Animali", "Atletica", "Intimidire", "Natura",
		"Percezione", "Sopravvivenza", "Acrobazia", "Arcano",
		"Indagare", "Ingannare", "Intuizione", "Storia",
		"Medicina", "Intrattenere", "Persuasione", "Religione",
		"Rapidità di mano", "Furtività",
	}

	for _, ability := range abilities {
		if strings.Contains(value, ability) {
			opzioni = append(opzioni, ability)
		}
	}

	// If no specific abilities found, check for "qualunque abilità"
	if len(opzioni) == 0 && strings.Contains(strings.ToLower(value), "qualunque") {
		for _, ability := range abilities {
			opzioni = append(opzioni, ability)
		}
	}

	return domain.NewScelta(uint8(numeroScelte), opzioni)
}

func (s *ClassiStrategy) parseArmiCompetenze(value string) []string {
	var competenze []string

	// Common weapon types
	weaponTypes := map[string]bool{
		"Armi semplici":  true,
		"armi semplici":  true,
		"Armi da guerra": true,
		"armi da guerra": true,
	}

	// Split and check each part
	parts := regexp.MustCompile(`\s*[,e]\s*|\s+e\s+`).Split(value, -1)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if weaponTypes[part] || strings.Contains(part, "Armi") || strings.Contains(part, "armi") {
			competenze = append(competenze, part)
		}
		// Also check for specific weapons
		weapons := []string{"Balestre a mano", "Spade lunghe", "Stocchi", "Spade corte"}
		for _, weapon := range weapons {
			if strings.Contains(part, weapon) {
				competenze = append(competenze, weapon)
			}
		}
	}

	return competenze
}

func (s *ClassiStrategy) parseArmatureCompetenze(value string) []domain.CompetenzaArmatura {
	var competenze []domain.CompetenzaArmatura

	valueLower := strings.ToLower(value)
	if strings.Contains(valueLower, "leggere") {
		competenze = append(competenze, domain.CompetenzaArmatureLeggere)
	}
	if strings.Contains(valueLower, "medie") {
		competenze = append(competenze, domain.CompetenzaArmatureMedie)
	}
	if strings.Contains(valueLower, "pesanti") {
		competenze = append(competenze, domain.CompetenzaArmaturePesanti)
	}
	if strings.Contains(valueLower, "scudi") {
		competenze = append(competenze, domain.CompetenzaScudi)
	}

	return competenze
}

func (s *ClassiStrategy) parseEquipaggiamentoInizialeOpzioni(value string) []domain.EquipaggiamentoOpzione {
	var opzioni []domain.EquipaggiamentoOpzione

	// Look for option patterns like "(A) ... oppure (B) ..."
	re := regexp.MustCompile(`\([A-Z]\)\s*([^;]+?)(?:\s*oppure|\s*;|$)`)
	matches := re.FindAllStringSubmatch(value, -1)

	for _, match := range matches {
		if len(match) > 1 {
			optionText := strings.TrimSpace(match[1])
			if optionText != "" {
				opzione := domain.EquipaggiamentoOpzione{
					Etichetta: optionText,
					Oggetti:   []domain.EquipaggiamentoSlug{}, // Would need more parsing for specific items
				}
				opzioni = append(opzioni, opzione)
			}
		}
	}

	// If no options found, treat the whole thing as one option
	if len(opzioni) == 0 && value != "" {
		opzione := domain.EquipaggiamentoOpzione{
			Etichetta: value,
			Oggetti:   []domain.EquipaggiamentoSlug{},
		}
		opzioni = append(opzioni, opzione)
	}

	return opzioni
}

func (s *ClassiStrategy) fillProgressionData(progressioni *domain.Progressioni, table []map[string]string) {
	for _, row := range table {
		levelStr := row["Livello"]
		if levelStr == "" {
			continue
		}

		level, err := strconv.Atoi(levelStr)
		if err != nil {
			continue
		}

		// Parse weapon mastery progression
		if maestriaStr := row["Maestria nelle armi"]; maestriaStr != "" {
			if maestria, err := strconv.Atoi(maestriaStr); err == nil {
				progressioni.MaestriaArmi[level] = maestria
			}
		}

		// Parse extra attacks (look for "Attacco extra" in privileges)
		if privileges := row["Privilegi di classe"]; strings.Contains(privileges, "Attacco extra") {
			progressioni.AttacchiExtra[level] = 1 // Additional attack
		}

		// Parse resources (like Ira, spell slots, etc.)
		for key, value := range row {
			if !slices.Contains([]string{"Livello", "Bonus competenza", "Privilegi di classe", "Maestria nelle armi"}, key) && value != "" {
				// This could be a class resource
				if intValue, err := strconv.Atoi(strings.TrimPrefix(value, "+")); err == nil {
					risorsa := domain.RisorsaClasse{
						Chiave:  key,
						Livelli: domain.ProgressioneLivelli{level: intValue},
					}
					// Check if this resource already exists
					found := false
					for i, existing := range progressioni.Risorse {
						if existing.Chiave == key {
							progressioni.Risorse[i].Livelli[level] = intValue
							found = true
							break
						}
					}
					if !found {
						progressioni.Risorse = append(progressioni.Risorse, risorsa)
					}
				}
			}
		}
	}
}

func (s *ClassiStrategy) ContentType() ContentType {
	return ContentTypeClassi
}

func (s *ClassiStrategy) Name() string {
	return "Classi Strategy"
}

func (s *ClassiStrategy) Description() string {
	return "Parses Italian D&D 5e classes (classi) from markdown content"
}

func (s *ClassiStrategy) Validate(content []string) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}
	return nil
}

