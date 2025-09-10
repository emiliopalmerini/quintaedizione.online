package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

type MostriStrategy struct{}

func NewMostriStrategy() *MostriStrategy {
	return &MostriStrategy{}
}

func (s *MostriStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
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

		// Check for new monster section (H2)
		if strings.HasPrefix(line, "## ") {
			// Process previous section if exists
			if inSection && len(currentSection) > 0 {
				mostro, err := s.parseMonsterSection(currentSection)
				if err != nil {
					context.Logger.Error(fmt.Sprintf("Failed to parse monster section: %v", err))
					continue
				}
				entities = append(entities, mostro)
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
		mostro, err := s.parseMonsterSection(currentSection)
		if err != nil {
			context.Logger.Error(fmt.Sprintf("Failed to parse last monster section: %v", err))
		} else {
			entities = append(entities, mostro)
		}
	}

	if len(entities) == 0 {
		return nil, ErrEmptyContent
	}

	return entities, nil
}

func (s *MostriStrategy) parseMonsterSection(section []string) (*domain.Mostro, error) {
	if len(section) == 0 {
		return nil, ErrEmptySectionContent
	}

	// Extract name from header
	header := section[0]
	if !strings.HasPrefix(header, "## ") {
		return nil, ErrMissingSectionTitle
	}
	nome := strings.TrimSpace(strings.TrimPrefix(header, "## "))

	// Initialize default values
	var taglia domain.Taglia = domain.TagliaMedia
	var tipo domain.TipoMostro = domain.TipoBestia
	var allineamento domain.Allineamento = domain.AllineamentoNonAllineato
	var gradoSfida int
	var puntiEsperienza domain.PuntiEsperienza
	var classeArmatura domain.ClasseArmatura = 10
	var puntiFerita domain.PuntiFerita
	var velocita domain.Velocita = domain.NewVelocita(9, domain.UnitaMetri)
	var caratteristiche []domain.Caratteristica
	var sensibilita domain.Sensibilita
	var tiriSalvezza domain.TiriSalvezza = make(map[domain.TipoCaratteristica]int)
	var abilita domain.AbilitaMostro = make(map[domain.AbilitaSlug]int)
	var immunita domain.Immunita
	var tratti []domain.Tratto
	var azioni []domain.Azione
	var reazioni []domain.ReazioneMostro
	var azioniLeggendarie []domain.AzioneLeggendaria
	var incantesimi domain.IncantesimiMostro
	contenuto := strings.Builder{}

	// Parse metadata line (line 1: *Aberrazione Grande, Legale Malvagio*)
	if len(section) > 1 {
		metaLine := section[1]
		contenuto.WriteString(metaLine + "\n")
		
		if strings.HasPrefix(metaLine, "*") && strings.HasSuffix(metaLine, "*") {
			metaContent := strings.Trim(metaLine, "*")
			taglia, tipo, allineamento = s.parseMetadata(metaContent)
		}
	}

	// Parse remaining lines
	inTraitsSection := false
	inActionsSection := false
	inReactionsSection := false
	inLegendaryActionsSection := false
	currentSectionLines := []string{}
	currentSectionType := ""
	
	for i := 2; i < len(section); i++ {
		line := section[i]
		contenuto.WriteString(line + "\n")

		// Handle section headers
		if strings.HasPrefix(line, "### ") {
			// Process previous section if exists
			s.processSectionContent(currentSectionLines, currentSectionType, &tratti, &azioni, &reazioni, &azioniLeggendarie)
			
			currentSectionLines = []string{}
			sectionTitle := strings.TrimSpace(strings.TrimPrefix(line, "### "))
			lowerTitle := strings.ToLower(sectionTitle)
			
			if strings.Contains(lowerTitle, "tratti") {
				inTraitsSection = true
				inActionsSection = false
				inReactionsSection = false
				inLegendaryActionsSection = false
				currentSectionType = "tratti"
			} else if strings.Contains(lowerTitle, "azioni leggendarie") {
				inTraitsSection = false
				inActionsSection = false
				inReactionsSection = false
				inLegendaryActionsSection = true
				currentSectionType = "azioni_leggendarie"
			} else if strings.Contains(lowerTitle, "azioni") {
				inTraitsSection = false
				inActionsSection = true
				inReactionsSection = false
				inLegendaryActionsSection = false
				currentSectionType = "azioni"
			} else if strings.Contains(lowerTitle, "reazioni") {
				inTraitsSection = false
				inActionsSection = false
				inReactionsSection = true
				inLegendaryActionsSection = false
				currentSectionType = "reazioni"
			} else {
				inTraitsSection = false
				inActionsSection = false
				inReactionsSection = false
				inLegendaryActionsSection = false
				currentSectionType = ""
			}
			continue
		}

		// Collect section content
		if (inTraitsSection || inActionsSection || inReactionsSection || inLegendaryActionsSection) && line != "" {
			currentSectionLines = append(currentSectionLines, line)
		}

		// Parse bullet point fields
		if strings.HasPrefix(line, "- **") && strings.Contains(line, ":**") {
			field, value := s.parseFieldLine(line)
			switch field {
			case "Classe Armatura":
				classeArmatura = s.parseClasseArmatura(value)
			case "Punti Ferita":
				puntiFerita = s.parsePuntiFerita(value)
			case "Velocità":
				velocita = s.parseVelocita(value)
			case "Abilità":
				abilita = s.parseAbilita(value)
			case "Sensi":
				// Parse senses but we need to handle this properly in domain
				// For now, we'll store basic info
			case "Linguaggi":
				// Store in contenuto for now
			case "GS":
				gradoSfida, puntiEsperienza = s.parseGradoSfidaEPE(value)
			}
		}

		// Parse characteristics table
		if strings.Contains(line, "|") && strings.Contains(line, "FOR") {
			// This is the characteristic table header, parse following rows
			caratteristiche = s.parseCaratteristicheTable(section, i)
		}
	}

	// Process final section if exists
	s.processSectionContent(currentSectionLines, currentSectionType, &tratti, &azioni, &reazioni, &azioniLeggendarie)

	mostro := domain.NewMostro(
		nome,
		taglia,
		tipo,
		allineamento,
		gradoSfida,
		puntiEsperienza,
		classeArmatura,
		puntiFerita,
		velocita,
		caratteristiche,
		sensibilita,
		tiriSalvezza,
		abilita,
		immunita,
		azioni,
		tratti,
		reazioni,
		azioniLeggendarie,
		incantesimi,
		strings.TrimSpace(contenuto.String()),
	)

	return mostro, nil
}

func (s *MostriStrategy) parseMetadata(content string) (domain.Taglia, domain.TipoMostro, domain.Allineamento) {
	// Parse format: "Aberrazione Grande, Legale Malvagio"
	parts := strings.Split(content, ",")
	
	var taglia domain.Taglia = domain.TagliaMedia
	var tipo domain.TipoMostro = domain.TipoBestia
	var allineamento domain.Allineamento = domain.AllineamentoNonAllineato

	if len(parts) >= 2 {
		// Parse alignment
		alignmentPart := strings.TrimSpace(parts[1])
		allineamento = s.parseAllineamento(alignmentPart)
	}

	if len(parts) >= 1 {
		// Parse type and size: "Aberrazione Grande"
		typePart := strings.TrimSpace(parts[0])
		words := strings.Fields(typePart)
		
		if len(words) >= 1 {
			tipo = s.parseTipoMostro(words[0])
		}
		if len(words) >= 2 {
			taglia = s.parseTaglia(words[1])
		}
	}

	return taglia, tipo, allineamento
}

func (s *MostriStrategy) parseTipoMostro(value string) domain.TipoMostro {
	switch value {
	case "Aberrazione":
		return domain.TipoAberrazione
	case "Bestia":
		return domain.TipoBestia
	case "Costrutto":
		return domain.TipoCostrutto
	case "Drago":
		return domain.TipoDrago
	case "Elementale":
		return domain.TipoElementale
	case "Fata":
		return domain.TipoFata
	case "Folletto":
		return domain.TipoFolletto
	case "Gigante":
		return domain.TipoGigante
	case "Umanoide":
		return domain.TipoUmanoide
	case "Melma":
		return domain.TipoMelma
	case "Mostrosoide":
		return domain.TipoMostrosoide
	case "Non Morto":
		return domain.TipoNonMorto
	case "Pianta":
		return domain.TipoPianta
	default:
		return domain.TipoBestia // default
	}
}

func (s *MostriStrategy) parseTaglia(value string) domain.Taglia {
	switch value {
	case "Minuscola":
		return domain.TagliaMinuscola
	case "Piccola":
		return domain.TagliaPiccola
	case "Media":
		return domain.TagliaMedia
	case "Grande":
		return domain.TagliaGrande
	case "Enorme":
		return domain.TagliaEnorme
	case "Mastodontica", "Colossale":
		return domain.TagliaColossale
	default:
		return domain.TagliaMedia // default
	}
}

func (s *MostriStrategy) parseAllineamento(value string) domain.Allineamento {
	switch value {
	case "Non Allineato":
		return domain.AllineamentoNonAllineato
	case "Legale Buono":
		return domain.AllineamentoLegaleBuono
	case "Neutrale Buono":
		return domain.AllineamentoNeutraleBuono
	case "Caotico Buono":
		return domain.AllineamentoCaoticoBuono
	case "Legale Neutrale":
		return domain.AllineamentoLegaleNeutrale
	case "Neutrale":
		return domain.AllineamentoNeutrale
	case "Caotico Neutrale":
		return domain.AllineamentoCaoticoNeutrale
	case "Legale Malvagio":
		return domain.AllineamentoLegaleMalvagio
	case "Neutrale Malvagio":
		return domain.AllineamentoNeutraleMalvagio
	case "Caotico Malvagio":
		return domain.AllineamentoCaoticoMalvagio
	default:
		return domain.AllineamentoNonAllineato // default
	}
}

func (s *MostriStrategy) parseFieldLine(line string) (string, string) {
	// Parse format: "- **Field:** value"
	if strings.HasPrefix(line, "- **") && strings.Contains(line, ":**") {
		parts := strings.SplitN(line, ":**", 2)
		if len(parts) == 2 {
			field := strings.TrimSpace(strings.TrimPrefix(parts[0], "- **"))
			value := strings.TrimSpace(parts[1])
			return field, value
		}
	}
	return "", ""
}

func (s *MostriStrategy) parseClasseArmatura(value string) domain.ClasseArmatura {
	// Parse format: "17" or "17 (Armatura Naturale)"
	caValue := 10 // default

	// Extract base CA value
	re := regexp.MustCompile(`(\d+)`)
	matches := re.FindStringSubmatch(value)
	if len(matches) > 1 {
		if val, err := strconv.Atoi(matches[1]); err == nil {
			caValue = val
		}
	}

	return domain.ClasseArmatura(caValue)
}

func (s *MostriStrategy) parsePuntiFerita(value string) domain.PuntiFerita {
	// Parse format: "150 (20d10 + 40)"
	pf := 0
	formula := ""

	// Extract PF value
	re := regexp.MustCompile(`(\d+)`)
	matches := re.FindStringSubmatch(value)
	if len(matches) > 1 {
		if val, err := strconv.Atoi(matches[1]); err == nil {
			pf = val
		}
	}

	// Extract formula if present
	if strings.Contains(value, "(") {
		re = regexp.MustCompile(`\((.*?)\)`)
		matches = re.FindStringSubmatch(value)
		if len(matches) > 1 {
			formula = matches[1]
		}
	}

	// Parse formula to create Dadi if possible
	var dadi domain.Dadi
	if formula != "" {
		// Try to parse formula like "20d10 + 40"
		re := regexp.MustCompile(`(\d+)d(\d+)(?:\s*\+\s*(\d+))?`)
		matches := re.FindStringSubmatch(formula)
		if len(matches) >= 3 {
			numero, _ := strconv.Atoi(matches[1])
			facce, _ := strconv.Atoi(matches[2])
			bonus := 0
			if len(matches) > 3 && matches[3] != "" {
				bonus, _ = strconv.Atoi(matches[3])
			}
			dadi = domain.NewDado(numero, facce, bonus)
		}
	}
	return domain.NewPuntiFerita(pf, dadi)
}

func (s *MostriStrategy) parseVelocita(value string) domain.Velocita {
	// Parse format: "3 m, Nuoto 12 m" - for now, just parse the first speed value
	// This is simplified - we could expand this to handle multiple speeds
	parts := strings.Split(value, ",")
	if len(parts) > 0 {
		firstPart := strings.TrimSpace(parts[0])
		
		// Extract speed value
		re := regexp.MustCompile(`(\d+)\s*m`)
		matches := re.FindStringSubmatch(firstPart)
		
		if len(matches) >= 2 {
			speedValue, _ := strconv.Atoi(matches[1])
			return domain.NewVelocita(speedValue, domain.UnitaMetri)
		}
	}
	
	return domain.NewVelocita(9, domain.UnitaMetri) // default
}

func (s *MostriStrategy) parseAbilita(value string) domain.AbilitaMostro {
	// Parse format: "Storia +12, Percezione +10"
	abilita := make(map[domain.AbilitaSlug]int)

	if value == "" {
		return abilita
	}

	parts := strings.Split(value, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		
		// Parse "Skill +X"
		re := regexp.MustCompile(`(.+?)\s+\+(\d+)`)
		matches := re.FindStringSubmatch(part)
		if len(matches) == 3 {
			nome := strings.TrimSpace(matches[1])
			bonus, _ := strconv.Atoi(matches[2])
			
			// Map Italian skill names to AbilitaSlug
			slug := s.mapAbilitaToSlug(nome)
			if slug != "" {
				abilita[domain.AbilitaSlug(slug)] = bonus
			}
		}
	}

	return abilita
}

func (s *MostriStrategy) mapAbilitaToSlug(nome string) string {
	// Map Italian skill names to slugs - simplified mapping
	switch strings.ToLower(nome) {
	case "storia":
		return "storia"
	case "percezione":
		return "percezione"
	case "atletica":
		return "atletica"
	case "furtività":
		return "furtivita"
	case "investigare":
		return "investigare"
	case "intuizione":
		return "intuizione"
	case "inganno":
		return "inganno"
	case "intimidire":
		return "intimidire"
	case "persuasione":
		return "persuasione"
	default:
		return strings.ToLower(strings.ReplaceAll(nome, " ", "_"))
	}
}

func (s *MostriStrategy) parseGradoSfidaEPE(value string) (int, domain.PuntiEsperienza) {
	// Parse format: "10 (PE 5,900, o 7,200 nella tana)"
	gs := 0
	peBase := 0
	peTana := 0

	// Extract GS value
	parts := strings.Split(value, "(")
	if len(parts) > 0 {
		gsValue := strings.TrimSpace(parts[0])
		gs, _ = strconv.Atoi(gsValue)
	}

	// Extract PE values if present
	if len(parts) > 1 {
		details := parts[1]
		
		// Parse PE base value
		peRe := regexp.MustCompile(`PE\s+(\d+(?:,\d+)?)`)
		peMatches := peRe.FindStringSubmatch(details)
		if len(peMatches) > 1 {
			peStr := strings.ReplaceAll(peMatches[1], ",", "")
			peBase, _ = strconv.Atoi(peStr)
		}

		// Parse PE tana value if present
		tanaRe := regexp.MustCompile(`o\s+(\d+(?:,\d+)?)\s+nella tana`)
		tanaMatches := tanaRe.FindStringSubmatch(details)
		if len(tanaMatches) > 1 {
			tanaStr := strings.ReplaceAll(tanaMatches[1], ",", "")
			peTana, _ = strconv.Atoi(tanaStr)
		}
	}

	pe := domain.PuntiEsperienza{
		Base: peBase,
		Tana: peTana,
	}

	return gs, pe
}

func (s *MostriStrategy) parseCaratteristicheTable(section []string, startIndex int) []domain.Caratteristica {
	var caratteristiche []domain.Caratteristica

	// Look for the data rows after the header
	for i := startIndex + 2; i < len(section); i++ { // Skip header and separator
		line := section[i]
		if !strings.Contains(line, "|") {
			break // End of table
		}

		// Parse table row: | FOR | 21 | +5 | +5 |
		parts := strings.Split(line, "|")
		if len(parts) >= 5 {
			nome := strings.TrimSpace(parts[1])
			if nome == "" || nome == "Caratteristica" {
				continue
			}

			valore, _ := strconv.Atoi(strings.TrimSpace(parts[2]))
			
			// Map string names to TipoCaratteristica
			var tipoCaratteristica domain.TipoCaratteristica
			switch nome {
			case "FOR":
				tipoCaratteristica = domain.CaratteristicaForza
			case "DES":
				tipoCaratteristica = domain.CaratteristicaDestrezza
			case "COS":
				tipoCaratteristica = domain.CaratteristicaCostituzione
			case "INT":
				tipoCaratteristica = domain.CaratteristicaIntelligenza
			case "SAG":
				tipoCaratteristica = domain.CaratteristicaSaggezza
			case "CAR":
				tipoCaratteristica = domain.CaratteristicaCarisma
			default:
				continue // Skip unknown characteristics
			}
			
			caratteristica := domain.NewCaratteristica(tipoCaratteristica, valore)
			caratteristiche = append(caratteristiche, caratteristica)
		}
	}

	return caratteristiche
}

func (s *MostriStrategy) processSectionContent(
	lines []string, 
	sectionType string, 
	tratti *[]domain.Tratto, 
	azioni *[]domain.Azione, 
	reazioni *[]domain.ReazioneMostro, 
	azioniLeggendarie *[]domain.AzioneLeggendaria,
) {
	if len(lines) == 0 {
		return
	}

	switch sectionType {
	case "tratti":
		if trait := s.parseTraitFromLines(lines); trait != nil {
			*tratti = append(*tratti, *trait)
		}
	case "azioni":
		if action := s.parseActionFromLines(lines); action != nil {
			*azioni = append(*azioni, *action)
		}
	case "reazioni":
		if reaction := s.parseReactionFromLines(lines); reaction != nil {
			*reazioni = append(*reazioni, *reaction)
		}
	case "azioni_leggendarie":
		if legAction := s.parseLegendaryActionFromLines(lines); legAction != nil {
			*azioniLeggendarie = append(*azioniLeggendarie, *legAction)
		}
	}
}

func (s *MostriStrategy) parseTraitFromLines(lines []string) *domain.Tratto {
	if len(lines) == 0 {
		return nil
	}

	// Join all lines for the trait
	content := strings.Join(lines, " ")
	
	// Extract trait name and description
	var nome string
	var descrizione string
	
	if strings.HasPrefix(content, "***") {
		// Format: ***Nome Tratto.*** Descrizione
		re := regexp.MustCompile(`\*\*\*(.+?)\.\*\*\*(.*)`)
		matches := re.FindStringSubmatch(content)
		if len(matches) == 3 {
			nome = strings.TrimSpace(matches[1])
			descrizione = strings.TrimSpace(matches[2])
		}
	} else {
		// Fallback: use first line as name
		nome = strings.TrimSpace(lines[0])
		if len(lines) > 1 {
			descrizione = strings.Join(lines[1:], " ")
		}
	}

	if nome == "" {
		return nil
	}

	return &domain.Tratto{
		Nome:        nome,
		Descrizione: descrizione,
	}
}

func (s *MostriStrategy) parseActionFromLines(lines []string) *domain.Azione {
	if len(lines) == 0 {
		return nil
	}

	// Join all lines for the action
	content := strings.Join(lines, " ")
	
	// Extract action name and description
	var nome string
	var descrizione string
	
	if strings.HasPrefix(content, "***") {
		// Format: ***Nome Azione.*** Descrizione
		re := regexp.MustCompile(`\*\*\*(.+?)\.\*\*\*(.*)`)
		matches := re.FindStringSubmatch(content)
		if len(matches) == 3 {
			nome = strings.TrimSpace(matches[1])
			descrizione = strings.TrimSpace(matches[2])
		}
	} else {
		// Fallback: use first line as name
		nome = strings.TrimSpace(lines[0])
		if len(lines) > 1 {
			descrizione = strings.Join(lines[1:], " ")
		}
	}

	if nome == "" {
		return nil
	}

	return &domain.Azione{
		Nome:        nome,
		Descrizione: descrizione,
	}
}

func (s *MostriStrategy) parseReactionFromLines(lines []string) *domain.ReazioneMostro {
	if len(lines) == 0 {
		return nil
	}

	// Join all lines for the reaction
	content := strings.Join(lines, " ")
	
	// Extract reaction name and description
	var nome string
	var descrizione string
	
	if strings.HasPrefix(content, "***") {
		// Format: ***Nome Reazione.*** Descrizione
		re := regexp.MustCompile(`\*\*\*(.+?)\.\*\*\*(.*)`)
		matches := re.FindStringSubmatch(content)
		if len(matches) == 3 {
			nome = strings.TrimSpace(matches[1])
			descrizione = strings.TrimSpace(matches[2])
		}
	} else {
		// Fallback: use first line as name
		nome = strings.TrimSpace(lines[0])
		if len(lines) > 1 {
			descrizione = strings.Join(lines[1:], " ")
		}
	}

	if nome == "" {
		return nil
	}

	return &domain.ReazioneMostro{
		Nome:        nome,
		Descrizione: descrizione,
	}
}

func (s *MostriStrategy) parseLegendaryActionFromLines(lines []string) *domain.AzioneLeggendaria {
	if len(lines) == 0 {
		return nil
	}

	// Join all lines for the legendary action
	content := strings.Join(lines, " ")
	
	// Extract legendary action name, cost, and description
	var nome string
	var descrizione string
	var costo int = 1 // default cost

	if strings.HasPrefix(content, "***") {
		// Format: ***Nome Azione (Costo X Azioni).*** Descrizione
		re := regexp.MustCompile(`\*\*\*(.+?)\.\*\*\*(.*)`)
		matches := re.FindStringSubmatch(content)
		if len(matches) == 3 {
			nomeWithCost := strings.TrimSpace(matches[1])
			descrizione = strings.TrimSpace(matches[2])
			
			// Check if cost is specified in the name
			costRe := regexp.MustCompile(`(.+?)\s*\((?:Costo|costa)\s*(\d+)\s*Azioni?\)`)
			costMatches := costRe.FindStringSubmatch(nomeWithCost)
			if len(costMatches) == 3 {
				nome = strings.TrimSpace(costMatches[1])
				costo, _ = strconv.Atoi(costMatches[2])
			} else {
				nome = nomeWithCost
			}
		}
	} else {
		// Fallback: use first line as name
		nome = strings.TrimSpace(lines[0])
		if len(lines) > 1 {
			descrizione = strings.Join(lines[1:], " ")
		}
	}

	if nome == "" {
		return nil
	}

	return &domain.AzioneLeggendaria{
		Nome:        nome,
		Costo:       costo,
		Descrizione: descrizione,
	}
}

func (s *MostriStrategy) ContentType() ContentType {
	return ContentTypeMostri
}

func (s *MostriStrategy) Name() string {
	return "Mostri Strategy"
}

func (s *MostriStrategy) Description() string {
	return "Parses Italian D&D 5e monsters (mostri) from markdown content"
}

func (s *MostriStrategy) Validate(content []string) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}
	return nil
}