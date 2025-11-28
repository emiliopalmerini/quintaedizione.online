package mongodb

import (
	"context"
	"fmt"
	"maps"

	"github.com/emiliopalmerini/quintaedizione.online/internal/domain"
	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/repositories"
	pkgMongodb "github.com/emiliopalmerini/quintaedizione.online/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type documentMongoRepository struct {
	client *pkgMongodb.Client
}

func NewDocumentMongoRepository(client *pkgMongodb.Client) repositories.DocumentRepository {
	return &documentMongoRepository{
		client: client,
	}
}

func (r *documentMongoRepository) getCollection(collection string) *mongo.Collection {
	return r.client.GetDatabase().Collection(collection)
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

		filter := bson.M{}
		for _, field := range uniqueFields {
			if val, ok := doc[field]; ok {
				filter[field] = val
			}
		}

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

	current, err := r.FindByID(ctx, currentID, collection)
	if err != nil {
		return nil, nil, err
	}

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

func (r *documentMongoRepository) FindMapByID(ctx context.Context, collection string, id string) (map[string]any, error) {
	coll := r.getCollection(collection)
	filter := bson.M{"_id": id}

	var result map[string]any
	err := coll.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to find document with id %s in collection %s: %w", id, collection, err)
	}

	return result, nil
}

func (r *documentMongoRepository) FindMaps(ctx context.Context, collection string, filter map[string]any, skip int64, limit int64) ([]map[string]any, int64, error) {
	coll := r.getCollection(collection)

	totalCount, err := r.CountWithFilter(ctx, collection, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	mongoFilter := bson.M{}
	maps.Copy(mongoFilter, filter)

	opts := options.Find()
	if skip > 0 {
		opts.SetSkip(skip)
	}
	if limit > 0 {
		opts.SetLimit(limit)
	}

	if _, hasTextSearch := filter["$text"]; hasTextSearch {
		opts.SetProjection(bson.M{"score": bson.M{"$meta": "textScore"}})
		opts.SetSort(bson.D{{Key: "score", Value: bson.M{"$meta": "textScore"}}})
	} else {
		opts.SetSort(bson.D{{Key: "title", Value: 1}})
	}

	cursor, err := coll.Find(ctx, mongoFilter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	var results []map[string]any
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, fmt.Errorf("failed to decode documents: %w", err)
	}

	return results, totalCount, nil
}

func (r *documentMongoRepository) CountWithFilter(ctx context.Context, collection string, filter map[string]any) (int64, error) {
	coll := r.getCollection(collection)

	mongoFilter := bson.M{}
	for key, value := range filter {
		mongoFilter[key] = value
	}

	count, err := coll.CountDocuments(ctx, mongoFilter)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents in %s: %w", collection, err)
	}

	return count, nil
}

func (r *documentMongoRepository) GetAdjacentMaps(ctx context.Context, collection string, currentID string) (prevID *string, nextID *string, err error) {
	coll := r.getCollection(collection)

	var current map[string]any
	err = coll.FindOne(ctx, bson.M{"_id": currentID}).Decode(&current)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find current document: %w", err)
	}

	currentTitle, ok := current["title"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("current document missing title field")
	}

	prevFilter := bson.M{"title": bson.M{"$lt": currentTitle}}
	prevOpts := options.FindOne().SetSort(bson.D{{Key: "title", Value: -1}})
	var prevDoc map[string]any
	err = coll.FindOne(ctx, prevFilter, prevOpts).Decode(&prevDoc)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, nil, err
	}
	if err == nil {
		if id, ok := prevDoc["_id"].(string); ok {
			prevID = &id
		}
	}

	nextFilter := bson.M{"title": bson.M{"$gt": currentTitle}}
	nextOpts := options.FindOne().SetSort(bson.D{{Key: "title", Value: 1}})
	var nextDoc map[string]any
	err = coll.FindOne(ctx, nextFilter, nextOpts).Decode(&nextDoc)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, nil, err
	}
	if err == nil {
		if id, ok := nextDoc["_id"].(string); ok {
			nextID = &id
		}
	}

	return prevID, nextID, nil
}

func (r *documentMongoRepository) GetAllCollectionStats(ctx context.Context) ([]map[string]any, error) {

	validCollections := []string{
		"incantesimi", "mostri", "classi", "backgrounds",
		"equipaggiamenti", "oggetti_magici", "armi", "armature",
		"talenti", "servizi", "strumenti", "animali",
		"regole", "cavalcature_veicoli",
	}

	stats := make([]map[string]any, 0, len(validCollections))

	for _, collection := range validCollections {
		count, err := r.Count(ctx, collection)
		if err != nil {

			continue
		}

		stats = append(stats, map[string]any{
			"collection": collection,
			"count":      count,
		})
	}

	return stats, nil
}

func (r *documentMongoRepository) DropCollection(ctx context.Context, collection string) error {
	coll := r.getCollection(collection)
	return coll.Drop(ctx)
}
