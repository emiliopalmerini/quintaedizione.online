package mongodb

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain/repositories"
	"github.com/emiliopalmerini/due-draghi-5e-srd/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BaseMongoRepository provides common MongoDB operations for any entity type
type BaseMongoRepository[T repositories.Entity] struct {
	client         *mongodb.Client
	collectionName string
	uniqueFields   []string
}

// NewBaseMongoRepository creates a new base MongoDB repository
func NewBaseMongoRepository[T repositories.Entity](client *mongodb.Client, collectionName string, uniqueFields []string) *BaseMongoRepository[T] {
	return &BaseMongoRepository[T]{
		client:         client,
		collectionName: collectionName,
		uniqueFields:   uniqueFields,
	}
}

// Create inserts a new entity
func (r *BaseMongoRepository[T]) Create(ctx context.Context, entity T) error {
	collection := r.client.GetCollection(r.collectionName)
	_, err := collection.InsertOne(ctx, entity)
	if err != nil {
		return fmt.Errorf("failed to create entity in %s: %w", r.collectionName, err)
	}
	return nil
}

// Update modifies an existing entity
func (r *BaseMongoRepository[T]) Update(ctx context.Context, entity T) error {
	collection := r.client.GetCollection(r.collectionName)

	// Build filter using unique fields
	filter := r.buildFilterFromEntity(entity)
	if len(filter) == 0 {
		return fmt.Errorf("no unique fields found to update entity in %s", r.collectionName)
	}

	result, err := collection.ReplaceOne(ctx, filter, entity)
	if err != nil {
		return fmt.Errorf("failed to update entity in %s: %w", r.collectionName, err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("entity not found in %s", r.collectionName)
	}

	return nil
}

// Delete removes an entity by its identifier
func (r *BaseMongoRepository[T]) Delete(ctx context.Context, id string) error {
	collection := r.client.GetCollection(r.collectionName)

	// Try multiple possible ID fields
	idFilters := []bson.M{
		{"_id": id},
		{"id": id},
		{"slug": id},
		{"nome": id},
	}

	for _, filter := range idFilters {
		result, err := collection.DeleteOne(ctx, filter)
		if err != nil {
			continue
		}
		if result.DeletedCount > 0 {
			return nil
		}
	}

	return fmt.Errorf("entity with id %s not found in %s", id, r.collectionName)
}

// FindByID retrieves an entity by its identifier
func (r *BaseMongoRepository[T]) FindByID(ctx context.Context, id string) (*T, error) {
	collection := r.client.GetCollection(r.collectionName)

	// Try multiple possible ID fields
	idFilters := []bson.M{
		{"_id": id},
		{"id": id},
		{"slug": id},
		{"nome": id},
	}

	var entity T
	for _, filter := range idFilters {
		err := collection.FindOne(ctx, filter).Decode(&entity)
		if err == nil {
			return &entity, nil
		}
		if err != mongo.ErrNoDocuments {
			return nil, fmt.Errorf("failed to find entity by id %s in %s: %w", id, r.collectionName, err)
		}
	}

	return nil, fmt.Errorf("entity with id %s not found in %s", id, r.collectionName)
}

// FindAll retrieves all entities with optional limit
func (r *BaseMongoRepository[T]) FindAll(ctx context.Context, limit int) ([]*T, error) {
	collection := r.client.GetCollection(r.collectionName)

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find entities in %s: %w", r.collectionName, err)
	}
	defer cursor.Close(ctx)

	var entities []*T
	for cursor.Next(ctx) {
		var entity T
		if err := cursor.Decode(&entity); err != nil {
			return nil, fmt.Errorf("failed to decode entity from %s: %w", r.collectionName, err)
		}
		entities = append(entities, &entity)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error in %s: %w", r.collectionName, err)
	}

	return entities, nil
}

// FindByFilter retrieves entities matching the filter
func (r *BaseMongoRepository[T]) FindByFilter(ctx context.Context, filter bson.M, limit int) ([]*T, error) {
	collection := r.client.GetCollection(r.collectionName)

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find entities in %s: %w", r.collectionName, err)
	}
	defer cursor.Close(ctx)

	var entities []*T
	for cursor.Next(ctx) {
		var entity T
		if err := cursor.Decode(&entity); err != nil {
			return nil, fmt.Errorf("failed to decode entity from %s: %w", r.collectionName, err)
		}
		entities = append(entities, &entity)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error in %s: %w", r.collectionName, err)
	}

	return entities, nil
}

// Count returns the total number of entities
func (r *BaseMongoRepository[T]) Count(ctx context.Context) (int64, error) {
	collection := r.client.GetCollection(r.collectionName)

	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("failed to count documents in %s: %w", r.collectionName, err)
	}

	return count, nil
}

// UpsertMany performs bulk upsert operations
func (r *BaseMongoRepository[T]) UpsertMany(ctx context.Context, entities []T) (int, error) {
	if len(entities) == 0 {
		return 0, nil
	}

	collection := r.client.GetCollection(r.collectionName)
	var operations []mongo.WriteModel

	for _, entity := range entities {
		filter := r.buildFilterFromEntity(entity)
		if len(filter) == 0 {
			continue // Skip entities without unique fields
		}

		operation := mongo.NewReplaceOneModel().
			SetFilter(filter).
			SetReplacement(entity).
			SetUpsert(true)

		operations = append(operations, operation)
	}

	if len(operations) == 0 {
		return 0, fmt.Errorf("no valid entities to upsert in %s", r.collectionName)
	}

	opts := options.BulkWrite().SetOrdered(false)
	result, err := collection.BulkWrite(ctx, operations, opts)
	if err != nil {
		return 0, fmt.Errorf("bulk upsert failed in %s: %w", r.collectionName, err)
	}

	return int(result.UpsertedCount + result.ModifiedCount), nil
}

// UpsertManyMaps performs bulk upsert operations from raw maps (for parser compatibility)
func (r *BaseMongoRepository[T]) UpsertManyMaps(ctx context.Context, uniqueFields []string, docs []map[string]any) (int, error) {
	if len(docs) == 0 {
		return 0, nil
	}

	// Use provided unique fields or fallback to repository defaults
	fields := uniqueFields
	if len(fields) == 0 {
		fields = r.uniqueFields
	}


	collection := r.client.GetCollection(r.collectionName)
	var operations []mongo.WriteModel

	for i, doc := range docs {
		filter := bson.M{}
		hasUniqueField := false

		for _, field := range fields {
			// Handle nested field paths (e.g., "value.nome")
			value := r.getNestedValue(doc, field)
			if value != nil && value != "" {
				filter[field] = value
				hasUniqueField = true
				break // Use first available unique field
			}
		}

		if !hasUniqueField {
			continue // Skip documents without unique fields
		}

		operation := mongo.NewReplaceOneModel().
			SetFilter(filter).
			SetReplacement(doc).
			SetUpsert(true)

		operations = append(operations, operation)
	}

	if len(operations) == 0 {
		return 0, fmt.Errorf("no valid documents to upsert in %s", r.collectionName)
	}

	opts := options.BulkWrite().SetOrdered(false)
	result, err := collection.BulkWrite(ctx, operations, opts)
	if err != nil {
		return 0, fmt.Errorf("bulk upsert failed in %s: %w", r.collectionName, err)
	}

	return int(result.UpsertedCount + result.ModifiedCount), nil
}

// buildFilterFromEntity creates a filter from entity using unique fields
func (r *BaseMongoRepository[T]) buildFilterFromEntity(entity T) bson.M {
	filter := bson.M{}

	// Use reflection to extract field values
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return filter
	}

	_ = v.Type()

	for _, fieldName := range r.uniqueFields {
		// Try multiple field name variations
		variations := []string{
			fieldName,
			strings.Title(fieldName),
			strings.ToUpper(fieldName[:1]) + fieldName[1:],
		}

		for _, variation := range variations {
			if field := v.FieldByName(variation); field.IsValid() && !field.IsZero() {
				// Convert field value to interface
				if value := field.Interface(); value != nil {
					filter[strings.ToLower(fieldName)] = value
					break
				}
			}
		}

		if len(filter) > 0 {
			break // Use first available unique field
		}
	}

	return filter
}

// GetCollectionName returns the collection name for this repository
func (r *BaseMongoRepository[T]) GetCollectionName() string {
	return r.collectionName
}

// getNestedValue extracts a value from a nested map using dot notation (e.g., "value.nome")
func (r *BaseMongoRepository[T]) getNestedValue(doc map[string]any, fieldPath string) any {
	parts := strings.Split(fieldPath, ".")
	current := doc

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part, return the value
			return current[part]
		}
		
		// Navigate deeper into the structure
		if next, exists := current[part]; exists {
			if nextMap, ok := next.(map[string]any); ok {
				current = nextMap
			} else {
				return nil // Path doesn't exist or isn't a map
			}
		} else {
			return nil // Path doesn't exist
		}
	}
	
	return nil
}

// getMapKeys returns all keys from a map for debugging
func getMapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
