package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/infrastructure"
	"github.com/emiliopalmerini/due-draghi-5e-srd/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ContentService provides business logic for content operations
type ContentService struct {
	mongoClient *mongodb.Client
	cache       *infrastructure.SimpleCache
}

// NewContentService creates a new ContentService instance
func NewContentService(mongoClient *mongodb.Client) *ContentService {
	return &ContentService{
		mongoClient: mongoClient,
		cache:       infrastructure.GetGlobalCache(),
	}
}

// GetCollectionItems retrieves items from a collection with pagination and search
func (s *ContentService) GetCollectionItems(ctx context.Context, collection, search string, page, limit int) ([]map[string]interface{}, int64, error) {
	// Validate collection name
	if !isValidCollection(collection) {
		return nil, 0, fmt.Errorf("invalid collection: %s", collection)
	}

	// Build filter
	filter := bson.M{}
	if search != "" {
		// Escape special regex characters for safety
		escapedSearch := regexp.QuoteMeta(search)
		filter["$or"] = []bson.M{
			{"nome": bson.M{"$regex": escapedSearch, "$options": "i"}},
			{"titolo": bson.M{"$regex": escapedSearch, "$options": "i"}},
			{"descrizione": bson.M{"$regex": escapedSearch, "$options": "i"}},
			{"contenuto_markdown": bson.M{"$regex": escapedSearch, "$options": "i"}},
		}
	}

	// Calculate skip
	skip := (page - 1) * limit

	// Get total count
	totalCount, err := s.mongoClient.Count(ctx, collection, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	// Find options
	opts := options.Find().
		SetSort(bson.D{{Key: "nome", Value: 1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	// Get items
	items, err := s.mongoClient.Find(ctx, collection, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find documents: %w", err)
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
// For simple fields, it returns the string representation
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

// getStructuredFieldValue extracts and formats structured domain value objects
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
	// Validate collection name
	if !isValidCollection(collection) {
		return nil, fmt.Errorf("invalid collection: %s", collection)
	}

	// Try cache first
	cacheKey := fmt.Sprintf("item:%s:%s", collection, slug)
	if cached, found := s.cache.Get(cacheKey); found {
		if item, ok := cached.(map[string]interface{}); ok {
			return item, nil
		}
	}

	filter := bson.M{"slug": slug}

	item, err := s.mongoClient.FindOne(ctx, collection, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find item: %w", err)
	}

	// Cache the item for 10 minutes
	s.cache.Set(cacheKey, item, 10*time.Minute)

	return item, nil
}

// GetStats retrieves database statistics
func (s *ContentService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := map[string]interface{}{
		"collections": make(map[string]int64),
		"total_items": int64(0),
	}

	validCollections := getValidCollections()

	for _, collection := range validCollections {
		count, err := s.mongoClient.Count(ctx, collection, bson.M{})
		if err != nil {
			continue
		}

		stats["collections"].(map[string]int64)[collection] = count
		stats["total_items"] = stats["total_items"].(int64) + count
	}

	return stats, nil
}

// GetCollectionStats retrieves statistics for all collections
func (s *ContentService) GetCollectionStats(ctx context.Context) ([]map[string]interface{}, error) {
	var collections []map[string]interface{}

	validCollections := getValidCollections()

	for _, collection := range validCollections {
		count, err := s.mongoClient.Count(ctx, collection, bson.M{})
		if err != nil {
			continue
		}

		collections = append(collections, map[string]interface{}{
			"name":  collection,
			"title": getCollectionTitle(collection),
			"count": count,
		})
	}

	return collections, nil
}

// Helper functions
func isValidCollection(collection string) bool {
	validCollections := getValidCollections()
	for _, valid := range validCollections {
		if valid == collection {
			return true
		}
	}
	return false
}

func getValidCollections() []string {
	return []string{
		"incantesimi",
		"mostri",
		"classi",
		"backgrounds",
		"equipaggiamento",
		"armi",
		"armature",
		"oggetti_magici",
		"talenti",
		"servizi",
		"strumenti",
		"animali",
		"documenti",
	}
}

func getCollectionTitle(collection string) string {
	titles := map[string]string{
		"incantesimi":     "Incantesimi",
		"mostri":          "Mostri",
		"classi":          "Classi",
		"backgrounds":     "Background",
		"equipaggiamento": "Equipaggiamento",
		"armi":            "Armi",
		"armature":        "Armature",
		"oggetti_magici":  "Oggetti Magici",
		"talenti":         "Talenti",
		"servizi":         "Servizi",
		"strumenti":       "Strumenti",
		"animali":         "Animali",
		"documenti":       "Documenti",
	}

	if title, exists := titles[collection]; exists {
		return title
	}

	return strings.Title(collection)
}
