package display

import (
	"fmt"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/adapters/web/dto"
)

// DisplayElementStrategy defines how to extract display elements for a collection
type DisplayElementStrategy interface {
	GetElements(doc map[string]interface{}) []dto.DisplayElementDTO
	GetCollectionType() string
}

// IncantesimiDisplayStrategy handles display elements for spells
type IncantesimiDisplayStrategy struct{}

func (s *IncantesimiDisplayStrategy) GetCollectionType() string {
	return "incantesimi"
}

func (s *IncantesimiDisplayStrategy) GetElements(doc map[string]interface{}) []dto.DisplayElementDTO {
	var elements []dto.DisplayElementDTO

	// Incantesimi - Level + School
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

	// Add generic fields (cost, weight)
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

// OggettiMagiciDisplayStrategy handles display elements for magic items
type OggettiMagiciDisplayStrategy struct{}

func (s *OggettiMagiciDisplayStrategy) GetCollectionType() string {
	return "oggetti_magici"
}

func (s *OggettiMagiciDisplayStrategy) GetElements(doc map[string]interface{}) []dto.DisplayElementDTO {
	var elements []dto.DisplayElementDTO

	// Oggetti magici - Rarity + Type
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

	// Add generic fields
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

// MostriDisplayStrategy handles display elements for monsters
type MostriDisplayStrategy struct{}

func (s *MostriDisplayStrategy) GetCollectionType() string {
	return "mostri"
}

func (s *MostriDisplayStrategy) GetElements(doc map[string]interface{}) []dto.DisplayElementDTO {
	var elements []dto.DisplayElementDTO

	// Mostri - Size + Type + CR
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

// ArmiDisplayStrategy handles display elements for weapons
type ArmiDisplayStrategy struct{}

func (s *ArmiDisplayStrategy) GetCollectionType() string {
	return "armi"
}

func (s *ArmiDisplayStrategy) GetElements(doc map[string]interface{}) []dto.DisplayElementDTO {
	var elements []dto.DisplayElementDTO

	// Armi - Category + Damage
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

	// Add generic fields
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

// ArmatureDisplayStrategy handles display elements for armor
type ArmatureDisplayStrategy struct{}

func (s *ArmatureDisplayStrategy) GetCollectionType() string {
	return "armature"
}

func (s *ArmatureDisplayStrategy) GetElements(doc map[string]interface{}) []dto.DisplayElementDTO {
	var elements []dto.DisplayElementDTO

	// Armature - Category + AC
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

	// Add generic fields
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

// DefaultDisplayStrategy is the fallback strategy for other collections
type DefaultDisplayStrategy struct{}

func (s *DefaultDisplayStrategy) GetCollectionType() string {
	return "default"
}

func (s *DefaultDisplayStrategy) GetElements(doc map[string]interface{}) []dto.DisplayElementDTO {
	var elements []dto.DisplayElementDTO

	// Add generic fields only
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

// Helper functions (extracted from content service)
func getFieldValue(doc map[string]interface{}, fieldNames ...string) string {
	for _, fieldName := range fieldNames {
		if value, exists := doc[fieldName]; exists && value != nil {
			if strValue := fmt.Sprintf("%v", value); strValue != "" && strValue != "0" {
				return strValue
			}
		}
	}
	return ""
}

func getStructuredFieldValue(doc map[string]interface{}, fieldName string) string {
	value, exists := doc[fieldName]
	if !exists || value == nil {
		return ""
	}

	// Handle different structured types
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
		// Fallback to simple field extraction
		if strValue := fmt.Sprintf("%v", value); strValue != "" && strValue != "0" {
			return strValue
		}
	}
	return ""
}

func formatCosto(value interface{}) string {
	if costoMap, ok := value.(map[string]interface{}); ok {
		valore, valoreOk := costoMap["valore"]
		valuta, valutaOk := costoMap["valuta"]

		if valoreOk && valutaOk {
			return fmt.Sprintf("%v %v", valore, valuta)
		}
	}
	return ""
}

func formatPeso(value interface{}) string {
	if pesoMap, ok := value.(map[string]interface{}); ok {
		valore, valoreOk := pesoMap["valore"]
		unita, unitaOk := pesoMap["unita"]

		if valoreOk && unitaOk {
			return fmt.Sprintf("%v %v", valore, unita)
		}
	}
	return ""
}

func formatVelocita(value interface{}) string {
	if velocitaMap, ok := value.(map[string]interface{}); ok {
		valore, valoreOk := velocitaMap["valore"]
		unita, unitaOk := velocitaMap["unita"]

		if valoreOk && unitaOk {
			return fmt.Sprintf("%v %v", valore, unita)
		}
	}
	return ""
}

func formatGittata(value interface{}) string {
	if gittataMap, ok := value.(map[string]interface{}); ok {
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