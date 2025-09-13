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

// RegolaMongoRepository implements repositories.RegolaRepository using MongoDB
type RegolaMongoRepository struct {
	*BaseMongoRepository[*domain.Regola]
}

// NewRegolaMongoRepository creates a new MongoDB Regola repository
func NewRegolaMongoRepository(client *mongodb.Client) repositories.RegolaRepository {
	base := NewBaseMongoRepository[*domain.Regola](
		client,
		"regole",
		[]string{"nome", "slug"},
	)
	
	return &RegolaMongoRepository{
		BaseMongoRepository: base,
	}
}

// extractRegolaFromDocument extracts Regola from the flattened document
func extractRegolaFromDocument(doc bson.M) (*domain.Regola, error) {
	docBytes, err := bson.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal document: %w", err)
	}

	var regola domain.Regola
	err = bson.Unmarshal(docBytes, &regola)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal regola: %w", err)
	}

	return &regola, nil
}

// FindByNome retrieves a rule by its name
func (r *RegolaMongoRepository) FindByNome(ctx context.Context, nome string) (*domain.Regola, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{"nome": nome}

	var doc bson.M
	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("failed to find regola by nome %s: %w", nome, err)
	}

	return extractRegolaFromDocument(doc)
}

// FindByCategory retrieves rules by category
func (r *RegolaMongoRepository) FindByCategory(ctx context.Context, category string, limit int) ([]*domain.Regola, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"categoria": primitive.Regex{Pattern: category, Options: "i"}},
			{"category": primitive.Regex{Pattern: category, Options: "i"}},
			{"tipo": primitive.Regex{Pattern: category, Options: "i"}},
			{"type": primitive.Regex{Pattern: category, Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find regole by category %s: %w", category, err)
	}
	defer cursor.Close(ctx)

	var regole []*domain.Regola
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		regola, err := extractRegolaFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract regola: %w", err)
		}
		regole = append(regole, regola)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return regole, nil
}

// FindByKeyword searches rules by keyword in content
func (r *RegolaMongoRepository) FindByKeyword(ctx context.Context, keyword string, limit int) ([]*domain.Regola, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"descrizione": primitive.Regex{Pattern: keyword, Options: "i"}},
			{"description": primitive.Regex{Pattern: keyword, Options: "i"}},
			{"contenuto": primitive.Regex{Pattern: keyword, Options: "i"}},
			{"nome": primitive.Regex{Pattern: keyword, Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find regole by keyword %s: %w", keyword, err)
	}
	defer cursor.Close(ctx)

	var regole []*domain.Regola
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		regola, err := extractRegolaFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract regola: %w", err)
		}
		regole = append(regole, regola)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return regole, nil
}