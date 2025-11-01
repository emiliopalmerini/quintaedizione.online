package mongodb

import (
	"context"
	"fmt"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain/repositories"
	pkgMongodb "github.com/emiliopalmerini/due-draghi-5e-srd/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type documentMongoRepository struct {
	client *pkgMongodb.Client
}

// NewDocumentMongoRepository creates a MongoDB implementation of DocumentRepository
func NewDocumentMongoRepository(client *pkgMongodb.Client) repositories.DocumentRepository {
	return &documentMongoRepository{
		client: client,
	}
}

func (r *documentMongoRepository) getCollection(collection string) *mongo.Collection {
	return r.client.Database().Collection(collection)
}

func (r *documentMongoRepository) Create(ctx context.Context, doc *domain.Document, collection string) error {
	coll := r.getCollection(collection)
	_, err := coll.InsertOne(ctx, doc)
	return err
}

func (r *documentMongoRepository) Update(ctx context.Context, doc *domain.Document, collection string) error {
	coll := r.getCollection(collection)
	filter := bson.M{"_id": doc.ID}
	_, err := coll.ReplaceOne(ctx, filter, doc)
	return err
}

func (r *documentMongoRepository) Delete(ctx context.Context, id domain.DocumentID, collection string) error {
	coll := r.getCollection(collection)
	filter := bson.M{"_id": id}
	_, err := coll.DeleteOne(ctx, filter)
	return err
}

func (r *documentMongoRepository) FindByID(ctx context.Context, id domain.DocumentID, collection string) (*domain.Document, error) {
	coll := r.getCollection(collection)
	filter := bson.M{"_id": id}

	var doc domain.Document
	err := coll.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("document not found: %s", id)
		}
		return nil, err
	}
	return &doc, nil
}

func (r *documentMongoRepository) FindAll(ctx context.Context, collection string, limit int) ([]*domain.Document, error) {
	coll := r.getCollection(collection)
	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	// Sort by title for consistent ordering
	opts.SetSort(bson.D{{Key: "title", Value: 1}})

	cursor, err := coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []*domain.Document
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func (r *documentMongoRepository) FindByFilters(ctx context.Context, collection string, filters map[string]any, limit int) ([]*domain.Document, error) {
	coll := r.getCollection(collection)

	// Build filter query - search within the filters field
	filter := bson.M{}
	for key, value := range filters {
		filter[fmt.Sprintf("filters.%s", key)] = value
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	opts.SetSort(bson.D{{Key: "title", Value: 1}})

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []*domain.Document
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func (r *documentMongoRepository) Count(ctx context.Context, collection string) (int64, error) {
	coll := r.getCollection(collection)
	return coll.CountDocuments(ctx, bson.M{})
}

func (r *documentMongoRepository) UpsertMany(ctx context.Context, collection string, documents []*domain.Document) (int, error) {
	if len(documents) == 0 {
		return 0, nil
	}

	coll := r.getCollection(collection)

	var models []mongo.WriteModel
	for _, doc := range documents {
		model := mongo.NewReplaceOneModel().
			SetFilter(bson.M{"_id": doc.ID}).
			SetReplacement(doc).
			SetUpsert(true)
		models = append(models, model)
	}

	result, err := coll.BulkWrite(ctx, models)
	if err != nil {
		return 0, err
	}

	return int(result.UpsertedCount + result.ModifiedCount), nil
}

func (r *documentMongoRepository) UpsertManyMaps(ctx context.Context, collection string, uniqueFields []string, docs []map[string]any) (int, error) {
	if len(docs) == 0 {
		return 0, nil
	}

	coll := r.getCollection(collection)

	var models []mongo.WriteModel
	for _, doc := range docs {
		// Build filter based on unique fields
		filter := bson.M{}
		for _, field := range uniqueFields {
			if val, ok := doc[field]; ok {
				filter[field] = val
			}
		}

		// If no unique fields specified, use _id
		if len(filter) == 0 {
			if id, ok := doc["_id"]; ok {
				filter["_id"] = id
			}
		}

		model := mongo.NewReplaceOneModel().
			SetFilter(filter).
			SetReplacement(doc).
			SetUpsert(true)
		models = append(models, model)
	}

	result, err := coll.BulkWrite(ctx, models)
	if err != nil {
		return 0, err
	}

	return int(result.UpsertedCount + result.ModifiedCount), nil
}

func (r *documentMongoRepository) GetCollectionStats(ctx context.Context, collection string) (map[string]any, error) {
	count, err := r.Count(ctx, collection)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"collection": collection,
		"count":      count,
	}, nil
}

func (r *documentMongoRepository) GetAdjacentDocuments(ctx context.Context, collection string, currentID domain.DocumentID) (prev *domain.Document, next *domain.Document, err error) {
	coll := r.getCollection(collection)

	// Get current document to find its position
	current, err := r.FindByID(ctx, currentID, collection)
	if err != nil {
		return nil, nil, err
	}

	// Find previous document (title < current.Title, sorted descending, limit 1)
	prevFilter := bson.M{"title": bson.M{"$lt": current.Title}}
	prevOpts := options.FindOne().SetSort(bson.D{{Key: "title", Value: -1}})
	var prevDoc domain.Document
	err = coll.FindOne(ctx, prevFilter, prevOpts).Decode(&prevDoc)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, nil, err
	}
	if err == nil {
		prev = &prevDoc
	}

	// Find next document (title > current.Title, sorted ascending, limit 1)
	nextFilter := bson.M{"title": bson.M{"$gt": current.Title}}
	nextOpts := options.FindOne().SetSort(bson.D{{Key: "title", Value: 1}})
	var nextDoc domain.Document
	err = coll.FindOne(ctx, nextFilter, nextOpts).Decode(&nextDoc)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, nil, err
	}
	if err == nil {
		next = &nextDoc
	}

	return prev, next, nil
}
