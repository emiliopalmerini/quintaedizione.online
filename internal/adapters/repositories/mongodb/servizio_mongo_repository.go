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

// ServizioMongoRepository implements repositories.ServizioRepository using MongoDB
type ServizioMongoRepository struct {
	*BaseMongoRepository[*domain.Servizio]
}

// NewServizioMongoRepository creates a new MongoDB Servizio repository
func NewServizioMongoRepository(client *mongodb.Client) repositories.ServizioRepository {
	base := NewBaseMongoRepository[*domain.Servizio](
		client,
		"servizi",
		[]string{"value.nome", "value.slug"},
	)
	
	return &ServizioMongoRepository{
		BaseMongoRepository: base,
	}
}

// extractServizioFromDocument extracts Servizio from the nested value field
func extractServizioFromDocument(doc bson.M) (*domain.Servizio, error) {
	return ExtractEntityFromDocument[domain.Servizio](doc, true)
}

// FindByNome retrieves a service by its name
func (r *ServizioMongoRepository) FindByNome(ctx context.Context, nome string) (*domain.Servizio, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{"value.nome": nome}

	var doc bson.M
	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("failed to find servizio by nome %s: %w", nome, err)
	}

	return extractServizioFromDocument(doc)
}

// FindByCategory retrieves services by category
func (r *ServizioMongoRepository) FindByCategory(ctx context.Context, category string, limit int) ([]*domain.Servizio, error) {
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
		return nil, fmt.Errorf("failed to find servizi by category %s: %w", category, err)
	}
	defer cursor.Close(ctx)

	var servizi []*domain.Servizio
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		servizio, err := extractServizioFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract servizio: %w", err)
		}
		servizi = append(servizi, servizio)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return servizi, nil
}

// FindByPriceRange retrieves services within price range
func (r *ServizioMongoRepository) FindByPriceRange(ctx context.Context, minPrice, maxPrice float64, limit int) ([]*domain.Servizio, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{
				"value.costo.valore": bson.M{
					"$gte": minPrice,
					"$lte": maxPrice,
				},
			},
			{
				"value.cost.value": bson.M{
					"$gte": minPrice,
					"$lte": maxPrice,
				},
			},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find servizi by price range %.2f-%.2f: %w", minPrice, maxPrice, err)
	}
	defer cursor.Close(ctx)

	var servizi []*domain.Servizio
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		servizio, err := extractServizioFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract servizio: %w", err)
		}
		servizi = append(servizi, servizio)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return servizi, nil
}