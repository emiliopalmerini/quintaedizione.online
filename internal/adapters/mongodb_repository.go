package adapters

import (
	"context"
	"fmt"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
	"github.com/emiliopalmerini/due-draghi-5e-srd/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoParserRepository implements the ParserRepository interface for MongoDB
type MongoParserRepository struct {
	client *mongodb.Client
}

// NewMongoParserRepository creates a new MongoDB parser repository
func NewMongoParserRepository(client *mongodb.Client) domain.ParserRepository {
	return &MongoParserRepository{
		client: client,
	}
}

// UpsertMany performs bulk upsert operations on a collection
func (r *MongoParserRepository) UpsertMany(collectionName string, uniqueFields []string, docs []map[string]any) (int, error) {
	if len(docs) == 0 {
		return 0, nil
	}

	collection := r.client.GetCollection(collectionName)
	ctx := context.Background()

	// Perform bulk upsert operations
	var operations []mongo.WriteModel

	for _, doc := range docs {
		// Create filter based on unique fields
		filter := bson.M{}
		hasUniqueField := false

		for _, field := range uniqueFields {
			if value, exists := doc[field]; exists && value != nil {
				filter[field] = value
				hasUniqueField = true
				break // Use first available unique field
			}
		}

		// Skip documents without unique fields
		if !hasUniqueField {
			continue
		}

		// Create upsert operation
		operation := mongo.NewReplaceOneModel().
			SetFilter(filter).
			SetReplacement(doc).
			SetUpsert(true)

		operations = append(operations, operation)
	}

	if len(operations) == 0 {
		return 0, fmt.Errorf("no valid documents to upsert")
	}

	// Execute bulk operations
	opts := options.BulkWrite().SetOrdered(false)
	result, err := collection.BulkWrite(ctx, operations, opts)
	if err != nil {
		return 0, fmt.Errorf("bulk upsert failed: %w", err)
	}

	// Return count of modified/inserted documents
	return int(result.UpsertedCount + result.ModifiedCount), nil
}

// Count returns the number of documents in a collection
func (r *MongoParserRepository) Count(collectionName string) (int64, error) {
	collection := r.client.GetCollection(collectionName)
	ctx := context.Background()

	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("failed to count documents in %s: %w", collectionName, err)
	}

	return count, nil
}

// FindByFilter finds documents matching a filter
func (r *MongoParserRepository) FindByFilter(collectionName string, filter bson.M, limit int) ([]map[string]any, error) {
	collection := r.client.GetCollection(collectionName)
	ctx := context.Background()

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents in %s: %w", collectionName, err)
	}
	defer cursor.Close(ctx)

	var results []map[string]any
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode documents from %s: %w", collectionName, err)
	}

	return results, nil
}

// GetCollections returns a list of all collections in the database
func (r *MongoParserRepository) GetCollections() ([]string, error) {
	ctx := context.Background()

	names, err := r.client.GetDatabase().ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	return names, nil
}
