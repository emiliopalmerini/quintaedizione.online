package display

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/dto"
)

type DisplayElementStrategy interface {
	GetElements(doc *domain.Document) []dto.DisplayElementDTO
	GetCollectionType() string
}

type IncantesimiDisplayStrategy struct{}

func (s *IncantesimiDisplayStrategy) GetCollectionType() string {
	return "incantesimi"
}

func (s *IncantesimiDisplayStrategy) GetElements(doc *domain.Document) []dto.DisplayElementDTO {
	var elements []dto.DisplayElementDTO

	rawContent := getRawContent(doc)

	if level, school := extractSpellLevelAndSchool(rawContent); level != "" || school != "" {
		if level != "" {
			elements = append(elements, dto.DisplayElementDTO{
				Value: level,
				Type:  "level",
			})
		}
		if school != "" {
			elements = append(elements, dto.DisplayElementDTO{
				Value: school,
				Type:  "school",
			})
		}
	}
	if tempo := extractSpellField(rawContent, "Tempo di Lancio"); tempo != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: tempo,
			Type:  "casting_time",
		})
	}
	if gittata := extractSpellField(rawContent, "Gittata"); gittata != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: gittata,
			Type:  "range",
		})
	}

	return elements
}

type OggettiMagiciDisplayStrategy struct{}

func (s *OggettiMagiciDisplayStrategy) GetCollectionType() string {
	return "oggetti_magici"
}

func (s *OggettiMagiciDisplayStrategy) GetElements(doc *domain.Document) []dto.DisplayElementDTO {
	var elements []dto.DisplayElementDTO

	if rarity := getFieldValue(doc, "rarita"); rarity != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: rarity,
			Type:  "rarity",
		})
	}
	if objType := getFieldValue(doc, "tipo"); objType != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: objType,
			Type:  "type",
		})
	}

	if cost := getStructuredFieldValue(doc, "costo"); cost != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: cost,
			Type:  "cost",
		})
	}
	if weight := getStructuredFieldValue(doc, "peso"); weight != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: weight,
			Type:  "weight",
		})
	}

	return elements
}

type MostriDisplayStrategy struct{}

func (s *MostriDisplayStrategy) GetCollectionType() string {
	return "mostri"
}

func (s *MostriDisplayStrategy) GetElements(doc *domain.Document) []dto.DisplayElementDTO {
	var elements []dto.DisplayElementDTO

	rawContent := getRawContent(doc)

	ca := getFieldValue(doc, "ac", "ca")
	if ca == "" {
		ca = extractMonsterCA(rawContent)
	}
	if ca != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: fmt.Sprintf("CA %s", ca),
			Type:  "ac",
		})
	}
	pf := getFieldValue(doc, "hp", "pf")
	if pf == "" {
		pf = extractMonsterPF(rawContent)
	}
	if pf != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: fmt.Sprintf("PF %s", pf),
			Type:  "hp",
		})
	}
	gs := getFieldValue(doc, "grado_sfida", "cr")
	if gs == "" {
		gs = extractMonsterGS(rawContent)
	}
	if gs != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: fmt.Sprintf("GS %s", gs),
			Type:  "challenge_rating",
		})
	}

	return elements
}

type EquipaggiamentiDisplayStrategy struct{}

func (s *EquipaggiamentiDisplayStrategy) GetCollectionType() string {
	return "equipaggiamenti"
}

func (s *EquipaggiamentiDisplayStrategy) GetElements(doc *domain.Document) []dto.DisplayElementDTO {
	var elements []dto.DisplayElementDTO

	if category := getFieldValue(doc, "categoria"); category != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: category,
			Type:  "category",
		})
	}

	// Weapon-specific: damage
	if damage := getFieldValue(doc, "danno"); damage != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: damage,
			Type:  "damage",
		})
	}

	// Armor-specific: AC
	if ac := getFieldValue(doc, "ca_base"); ac != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: fmt.Sprintf("CA %s", ac),
			Type:  "ac",
		})
	}

	if cost := getStructuredFieldValue(doc, "costo"); cost != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: cost,
			Type:  "cost",
		})
	}
	if weight := getStructuredFieldValue(doc, "peso"); weight != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: weight,
			Type:  "weight",
		})
	}

	return elements
}

type BackgroundsDisplayStrategy struct{}

func (s *BackgroundsDisplayStrategy) GetCollectionType() string {
	return "backgrounds"
}

func (s *BackgroundsDisplayStrategy) GetElements(doc *domain.Document) []dto.DisplayElementDTO {
	var elements []dto.DisplayElementDTO

	rawContent := getRawContent(doc)

	abilita := getFieldValue(doc, "skill_proficiencies", "competenze_abilita", "competenze_nelle_abilita")
	if abilita == "" {
		abilita = extractBackgroundFieldAny(rawContent, "Competenze nelle Abilita", "Competenze nelle Abilità", "Competenze in Abilita", "Competenze in Abilità")
	}
	if abilita != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: abilita,
			Type:  "skills",
		})
	}
	talento := getFieldValue(doc, "feat", "talento")
	if talento == "" {
		talento = extractBackgroundField(rawContent, "Talento")
	}
	if talento != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: talento,
			Type:  "feat",
		})
	}

	return elements
}

type TalentiDisplayStrategy struct{}

func (s *TalentiDisplayStrategy) GetCollectionType() string {
	return "talenti"
}

func (s *TalentiDisplayStrategy) GetElements(doc *domain.Document) []dto.DisplayElementDTO {
	var elements []dto.DisplayElementDTO

	rawContent := getRawContent(doc)

	category := getFieldValue(doc, "categoria")
	if category == "" {
		category = extractFeatCategory(rawContent)
	}
	if category != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: category,
			Type:  "category",
		})
	}

	return elements
}

type ClassiDisplayStrategy struct{}

func (s *ClassiDisplayStrategy) GetCollectionType() string {
	return "classi"
}

func (s *ClassiDisplayStrategy) GetElements(doc *domain.Document) []dto.DisplayElementDTO {
	var elements []dto.DisplayElementDTO

	rawContent := getRawContent(doc)

	hitDie := getFieldValue(doc, "hit_die")
	if hitDie == "" {
		hitDie = extractClassHitDie(rawContent)
	}
	if hitDie != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: hitDie,
			Type:  "hit_die",
		})
	}

	return elements
}

type GlossarioDisplayStrategy struct{}

func (s *GlossarioDisplayStrategy) GetCollectionType() string {
	return "glossario"
}

func (s *GlossarioDisplayStrategy) GetElements(doc *domain.Document) []dto.DisplayElementDTO {
	var elements []dto.DisplayElementDTO

	if category := getFieldValue(doc, "categoria"); category != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: category,
			Type:  "category",
		})
	}

	return elements
}

type SpecieDisplayStrategy struct{}

func (s *SpecieDisplayStrategy) GetCollectionType() string {
	return "specie"
}

func (s *SpecieDisplayStrategy) GetElements(doc *domain.Document) []dto.DisplayElementDTO {
	var elements []dto.DisplayElementDTO

	if creatureType := getFieldValue(doc, "tipo_creatura"); creatureType != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: creatureType,
			Type:  "creature_type",
		})
	}
	if size := getFieldValue(doc, "taglia"); size != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: size,
			Type:  "size",
		})
	}
	if speed := getFieldValue(doc, "velocita"); speed != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: speed,
			Type:  "speed",
		})
	}

	return elements
}

type DefaultDisplayStrategy struct{}

func (s *DefaultDisplayStrategy) GetCollectionType() string {
	return "default"
}

func (s *DefaultDisplayStrategy) GetElements(doc *domain.Document) []dto.DisplayElementDTO {
	var elements []dto.DisplayElementDTO

	if cost := getStructuredFieldValue(doc, "costo"); cost != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: cost,
			Type:  "cost",
		})
	}
	if weight := getStructuredFieldValue(doc, "peso"); weight != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: weight,
			Type:  "weight",
		})
	}

	return elements
}

func getFieldValue(doc *domain.Document, fieldNames ...string) string {
	for _, fieldName := range fieldNames {
		if value := doc.GetField(fieldName); value != nil {
			if strValue := fmt.Sprintf("%v", value); strValue != "" && strValue != "0" {
				return strValue
			}
		}
	}
	return ""
}

func getStructuredFieldValue(doc *domain.Document, fieldName string) string {
	value := doc.GetField(fieldName)
	if value == nil {
		return ""
	}

	switch fieldName {
	case "costo":
		return formatCosto(value)
	case "peso":
		return formatPeso(value)
	case "velocita":
		return formatVelocita(value)
	case "gittata":
		return formatGittata(value)
	default:

		if strValue := fmt.Sprintf("%v", value); strValue != "" && strValue != "0" {
			return strValue
		}
	}
	return ""
}

func formatCosto(value interface{}) string {
	if costoMap, ok := value.(map[string]any); ok {
		valore, valoreOk := costoMap["valore"]
		valuta, valutaOk := costoMap["valuta"]

		if valoreOk && valutaOk {
			return fmt.Sprintf("%v %v", valore, valuta)
		}
	}
	return ""
}

func formatPeso(value interface{}) string {
	if pesoMap, ok := value.(map[string]any); ok {
		valore, valoreOk := pesoMap["valore"]
		unita, unitaOk := pesoMap["unita"]

		if valoreOk && unitaOk {
			return fmt.Sprintf("%v %v", valore, unita)
		}
	}
	return ""
}

func formatVelocita(value interface{}) string {
	if velocitaMap, ok := value.(map[string]any); ok {
		valore, valoreOk := velocitaMap["valore"]
		unita, unitaOk := velocitaMap["unita"]

		if valoreOk && unitaOk {
			return fmt.Sprintf("%v %v", valore, unita)
		}
	}
	return ""
}

func formatGittata(value interface{}) string {
	if gittataMap, ok := value.(map[string]any); ok {
		normale, normaleOk := gittataMap["normale"]
		lunga, lungaOk := gittataMap["lunga"]

		if normaleOk && lungaOk {
			return fmt.Sprintf("%v/%v", normale, lunga)
		} else if normaleOk {
			return fmt.Sprintf("%v", normale)
		}
	}
	return ""
}

func getRawContent(doc *domain.Document) string {
	return doc.RawContent.String()
}

func extractMonsterCA(rawContent string) string {
	re := regexp.MustCompile(`\*\*Classe Armatura:\*\*\s*(\d+)`)
	if matches := re.FindStringSubmatch(rawContent); len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func extractMonsterPF(rawContent string) string {
	re := regexp.MustCompile(`\*\*Punti Ferita:\*\*\s*(\d+)`)
	if matches := re.FindStringSubmatch(rawContent); len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func extractMonsterGS(rawContent string) string {
	re := regexp.MustCompile(`\*\*GS\*\*\s*(\d+(?:/\d+)?)|GS\s+(\d+(?:/\d+)?)`)
	if matches := re.FindStringSubmatch(rawContent); len(matches) > 0 {
		if matches[1] != "" {
			return matches[1]
		}
		if len(matches) > 2 && matches[2] != "" {
			return matches[2]
		}
	}
	return ""
}

func extractSpellLevelAndSchool(rawContent string) (level string, school string) {
	re := regexp.MustCompile(`^\*(?:Livello\s+(\d+)|Trucchetto(?:\s+di)?)\s+(\w+)`)
	if matches := re.FindStringSubmatch(rawContent); len(matches) > 0 {
		if matches[1] != "" {
			level = fmt.Sprintf("Livello %s", matches[1])
		} else {
			level = "Trucchetto"
		}
		if len(matches) > 2 {
			school = matches[2]
		}
	}
	return level, school
}

func extractSpellField(rawContent string, fieldName string) string {
	pattern := fmt.Sprintf(`\*\*%s:\*\*\s*([^\n]+)`, regexp.QuoteMeta(fieldName))
	re := regexp.MustCompile(pattern)
	if matches := re.FindStringSubmatch(rawContent); len(matches) > 1 {
		return strings.TrimSuffix(strings.TrimSpace(matches[1]), ".")
	}
	return ""
}

func extractBackgroundField(rawContent string, fieldName string) string {
	pattern := fmt.Sprintf(`\*\*%s:\*\*\s*([^\n]+)`, regexp.QuoteMeta(fieldName))
	re := regexp.MustCompile(pattern)
	if matches := re.FindStringSubmatch(rawContent); len(matches) > 1 {
		value := strings.TrimSpace(matches[1])
		value = strings.TrimSuffix(value, "  ")
		return value
	}
	return ""
}

func extractBackgroundFieldAny(rawContent string, fieldNames ...string) string {
	for _, fieldName := range fieldNames {
		if value := extractBackgroundField(rawContent, fieldName); value != "" {
			return value
		}
	}
	return ""
}

func extractFeatCategory(rawContent string) string {
	re := regexp.MustCompile(`^\*Talento\s+(?:di\s+)?(\w+(?:\s+di\s+\w+)?)\s*(?:\([^)]*\))?\*`)
	if matches := re.FindStringSubmatch(rawContent); len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func extractClassHitDie(rawContent string) string {
	re := regexp.MustCompile(`Dado Punti Ferita\s*\|\s*(D\d+)`)
	if matches := re.FindStringSubmatch(rawContent); len(matches) > 1 {
		return matches[1]
	}
	return ""
}
