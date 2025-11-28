package display

import (
	"fmt"

	"github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web/dto"
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

	if level := getFieldValue(doc, "livello"); level != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: fmt.Sprintf("Livello %s", level),
			Type:  "level",
		})
	}
	if school := getFieldValue(doc, "scuola"); school != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: school,
			Type:  "school",
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

	if size := getFieldValue(doc, "taglia"); size != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: size,
			Type:  "size",
		})
	}
	if cr := getFieldValue(doc, "cr", "gs", "grado_sfida"); cr != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: fmt.Sprintf("GS %s", cr),
			Type:  "challenge_rating",
		})
	}

	return elements
}

type ArmiDisplayStrategy struct{}

func (s *ArmiDisplayStrategy) GetCollectionType() string {
	return "armi"
}

func (s *ArmiDisplayStrategy) GetElements(doc map[string]any) []dto.DisplayElementDTO {
	var elements []dto.DisplayElementDTO

	if category := getFieldValue(doc, "categoria"); category != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: category,
			Type:  "category",
		})
	}
	if damage := getFieldValue(doc, "danno"); damage != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: damage,
			Type:  "damage",
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

type ArmatureDisplayStrategy struct{}

func (s *ArmatureDisplayStrategy) GetCollectionType() string {
	return "armature"
}

func (s *ArmatureDisplayStrategy) GetElements(doc map[string]any) []dto.DisplayElementDTO {
	var elements []dto.DisplayElementDTO

	if category := getFieldValue(doc, "categoria"); category != "" {
		elements = append(elements, dto.DisplayElementDTO{
			Value: category,
			Type:  "category",
		})
	}
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
