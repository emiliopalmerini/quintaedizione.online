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

// ArmaMongoRepository implements ArmaRepository for MongoDB
type ArmaMongoRepository struct {
	*BaseMongoRepository[*domain.Arma]
}

// extractArmaFromDocument extracts Arma from the nested value field
func extractArmaFromDocument(doc bson.M) (*domain.Arma, error) {
	return ExtractEntityFromDocument[domain.Arma](doc, true)
}

// NewArmaMongoRepository creates a new ArmaMongoRepository
func NewArmaMongoRepository(client *mongodb.Client) repositories.ArmaRepository {
	base := NewBaseMongoRepository[*domain.Arma](
		client,
		"armi",
		[]string{"value.nome", "value.slug"},
	)

	return &ArmaMongoRepository{
		BaseMongoRepository: base,
	}
}

// FindByNome retrieves a weapon by its name
func (r *ArmaMongoRepository) FindByNome(ctx context.Context, nome string) (*domain.Arma, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{"value.nome": nome}

	var doc bson.M
	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("failed to find arma by nome %s: %w", nome, err)
	}

	return extractArmaFromDocument(doc)
}

// FindByCategory retrieves weapons by category (Semplici, Militari, etc.)
func (r *ArmaMongoRepository) FindByCategory(ctx context.Context, category string, limit int) ([]*domain.Arma, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"value.categoria": primitive.Regex{Pattern: category, Options: "i"}},
			{"value.category": primitive.Regex{Pattern: category, Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find armi by category %s: %w", category, err)
	}
	defer cursor.Close(ctx)

	var armi []*domain.Arma
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		arma, err := extractArmaFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract arma: %w", err)
		}
		armi = append(armi, arma)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return armi, nil
}

// FindByDamageType retrieves weapons by damage type
func (r *ArmaMongoRepository) FindByDamageType(ctx context.Context, damageType string, limit int) ([]*domain.Arma, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"value.danno": primitive.Regex{Pattern: damageType, Options: "i"}},
			{"value.damage": primitive.Regex{Pattern: damageType, Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find armi by damage type %s: %w", damageType, err)
	}
	defer cursor.Close(ctx)

	var armi []*domain.Arma
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		arma, err := extractArmaFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract arma: %w", err)
		}
		armi = append(armi, arma)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return armi, nil
}

// FindRangedWeapons retrieves ranged weapons
func (r *ArmaMongoRepository) FindRangedWeapons(ctx context.Context, limit int) ([]*domain.Arma, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"value.categoria": primitive.Regex{Pattern: "Distanza", Options: "i"}},
			{"value.category": primitive.Regex{Pattern: "Ranged", Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find ranged weapons: %w", err)
	}
	defer cursor.Close(ctx)

	var armi []*domain.Arma
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		arma, err := extractArmaFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract arma: %w", err)
		}
		armi = append(armi, arma)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return armi, nil
}

// FindMeleeWeapons retrieves melee weapons
func (r *ArmaMongoRepository) FindMeleeWeapons(ctx context.Context, limit int) ([]*domain.Arma, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"value.categoria": primitive.Regex{Pattern: "Mischia", Options: "i"}},
			{"value.category": primitive.Regex{Pattern: "Melee", Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find melee weapons: %w", err)
	}
	defer cursor.Close(ctx)

	var armi []*domain.Arma
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		arma, err := extractArmaFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract arma: %w", err)
		}
		armi = append(armi, arma)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return armi, nil
}

// FindByProperty retrieves weapons with specific properties (Finezza, Pesante, etc.)
func (r *ArmaMongoRepository) FindByProperty(ctx context.Context, property string, limit int) ([]*domain.Arma, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"value.proprieta": primitive.Regex{Pattern: property, Options: "i"}},
			{"value.properties": primitive.Regex{Pattern: property, Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find armi by property %s: %w", property, err)
	}
	defer cursor.Close(ctx)

	var armi []*domain.Arma
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		arma, err := extractArmaFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract arma: %w", err)
		}
		armi = append(armi, arma)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return armi, nil
}