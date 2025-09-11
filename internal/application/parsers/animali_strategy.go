package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

type AnimaliStrategy struct{}

func NewAnimaliStrategy() *AnimaliStrategy {
	return &AnimaliStrategy{}
}

func (s *AnimaliStrategy) Parse(content []string, context *ParsingContext) ([]domain.ParsedEntity, error) {
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

		// Check for new animal section (H2)
		if strings.HasPrefix(line, "## ") {
			// Process previous section if exists
			if inSection && len(currentSection) > 0 {
				animale, err := s.parseAnimalSection(currentSection)
				if err != nil {
					context.Logger.Error(fmt.Sprintf("Failed to parse animal section: %v", err))
					continue
				}
				entities = append(entities, animale)
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
		animale, err := s.parseAnimalSection(currentSection)
		if err != nil {
			context.Logger.Error(fmt.Sprintf("Failed to parse last animal section: %v", err))
		} else {
			entities = append(entities, animale)
		}
	}

	if len(entities) == 0 {
		return nil, ErrEmptyContent
	}

	return entities, nil
}

func (s *AnimaliStrategy) parseAnimalSection(section []string) (*domain.Animale, error) {
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
	var taglia domain.Taglia
	var tipo domain.TipoAnimale
	var sottotipo string
	var allineamento domain.Allineamento
	var ca domain.ClasseArmatura
	var pf domain.PuntiFerita
	var velocitaMultipla domain.VelocitaMultipla
	var caratteristiche []domain.Caratteristica
	var abilita []domain.AbilitaAnimale
	var sensi domain.Sensi
	var gradoSfida domain.GradoSfida
	var tratti []domain.Tratto
	var azioni []domain.Azione
	contenuto := strings.Builder{}

	// Parse metadata line (line 1: *Bestia Grande (Dinosauro), Non Allineato*)
	if len(section) > 1 {
		metaLine := section[1]
		contenuto.WriteString(metaLine + "\n")
		
		if strings.HasPrefix(metaLine, "*") && strings.HasSuffix(metaLine, "*") {
			metaContent := strings.Trim(metaLine, "*")
			taglia, tipo, sottotipo, allineamento = s.parseMetadata(metaContent)
		}
	}

	// Parse remaining lines
	inTraitsSection := false
	inActionsSection := false
	currentTraitLines := []string{}
	currentActionLines := []string{}
	
	for i := 2; i < len(section); i++ {
		line := section[i]
		contenuto.WriteString(line + "\n")

		// Handle section headers
		if strings.HasPrefix(line, "### ") {
			// Process previous trait/action if exists
			if inTraitsSection && len(currentTraitLines) > 0 {
				trait := s.parseTraitFromLines(currentTraitLines)
				if trait != nil {
					tratti = append(tratti, *trait)
				}
			}
			if inActionsSection && len(currentActionLines) > 0 {
				action := s.parseActionFromLines(currentActionLines)
				if action != nil {
					azioni = append(azioni, *action)
				}
			}
			
			currentTraitLines = []string{}
			currentActionLines = []string{}
			
			sectionTitle := strings.TrimSpace(strings.TrimPrefix(line, "### "))
			if strings.Contains(strings.ToLower(sectionTitle), "azioni") {
				inActionsSection = true
				inTraitsSection = false
			} else if strings.Contains(strings.ToLower(sectionTitle), "tratti") {
				inTraitsSection = true
				inActionsSection = false
			} else {
				inTraitsSection = false
				inActionsSection = false
			}
			continue
		}

		// Handle traits and actions
		if inTraitsSection && line != "" {
			currentTraitLines = append(currentTraitLines, line)
		} else if inActionsSection && line != "" {
			currentActionLines = append(currentActionLines, line)
		}

		// Parse bullet point fields
		if strings.HasPrefix(line, "- **") && strings.Contains(line, ":**") {
			field, value := s.parseFieldLine(line)
			switch field {
			case "Classe Armatura":
				ca = s.parseClasseArmatura(value)
			case "Punti Ferita":
				pf = s.parsePuntiFerita(value)
			case "Velocità":
				velocitaMultipla = s.parseVelocita(value)
			case "Abilità":
				abilita = s.parseAbilita(value)
			case "Sensi":
				sensi = s.parseSensi(value)
			case "GS":
				gradoSfida = s.parseGradoSfida(value)
			}
		}

		// Parse characteristics table
		if strings.Contains(line, "|") && strings.Contains(line, "FOR") {
			// This is the characteristic table header, parse following rows
			caratteristiche = s.parseCaratteristicheTable(section, i)
		}
	}

	// Process final trait/action if exists
	if inTraitsSection && len(currentTraitLines) > 0 {
		trait := s.parseTraitFromLines(currentTraitLines)
		if trait != nil {
			tratti = append(tratti, *trait)
		}
	}
	if inActionsSection && len(currentActionLines) > 0 {
		action := s.parseActionFromLines(currentActionLines)
		if action != nil {
			azioni = append(azioni, *action)
		}
	}

	animale := domain.NewAnimale(
		nome,
		taglia,
		tipo,
		sottotipo,
		allineamento,
		ca,
		pf,
		velocitaMultipla,
		caratteristiche,
		abilita,
		sensi,
		gradoSfida,
		tratti,
		azioni,
		strings.TrimSpace(contenuto.String()),
	)

	return animale, nil
}

func (s *AnimaliStrategy) parseMetadata(content string) (domain.Taglia, domain.TipoAnimale, string, domain.Allineamento) {
	// Parse format: "Bestia Grande (Dinosauro), Non Allineato"
	parts := strings.Split(content, ",")
	
	var taglia domain.Taglia = domain.TagliaMedia
	var tipo domain.TipoAnimale = domain.TipoAnimaleBestia
	var sottotipo string
	var allineamento domain.Allineamento = domain.AllineamentoNonAllineato

	if len(parts) >= 2 {
		// Parse alignment
		alignmentPart := strings.TrimSpace(parts[1])
		allineamento = s.parseAllineamento(alignmentPart)
	}

	if len(parts) >= 1 {
		// Parse type, size and subtype: "Bestia Grande (Dinosauro)"
		typePart := strings.TrimSpace(parts[0])
		
		// Extract subtype if present
		if strings.Contains(typePart, "(") {
			re := regexp.MustCompile(`\((.*?)\)`)
			matches := re.FindStringSubmatch(typePart)
			if len(matches) > 1 {
				sottotipo = matches[1]
			}
			// Remove subtype from main part
			typePart = re.ReplaceAllString(typePart, "")
			typePart = strings.TrimSpace(typePart)
		}
		
		// Split remaining into type and size
		words := strings.Fields(typePart)
		if len(words) >= 1 {
			tipo = s.parseTipoAnimale(words[0])
		}
		if len(words) >= 2 {
			taglia = s.parseTaglia(words[1])
		}
	}

	return taglia, tipo, sottotipo, allineamento
}

func (s *AnimaliStrategy) parseTipoAnimale(value string) domain.TipoAnimale {
	switch value {
	case "Animale":
		return domain.TipoAnimaleAnimale
	case "Bestia":
		return domain.TipoAnimaleBestia
	default:
		return domain.TipoAnimaleBestia // default
	}
}

func (s *AnimaliStrategy) parseTaglia(value string) domain.Taglia {
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
	case "Mastodontica":
		return domain.TagliaColossale
	default:
		return domain.TagliaMedia // default
	}
}

func (s *AnimaliStrategy) parseAllineamento(value string) domain.Allineamento {
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

func (s *AnimaliStrategy) parseFieldLine(line string) (string, string) {
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

func (s *AnimaliStrategy) parseClasseArmatura(value string) domain.ClasseArmatura {
	// Parse format: "13" or "13 (Armatura Naturale)"
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

func (s *AnimaliStrategy) parsePuntiFerita(value string) domain.PuntiFerita {
	// Parse format: "51 (6d10 + 18)"
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
		// Try to parse formula like "6d10 + 18"
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

func (s *AnimaliStrategy) parseVelocita(value string) domain.VelocitaMultipla {
	// Parse format: "18 m" or "9 m, Nuoto 12 m"
	velocitaBase := domain.NewVelocita(9, domain.UnitaMetri) // default
	speciali := make(map[string]domain.Velocita)

	parts := strings.Split(value, ",")
	for i, part := range parts {
		part = strings.TrimSpace(part)
		
		// Extract speed value and type
		re := regexp.MustCompile(`(\w+\s+)?(\d+)\s*m`)
		matches := re.FindStringSubmatch(part)
		
		if len(matches) >= 3 {
			speedValue, _ := strconv.Atoi(matches[2])
			vel := domain.NewVelocita(speedValue, domain.UnitaMetri)
			
			if i == 0 && matches[1] == "" {
				// First entry without type is base speed
				velocitaBase = vel
			} else {
				// Named special speed
				speedType := strings.TrimSpace(matches[1])
				if speedType == "" {
					speedType = "Base"
				}
				speciali[speedType] = vel
			}
		}
	}

	return domain.VelocitaMultipla{
		Base:     velocitaBase,
		Speciali: speciali,
	}
}

func (s *AnimaliStrategy) parseAbilita(value string) []domain.AbilitaAnimale {
	// Parse format: "Percezione +5" or "Percezione +5, Atletica +7"
	var abilita []domain.AbilitaAnimale

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
			
			abilita = append(abilita, domain.AbilitaAnimale{
				Nome:  nome,
				Bonus: bonus,
			})
		}
	}

	return abilita
}

func (s *AnimaliStrategy) parseSensi(value string) domain.Sensi {
	// Parse format: "Percezione Passiva 15" or "Scurovisione 18 m, Percezione Passiva 13"
	sensi := domain.Sensi{}

	if value == "" {
		return sensi
	}

	parts := strings.Split(value, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		
		if strings.Contains(strings.ToLower(part), "percezione passiva") {
			re := regexp.MustCompile(`percezione passiva\s+(\d+)`)
			matches := re.FindStringSubmatch(strings.ToLower(part))
			if len(matches) > 1 {
				val, _ := strconv.Atoi(matches[1])
				sensi.PercezionePassiva = val
			}
		} else if strings.Contains(strings.ToLower(part), "scurovisione") {
			re := regexp.MustCompile(`scurovisione\s+(\d+)`)
			matches := re.FindStringSubmatch(strings.ToLower(part))
			if len(matches) > 1 {
				val, _ := strconv.Atoi(matches[1])
				sensi.Scurovisione = val
			}
		} else if strings.Contains(strings.ToLower(part), "sensi ciechi") {
			re := regexp.MustCompile(`sensi ciechi\s+(\d+)`)
			matches := re.FindStringSubmatch(strings.ToLower(part))
			if len(matches) > 1 {
				val, _ := strconv.Atoi(matches[1])
				sensi.SensiCiechi = val
			}
		} else if strings.Contains(strings.ToLower(part), "tremosensore") {
			re := regexp.MustCompile(`tremosensore\s+(\d+)`)
			matches := re.FindStringSubmatch(strings.ToLower(part))
			if len(matches) > 1 {
				val, _ := strconv.Atoi(matches[1])
				sensi.Tremosensore = val
			}
		}
	}

	return sensi
}

func (s *AnimaliStrategy) parseGradoSfida(value string) domain.GradoSfida {
	// Parse format: "2 (PE 450; PB +2)"
	gs := domain.GradoSfida{}

	// Extract GS value
	parts := strings.Split(value, "(")
	if len(parts) > 0 {
		gsValue := strings.TrimSpace(parts[0])
		gs.Valore = gsValue
	}

	// Extract PE and PB if present
	if len(parts) > 1 {
		details := parts[1]
		
		// Parse PE
		peRe := regexp.MustCompile(`PE\s+(\d+)`)
		peMatches := peRe.FindStringSubmatch(details)
		if len(peMatches) > 1 {
			pe, _ := strconv.Atoi(peMatches[1])
			gs.PuntiEsperienza = pe
		}

		// Parse PB
		pbRe := regexp.MustCompile(`PB\s+\+(\d+)`)
		pbMatches := pbRe.FindStringSubmatch(details)
		if len(pbMatches) > 1 {
			pb, _ := strconv.Atoi(pbMatches[1])
			gs.BonusCompetenza = pb
		}
	}

	return gs
}

func (s *AnimaliStrategy) parseCaratteristicheTable(section []string, startIndex int) []domain.Caratteristica {
	var caratteristiche []domain.Caratteristica

	// Look for the data rows after the header
	for i := startIndex + 2; i < len(section); i++ { // Skip header and separator
		line := section[i]
		if !strings.Contains(line, "|") {
			break // End of table
		}

		// Parse table row: | FOR | 19 | +4 | +4 |
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

func (s *AnimaliStrategy) parseTraitFromLines(lines []string) *domain.Tratto {
	if len(lines) == 0 {
		return nil
	}

	// Join all lines for the trait
	content := strings.Join(lines, " ")
	
	// Extract trait name (usually in bold at the beginning)
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

func (s *AnimaliStrategy) parseActionFromLines(lines []string) *domain.Azione {
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

func (s *AnimaliStrategy) ContentType() ContentType {
	return ContentTypeAnimali
}

func (s *AnimaliStrategy) Name() string {
	return "Animali Strategy"
}

func (s *AnimaliStrategy) Description() string {
	return "Parses Italian D&D 5e animals (animali) from markdown content"
}

func (s *AnimaliStrategy) Validate(content []string) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}
	return nil
}