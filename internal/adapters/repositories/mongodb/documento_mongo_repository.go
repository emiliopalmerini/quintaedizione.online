package mongodb

import (
	"context"
	"fmt"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain/repositories"
	"github.com/emiliopalmerini/due-draghi-5e-srd/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DocumentoMongoRepository implements DocumentoRepository for MongoDB
type DocumentoMongoRepository struct {
	*BaseMongoRepository[*domain.Documento]
}

// NewDocumentoMongoRepository creates a new DocumentoMongoRepository
func NewDocumentoMongoRepository(client *mongodb.Client) repositories.DocumentoRepository {
	base := NewBaseMongoRepository[*domain.Documento](
		client,
		"documenti",
		[]string{"slug"},
	)

	return &DocumentoMongoRepository{
		BaseMongoRepository: base,
	}
}

// FindBySlug retrieves a document by its slug
func (r *DocumentoMongoRepository) FindBySlug(ctx context.Context, slug string) (*domain.Documento, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{"slug": slug}

	var documento domain.Documento
	err := collection.FindOne(ctx, filter).Decode(&documento)
	if err != nil {
		return nil, fmt.Errorf("failed to find documento by slug %s: %w", slug, err)
	}

	return &documento, nil
}

// FindByTitle searches documents by title (partial match)
func (r *DocumentoMongoRepository) FindByTitle(ctx context.Context, title string, limit int) ([]*domain.Documento, error) {
	collection := r.client.GetCollection(r.collectionName)

	// Use regex for partial matching
	filter := bson.M{
		"$or": []bson.M{
			{"titolo": primitive.Regex{Pattern: title, Options: "i"}},
			{"title": primitive.Regex{Pattern: title, Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find documenti by title %s: %w", title, err)
	}
	defer cursor.Close(ctx)

	var documenti []*domain.Documento
	for cursor.Next(ctx) {
		var documento domain.Documento
		if err := cursor.Decode(&documento); err != nil {
			return nil, fmt.Errorf("failed to decode documento: %w", err)
		}
		documenti = append(documenti, &documento)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return documenti, nil
}

// FindByContent searches documents by content (text search)
func (r *DocumentoMongoRepository) FindByContent(ctx context.Context, searchText string, limit int) ([]*domain.Documento, error) {
	collection := r.client.GetCollection(r.collectionName)

	// Use text search or regex on content fields
	filter := bson.M{
		"$or": []bson.M{
			{"contenuto.markdown": primitive.Regex{Pattern: searchText, Options: "i"}},
			{"content.markdown": primitive.Regex{Pattern: searchText, Options: "i"}},
			{"contenuto.html": primitive.Regex{Pattern: searchText, Options: "i"}},
			{"content.html": primitive.Regex{Pattern: searchText, Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to search documenti by content: %w", err)
	}
	defer cursor.Close(ctx)

	var documenti []*domain.Documento
	for cursor.Next(ctx) {
		var documento domain.Documento
		if err := cursor.Decode(&documento); err != nil {
			return nil, fmt.Errorf("failed to decode documento: %w", err)
		}
		documenti = append(documenti, &documento)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return documenti, nil
}
