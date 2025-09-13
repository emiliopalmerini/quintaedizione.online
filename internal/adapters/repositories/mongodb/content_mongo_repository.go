package mongodb

import (
	"context"
	"fmt"
	"slices"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain/repositories"
	pkgMongodb "github.com/emiliopalmerini/due-draghi-5e-srd/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ContentMongoRepository implements ContentRepository using MongoDB
type ContentMongoRepository struct {
	client *pkgMongodb.Client
}

// NewContentMongoRepository creates a new MongoDB content repository
func NewContentMongoRepository(client *pkgMongodb.Client) repositories.ContentRepository {
	return &ContentMongoRepository{
		client: client,
	}
}

// GetCollectionItems retrieves items from any collection with pagination and pre-built filter
func (r *ContentMongoRepository) GetCollectionItems(ctx context.Context, collection string, filter bson.M, skip int64, limit int64) ([]map[string]any, int64, error) {
	if !r.isValidCollection(collection) {
		return nil, 0, fmt.Errorf("invalid collection: %s", collection)
	}

	// Get total count
	totalCount, err := r.CountCollection(ctx, collection, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	// Get items
	items, err := r.FindCollectionMaps(ctx, collection, filter, skip, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find documents: %w", err)
	}

	return items, totalCount, nil
}

// GetCollectionItemsWithFilters retrieves items from any collection with pagination and pre-built filter
func (r *ContentMongoRepository) GetCollectionItemsWithFilters(ctx context.Context, collection string, filter bson.M, skip int64, limit int64) ([]map[string]any, int64, error) {
	if !r.isValidCollection(collection) {
		return nil, 0, fmt.Errorf("invalid collection: %s", collection)
	}

	// Get total count
	totalCount, err := r.CountCollection(ctx, collection, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	// Get items
	items, err := r.FindCollectionMaps(ctx, collection, filter, skip, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find documents: %w", err)
	}

	return items, totalCount, nil
}

// GetItemBySlug retrieves a specific item by slug from any collection
func (r *ContentMongoRepository) GetItemBySlug(ctx context.Context, collection, slug string) (map[string]any, error) {
	if !r.isValidCollection(collection) {
		return nil, fmt.Errorf("invalid collection: %s", collection)
	}

	filter := bson.M{"slug": slug}
	
	coll := r.client.GetCollection(collection)
	var result map[string]any
	
	err := coll.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to find item with slug %s in collection %s: %w", slug, collection, err)
	}

	return result, nil
}

// GetCollectionStats retrieves statistics for all collections
func (r *ContentMongoRepository) GetCollectionStats(ctx context.Context) ([]map[string]any, error) {
	validCollections := r.getValidCollections()
	stats := make([]map[string]any, 0, len(validCollections))

	for _, collection := range validCollections {
		count, err := r.CountCollection(ctx, collection, bson.M{})
		if err != nil {
			// Don't fail completely, just log and continue
			count = 0
		}
		
		stats = append(stats, map[string]any{
			"name":  collection,
			"count": count,
			"title": r.getCollectionTitle(collection),
		})
	}

	return stats, nil
}

// CountCollection counts items in a specific collection with optional filter
func (r *ContentMongoRepository) CountCollection(ctx context.Context, collection string, filter bson.M) (int64, error) {
	if !r.isValidCollection(collection) {
		return 0, fmt.Errorf("invalid collection: %s", collection)
	}

	coll := r.client.GetCollection(collection)
	return coll.CountDocuments(ctx, filter)
}

// FindCollectionMaps finds documents as maps in a collection
func (r *ContentMongoRepository) FindCollectionMaps(ctx context.Context, collection string, filter bson.M, skip, limit int64) ([]map[string]any, error) {
	if !r.isValidCollection(collection) {
		return nil, fmt.Errorf("invalid collection: %s", collection)
	}

	coll := r.client.GetCollection(collection)
	
	opts := options.Find()
	if skip > 0 {
		opts.SetSkip(skip)
	}
	if limit > 0 {
		opts.SetLimit(limit)
	}
	// Sort by nome for consistent ordering
	opts.SetSort(bson.D{{Key: "nome", Value: 1}})

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents in collection %s: %w", collection, err)
	}
	defer cursor.Close(ctx)

	var results []map[string]any
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode documents from collection %s: %w", collection, err)
	}

	return results, nil
}

// GetAdjacentItems gets the previous and next items for navigation
func (r *ContentMongoRepository) GetAdjacentItems(ctx context.Context, collection, currentSlug string) (prevSlug, nextSlug *string, err error) {
	if !r.isValidCollection(collection) {
		return nil, nil, fmt.Errorf("invalid collection: %s", collection)
	}

	coll := r.client.GetCollection(collection)
	
	// Get current item to find its position
	currentFilter := bson.M{"slug": currentSlug}
	var currentItem map[string]any
	err = coll.FindOne(ctx, currentFilter).Decode(&currentItem)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find current item: %w", err)
	}
	
	currentNome, ok := currentItem["nome"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("current item has no valid nome field")
	}

	// Find previous item (nome < current, ordered desc, limit 1)
	prevFilter := bson.M{
		"nome": bson.M{"$lt": currentNome},
	}
	prevOpts := options.FindOne().SetSort(bson.D{{Key: "nome", Value: -1}})
	
	var prevItem map[string]any
	err = coll.FindOne(ctx, prevFilter, prevOpts).Decode(&prevItem)
	if err == nil {
		if slug, ok := prevItem["slug"].(string); ok {
			prevSlug = &slug
		}
	}

	// Find next item (nome > current, ordered asc, limit 1)
	nextFilter := bson.M{
		"nome": bson.M{"$gt": currentNome},
	}
	nextOpts := options.FindOne().SetSort(bson.D{{Key: "nome", Value: 1}})
	
	var nextItem map[string]any
	err = coll.FindOne(ctx, nextFilter, nextOpts).Decode(&nextItem)
	if err == nil {
		if slug, ok := nextItem["slug"].(string); ok {
			nextSlug = &slug
		}
	}

	return prevSlug, nextSlug, nil
}

// isValidCollection checks if a collection name is valid
func (r *ContentMongoRepository) isValidCollection(collection string) bool {
	validCollections := r.getValidCollections()
	return slices.Contains(validCollections, collection)
}

// getValidCollections returns the list of valid collection names
func (r *ContentMongoRepository) getValidCollections() []string {
	return []string{
		"incantesimi", "mostri", "classi", "backgrounds", "equipaggiamenti",
		"oggetti_magici", "armi", "armature", "talenti", "servizi",
		"strumenti", "animali", "regole", "cavalcature_veicoli",
	}
}

// getCollectionTitle returns a display title for a collection
func (r *ContentMongoRepository) getCollectionTitle(collection string) string {
	titles := map[string]string{
		"incantesimi":         "Incantesimi",
		"mostri":              "Mostri",
		"classi":              "Classi",
		"backgrounds":         "Background",
		"equipaggiamenti":     "Equipaggiamento",
		"armi":                "Armi",
		"armature":            "Armature",
		"oggetti_magici":      "Oggetti Magici",
		"talenti":             "Talenti",
		"servizi":             "Servizi",
		"strumenti":           "Strumenti",
		"animali":             "Animali",
		"regole":              "Regole",
		"cavalcature_veicoli": "Cavalcature e Veicoli",
	}

	if title, exists := titles[collection]; exists {
		return title
	}

	return collection
}

