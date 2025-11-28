package content_repository

import (
	"context"
	"fmt"
	"slices"

	"github.com/emiliopalmerini/quintaedizione.online/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBRepository struct {
	client *mongodb.Client
}

func NewMongoDBRepository(client *mongodb.Client) Repository {
	return &MongoDBRepository{
		client: client,
	}
}

func (r *MongoDBRepository) getValidCollections() []string {
	return []string{
		"incantesimi",
		"mostri",
		"classi",
		"backgrounds",
		"equipaggiamenti",
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

func (r *MongoDBRepository) isValidCollection(collection string) bool {
	validCollections := r.getValidCollections()
	return slices.Contains(validCollections, collection)
}

func (r *MongoDBRepository) FindMaps(ctx context.Context, collection string, filter bson.M, skip, limit int64) ([]map[string]any, error) {
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

	var items []map[string]any
	for cursor.Next(ctx) {
		var doc map[string]any
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		if value, exists := doc["value"]; exists {
			if valueMap, ok := value.(map[string]any); ok {

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

func (r *MongoDBRepository) FindOneMap(ctx context.Context, collection string, filter bson.M) (map[string]any, error) {
	if !r.isValidCollection(collection) {
		return nil, fmt.Errorf("invalid collection: %s", collection)
	}

	mongoCollection := r.client.GetCollection(collection)

	var doc map[string]any
	err := mongoCollection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("failed to find document in %s: %w", collection, err)
	}

	if value, exists := doc["value"]; exists {
		if valueMap, ok := value.(map[string]any); ok {

			for k, v := range valueMap {
				doc[k] = v
			}
		}
	}

	return doc, nil
}

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

func (r *MongoDBRepository) GetCollectionStats(ctx context.Context) ([]map[string]any, error) {
	var collections []map[string]any

	validCollections := r.getValidCollections()
	collectionTitles := map[string]string{
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

	for _, collection := range validCollections {
		count, err := r.Count(ctx, collection, bson.M{})
		if err != nil {

			continue
		}

		title := collectionTitles[collection]
		if title == "" {
			title = collection
		}

		collections = append(collections, map[string]any{
			"name":  collection,
			"title": title,
			"count": count,
		})
	}

	return collections, nil
}
