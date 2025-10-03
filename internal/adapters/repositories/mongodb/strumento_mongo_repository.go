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

// StrumentoMongoRepository implements repositories.StrumentoRepository using MongoDB
type StrumentoMongoRepository struct {
	*BaseMongoRepository[*domain.Strumento]
}

// NewStrumentoMongoRepository creates a new MongoDB Strumento repository
func NewStrumentoMongoRepository(client *mongodb.Client) repositories.StrumentoRepository {
	base := NewBaseMongoRepository[*domain.Strumento](
		client,
		"strumenti",
		[]string{"value.nome", "value.slug"},
	)
	
	return &StrumentoMongoRepository{
		BaseMongoRepository: base,
	}
}

// extractStrumentoFromDocument extracts Strumento from the nested value field
func extractStrumentoFromDocument(doc bson.M) (*domain.Strumento, error) {
	return ExtractEntityFromDocument[domain.Strumento](doc, true)
}

// FindByNome retrieves a tool by its name
func (r *StrumentoMongoRepository) FindByNome(ctx context.Context, nome string) (*domain.Strumento, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{"value.nome": nome}

	var doc bson.M
	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("failed to find strumento by nome %s: %w", nome, err)
	}

	return extractStrumentoFromDocument(doc)
}

// FindByCategory retrieves tools by category
func (r *StrumentoMongoRepository) FindByCategory(ctx context.Context, category string, limit int) ([]*domain.Strumento, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"value.categoria": primitive.Regex{Pattern: category, Options: "i"}},
			{"value.category": primitive.Regex{Pattern: category, Options: "i"}},
			{"value.tipo": primitive.Regex{Pattern: category, Options: "i"}},
			{"value.type": primitive.Regex{Pattern: category, Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find strumenti by category %s: %w", category, err)
	}
	defer cursor.Close(ctx)

	var strumenti []*domain.Strumento
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		strumento, err := extractStrumentoFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract strumento: %w", err)
		}
		strumenti = append(strumenti, strumento)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return strumenti, nil
}

// FindByUse retrieves tools by their usage type
func (r *StrumentoMongoRepository) FindByUse(ctx context.Context, useType string, limit int) ([]*domain.Strumento, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"value.uso": primitive.Regex{Pattern: useType, Options: "i"}},
			{"value.use": primitive.Regex{Pattern: useType, Options: "i"}},
			{"value.descrizione": primitive.Regex{Pattern: useType, Options: "i"}},
			{"contenuto": primitive.Regex{Pattern: useType, Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find strumenti by use type %s: %w", useType, err)
	}
	defer cursor.Close(ctx)

	var strumenti []*domain.Strumento
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		strumento, err := extractStrumentoFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract strumento: %w", err)
		}
		strumenti = append(strumenti, strumento)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return strumenti, nil
}