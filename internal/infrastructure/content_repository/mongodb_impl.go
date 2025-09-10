package content_repository

import (
	"context"
	"fmt"
	"slices"

	"github.com/emiliopalmerini/due-draghi-5e-srd/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBRepository implements Repository using MongoDB client directly for map access
type MongoDBRepository struct {
	client *mongodb.Client
}

// NewMongoDBRepository creates a new MongoDBRepository
func NewMongoDBRepository(client *mongodb.Client) Repository {
	return &MongoDBRepository{
		client: client,
	}
}

// getValidCollections returns the list of valid collection names
func (r *MongoDBRepository) getValidCollections() []string {
	return []string{
		"incantesimi",
		"mostri", 
		"classi",
		"backgrounds",
		"equipaggiamenti", // Note: MongoDB has "equipaggiamenti" but repository uses "equipaggiamento"
		"armi",
		"armature",
		"oggetti_magici",
		"talenti",
		"servizi",
		"strumenti", 
		"animali",
		"regole",
		"cavalcature_veicoli",
	}
}

// isValidCollection validates if a collection name is supported
func (r *MongoDBRepository) isValidCollection(collection string) bool {
	validCollections := r.getValidCollections()
	return slices.Contains(validCollections, collection)
}

// FindMaps retrieves items as maps with pagination and search
func (r *MongoDBRepository) FindMaps(ctx context.Context, collection string, filter bson.M, skip, limit int64) ([]map[string]interface{}, error) {
	if !r.isValidCollection(collection) {
		return nil, fmt.Errorf("invalid collection: %s", collection)
	}

	mongoCollection := r.client.GetCollection(collection)
	
	opts := options.Find().
		SetSort(bson.D{{Key: "value.nome", Value: 1}}).
		SetSkip(skip).
		SetLimit(limit)

	cursor, err := mongoCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents in %s: %w", collection, err)
	}
	defer cursor.Close(ctx)

	var items []map[string]interface{}
	for cursor.Next(ctx) {
		var doc map[string]interface{}
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}
		
		// Extract the 'value' field and add to root level for compatibility
		if value, exists := doc["value"]; exists {
			if valueMap, ok := value.(map[string]interface{}); ok {
				// Merge value fields into root document for compatibility
				for k, v := range valueMap {
					doc[k] = v
				}
			}
		}
		
		items = append(items, doc)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return items, nil
}

// FindOneMap retrieves a single item as a map
func (r *MongoDBRepository) FindOneMap(ctx context.Context, collection string, filter bson.M) (map[string]interface{}, error) {
	if !r.isValidCollection(collection) {
		return nil, fmt.Errorf("invalid collection: %s", collection)
	}

	mongoCollection := r.client.GetCollection(collection)
	
	var doc map[string]interface{}
	err := mongoCollection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("failed to find document in %s: %w", collection, err)
	}
	
	// Extract the 'value' field and add to root level for compatibility
	if value, exists := doc["value"]; exists {
		if valueMap, ok := value.(map[string]interface{}); ok {
			// Merge value fields into root document for compatibility
			for k, v := range valueMap {
				doc[k] = v
			}
		}
	}
	
	// Ensure important root-level fields are preserved (contenuto, created_at, etc.)
	// These are already in doc, no need to extract them separately

	return doc, nil
}

// Count returns the total number of items matching the filter
func (r *MongoDBRepository) Count(ctx context.Context, collection string, filter bson.M) (int64, error) {
	if !r.isValidCollection(collection) {
		return 0, fmt.Errorf("invalid collection: %s", collection)
	}

	mongoCollection := r.client.GetCollection(collection)
	count, err := mongoCollection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents in %s: %w", collection, err)
	}

	return count, nil
}

// GetCollectionStats returns statistics for all collections  
func (r *MongoDBRepository) GetCollectionStats(ctx context.Context) ([]map[string]interface{}, error) {
	var collections []map[string]interface{}

	validCollections := r.getValidCollections()
	collectionTitles := map[string]string{
		"incantesimi":         "Incantesimi",
		"mostri":              "Mostri",
		"classi":              "Classi", 
		"backgrounds":         "Background",
		"equipaggiamenti":     "Equipaggiamento", // Note: collection name mismatch
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

	for _, collection := range validCollections {
		count, err := r.Count(ctx, collection, bson.M{})
		if err != nil {
			// Skip collections that have errors
			continue
		}

		title := collectionTitles[collection]
		if title == "" {
			title = collection
		}

		collections = append(collections, map[string]interface{}{
			"name":  collection,
			"title": title,
			"count": count,
		})
	}

	return collections, nil
}