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

// RegolaMongoRepository implements RegolaRepository for MongoDB
type RegolaMongoRepository struct {
	*BaseMongoRepository[*domain.Regola]
}

// extractRegolaFromDocument extracts Regola from the nested value field
func extractRegolaFromDocument(doc bson.M) (*domain.Regola, error) {
	valueData, exists := doc["value"]
	if !exists {
		return nil, fmt.Errorf("regola document missing value field")
	}

	valueBytes, err := bson.Marshal(valueData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value data: %w", err)
	}

	var regola domain.Regola
	err = bson.Unmarshal(valueBytes, &regola)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal regola: %w", err)
	}

	return &regola, nil
}

// NewRegolaMongoRepository creates a new RegolaMongoRepository
func NewRegolaMongoRepository(client *mongodb.Client) repositories.RegolaRepository {
	base := NewBaseMongoRepository[*domain.Regola](
		client,
		"regole",
		[]string{"value.nome", "value.slug"},
	)

	return &RegolaMongoRepository{
		BaseMongoRepository: base,
	}
}

// FindByNome retrieves a rule by its name
func (r *RegolaMongoRepository) FindByNome(ctx context.Context, nome string) (*domain.Regola, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{"value.nome": nome}

	var doc bson.M
	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("failed to find regola by nome %s: %w", nome, err)
	}

	return extractRegolaFromDocument(doc)
}

// FindByCategory retrieves rules by category (extracted from tags in the name)
func (r *RegolaMongoRepository) FindByCategory(ctx context.Context, category string, limit int) ([]*domain.Regola, error) {
	collection := r.client.GetCollection(r.collectionName)

	// Search for rules with category tags like [Azione], [Condizione], etc.
	filter := bson.M{
		"$or": []bson.M{
			{"value.nome": primitive.Regex{Pattern: "\\[" + category + "\\]", Options: "i"}},
			{"contenuto": primitive.Regex{Pattern: category, Options: "i"}},
			{"value.contenuto": primitive.Regex{Pattern: category, Options: "i"}},
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
			{"value.nome": primitive.Regex{Pattern: keyword, Options: "i"}},
			{"value.contenuto": primitive.Regex{Pattern: keyword, Options: "i"}},
			{"contenuto": primitive.Regex{Pattern: keyword, Options: "i"}},
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