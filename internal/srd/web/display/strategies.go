package display

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/dto"
)

type DisplayElementStrategy interface {
	GetElements(doc map[string]any) []dto.DisplayElementDTO
	GetCollectionType() string
}

type IncantesimiDisplayStrategy struct{}

func (s *IncantesimiDisplayStrategy) GetCollectionType() string {
	return "incantesimi"
}

func (s *IncantesimiDisplayStrategy) GetElements(doc map[string]any) []dto.DisplayElementDTO {
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

func (s *OggettiMagiciDisplayStrategy) GetElements(doc map[string]any) []dto.DisplayElementDTO {
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

func (s *MostriDisplayStrategy) GetElements(doc map[string]any) []dto.DisplayElementDTO {
	var elements []dto.DisplayElementDTO

	rawContent := getRawContent(doc)

	if ca := extractMonsterCA(rawContent); ca != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: fmt.Sprintf("CA %s", ca),
			Type:  "ac",
		})
	}
	if pf := extractMonsterPF(rawContent); pf != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: fmt.Sprintf("PF %s", pf),
			Type:  "hp",
		})
	}
	if gs := extractMonsterGS(rawContent); gs != "" {
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

func (s *EquipaggiamentiDisplayStrategy) GetElements(doc map[string]any) []dto.DisplayElementDTO {
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

func (s *BackgroundsDisplayStrategy) GetElements(doc map[string]any) []dto.DisplayElementDTO {
	var elements []dto.DisplayElementDTO

	rawContent := getRawContent(doc)

	if abilita := extractBackgroundField(rawContent, "Competenze in Abilità"); abilita != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: abilita,
			Type:  "skills",
		})
	}
	if talento := extractBackgroundField(rawContent, "Talento"); talento != "" {
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

func (s *TalentiDisplayStrategy) GetElements(doc map[string]any) []dto.DisplayElementDTO {
	var elements []dto.DisplayElementDTO

	rawContent := getRawContent(doc)

	if category := extractFeatCategory(rawContent); category != "" {
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

func (s *ClassiDisplayStrategy) GetElements(doc map[string]any) []dto.DisplayElementDTO {
	var elements []dto.DisplayElementDTO

	rawContent := getRawContent(doc)

	if hitDie := extractClassHitDie(rawContent); hitDie != "" {
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

func (s *GlossarioDisplayStrategy) GetElements(doc map[string]any) []dto.DisplayElementDTO {
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

func (s *SpecieDisplayStrategy) GetElements(doc map[string]any) []dto.DisplayElementDTO {
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

func (s *DefaultDisplayStrategy) GetElements(doc map[string]any) []dto.DisplayElementDTO {
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

func getFieldValue(doc map[string]any, fieldNames ...string) string {
	for _, fieldName := range fieldNames {
		if value, exists := doc[fieldName]; exists && value != nil {
			if strValue := fmt.Sprintf("%v", value); strValue != "" && strValue != "0" {
				return strValue
			}
		}
	}
	return ""
}

func getStructuredFieldValue(doc map[string]any, fieldName string) string {
	value, exists := doc[fieldName]
	if !exists || value == nil {
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

func getRawContent(doc map[string]any) string {
	if content, ok := doc["raw_content"].(string); ok {
		return content
	}
	return ""
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
