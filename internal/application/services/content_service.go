package services

import (
	"context"
	"fmt"
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
		filter["$or"] = []bson.M{
			{"nome": bson.M{"$regex": search, "$options": "i"}},
			{"descrizione": bson.M{"$regex": search, "$options": "i"}},
			{"contenuto_markdown": bson.M{"$regex": search, "$options": "i"}},
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
		SetSort(bson.D{{"nome", 1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))
	
	// Get items
	items, err := s.mongoClient.Find(ctx, collection, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find documents: %w", err)
	}
	
	return items, totalCount, nil
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

// Search performs cross-collection search
func (s *ContentService) Search(ctx context.Context, query string, collections []string, limit int) ([]map[string]interface{}, error) {
	var allResults []map[string]interface{}
	
	// Validate collections
	validCollections := make([]string, 0, len(collections))
	for _, col := range collections {
		if isValidCollection(col) {
			validCollections = append(validCollections, col)
		}
	}
	
	if len(validCollections) == 0 {
		return nil, fmt.Errorf("no valid collections provided")
	}
	
	// Build search filter
	searchFilter := bson.M{
		"$or": []bson.M{
			{"nome": bson.M{"$regex": query, "$options": "i"}},
			{"descrizione": bson.M{"$regex": query, "$options": "i"}},
			{"contenuto_markdown": bson.M{"$regex": query, "$options": "i"}},
		},
	}
	
	// Search each collection
	for _, collection := range validCollections {
		opts := options.Find().
			SetSort(bson.D{{"nome", 1}}).
			SetLimit(int64(limit / len(validCollections)))
		
		items, err := s.mongoClient.Find(ctx, collection, searchFilter, opts)
		if err != nil {
			continue // Skip errors for individual collections
		}
		
		// Add collection info to each item
		for _, item := range items {
			item["_collection"] = collection
			item["_collection_title"] = getCollectionTitle(collection)
		}
		
		allResults = append(allResults, items...)
	}
	
	// Limit total results
	if len(allResults) > limit {
		allResults = allResults[:limit]
	}
	
	return allResults, nil
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