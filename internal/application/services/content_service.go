package services

import (
	"context"
	"fmt"
	"time"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain/repositories"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/infrastructure"
)

// ContentService provides business logic for content operations
type ContentService struct {
	contentRepo repositories.ContentRepository
	cache       *infrastructure.SimpleCache
}

// NewContentService creates a new ContentService instance
func NewContentService(contentRepo repositories.ContentRepository) *ContentService {
	return &ContentService{
		contentRepo: contentRepo,
		cache:       infrastructure.GetGlobalCache(),
	}
}

// GetCollectionItems retrieves items from a collection with pagination and search
func (s *ContentService) GetCollectionItems(ctx context.Context, collection, search string, page, limit int) ([]map[string]interface{}, int64, error) {
	// Calculate skip
	skip := int64((page - 1) * limit)

	// Get items using the domain repository
	items, totalCount, err := s.contentRepo.GetCollectionItems(ctx, collection, search, skip, int64(limit))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get collection items: %w", err)
	}

	// Add display elements for each item
	for i, item := range items {
		items[i]["display_elements"] = s.getDisplayElements(collection, item)
	}

	return items, totalCount, nil
}

// getDisplayElements returns an array of display elements for a document based on collection type
func (s *ContentService) getDisplayElements(collection string, doc map[string]interface{}) []map[string]interface{} {
	var elements []map[string]interface{}

	switch collection {
	case "incantesimi":
		// Incantesimi - Level + School
		if level := getFieldValue(doc, "livello"); level != "" {
			elements = append(elements, map[string]interface{}{
				"value": fmt.Sprintf("Livello %s", level),
				"type":  "level",
			})
		}
		if school := getFieldValue(doc, "scuola"); school != "" {
			elements = append(elements, map[string]interface{}{
				"value": school,
				"type":  "school",
			})
		}

	case "oggetti_magici":
		// Oggetti magici - Rarity + Type
		if rarity := getFieldValue(doc, "rarita"); rarity != "" {
			elements = append(elements, map[string]interface{}{
				"value": rarity,
				"type":  "rarity",
			})
		}
		if objType := getFieldValue(doc, "tipo"); objType != "" {
			elements = append(elements, map[string]interface{}{
				"value": objType,
				"type":  "type",
			})
		}

	case "mostri":
		// Mostri - Size + Type + CR
		if size := getFieldValue(doc, "taglia"); size != "" {
			elements = append(elements, map[string]interface{}{
				"value": size,
				"type":  "size",
			})
		}
		if cr := getFieldValue(doc, "cr", "gs", "grado_sfida"); cr != "" {
			elements = append(elements, map[string]interface{}{
				"value": fmt.Sprintf("GS %s", cr),
				"type":  "challenge_rating",
			})
		}

	case "armi":
		// Armi - Category + Damage
		if category := getFieldValue(doc, "categoria"); category != "" {
			elements = append(elements, map[string]interface{}{
				"value": category,
				"type":  "category",
			})
		}
		if damage := getFieldValue(doc, "danno"); damage != "" {
			elements = append(elements, map[string]interface{}{
				"value": damage,
				"type":  "damage",
			})
		}

	case "armature":
		// Armature - Category + AC
		if category := getFieldValue(doc, "categoria"); category != "" {
			elements = append(elements, map[string]interface{}{
				"value": category,
				"type":  "category",
			})
		}
		if ac := getFieldValue(doc, "ca_base"); ac != "" {
			elements = append(elements, map[string]interface{}{
				"value": fmt.Sprintf("CA %s", ac),
				"type":  "ac",
			})
		}
	}

	// Add generic fields (cost, weight) for all collections
	// Handle structured value objects first, fallback to simple fields
	if cost := getStructuredFieldValue(doc, "costo"); cost != "" {
		elements = append(elements, map[string]interface{}{
			"value": cost,
			"type":  "cost",
		})
	}
	if weight := getStructuredFieldValue(doc, "peso"); weight != "" {
		elements = append(elements, map[string]interface{}{
			"value": weight,
			"type":  "weight",
		})
	}

	return elements
}

// getFieldValue returns the first non-empty value from the given field names
// It looks in the "value" object first, then at the document root level
func getFieldValue(doc map[string]interface{}, fieldNames ...string) string {
	// First try to access fields from the "value" object
	if valueObj, ok := doc["value"].(map[string]interface{}); ok {
		for _, fieldName := range fieldNames {
			if value, exists := valueObj[fieldName]; exists && value != nil {
				if strValue := fmt.Sprintf("%v", value); strValue != "" && strValue != "0" {
					return strValue
				}
			}
		}
	}
	
	// Fallback to document root level
	for _, fieldName := range fieldNames {
		if value, exists := doc[fieldName]; exists && value != nil {
			if strValue := fmt.Sprintf("%v", value); strValue != "" && strValue != "0" {
				return strValue
			}
		}
	}
	return ""
}

// getStructuredFieldValue extracts and formats structured domain value objects
func getStructuredFieldValue(doc map[string]interface{}, fieldName string) string {
	var value interface{}
	var exists bool
	
	// First try to access from the "value" object
	if valueObj, ok := doc["value"].(map[string]interface{}); ok {
		value, exists = valueObj[fieldName]
	}
	
	// Fallback to document root level
	if !exists || value == nil {
		value, exists = doc[fieldName]
	}
	
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

// formatCosto formats a Costo value object to display string
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

// formatPeso formats a Peso value object to display string
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

// formatVelocita formats a Velocita value object to display string
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

// formatGittata formats a GittataArma value object to display string
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

// GetItem retrieves a specific item by slug
func (s *ContentService) GetItem(ctx context.Context, collection, slug string) (map[string]interface{}, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("item:%s:%s", collection, slug)
	if cached, found := s.cache.Get(cacheKey); found {
		if item, ok := cached.(map[string]interface{}); ok {
			return item, nil
		}
	}

	item, err := s.contentRepo.GetItemBySlug(ctx, collection, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to find item: %w", err)
	}

	// Cache the item for 10 minutes
	s.cache.Set(cacheKey, item, 10*time.Minute)

	return item, nil
}

// GetStats retrieves database statistics
func (s *ContentService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	collections, err := s.contentRepo.GetCollectionStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection stats: %w", err)
	}

	stats := map[string]interface{}{
		"collections": make(map[string]int64),
		"total_items": int64(0),
	}

	for _, collection := range collections {
		if name, ok := collection["name"].(string); ok {
			if count, ok := collection["count"].(int64); ok {
				stats["collections"].(map[string]int64)[name] = count
				stats["total_items"] = stats["total_items"].(int64) + count
			}
		}
	}

	return stats, nil
}

// GetCollectionStats retrieves statistics for all collections
func (s *ContentService) GetCollectionStats(ctx context.Context) ([]map[string]interface{}, error) {
	return s.contentRepo.GetCollectionStats(ctx)
}

// GetAdjacentItems gets the previous and next items for navigation
func (s *ContentService) GetAdjacentItems(ctx context.Context, collection, currentSlug string) (prevSlug, nextSlug *string, err error) {
	return s.contentRepo.GetAdjacentItems(ctx, collection, currentSlug)
}

