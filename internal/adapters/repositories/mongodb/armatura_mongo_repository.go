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

// ArmaturaMongoRepository implements ArmaturaRepository for MongoDB
type ArmaturaMongoRepository struct {
	*BaseMongoRepository[*domain.Armatura]
}

// extractArmaturaFromDocument extracts Armatura from the nested value field
func extractArmaturaFromDocument(doc bson.M) (*domain.Armatura, error) {
	return ExtractEntityFromDocument[domain.Armatura](doc, true)
}

// NewArmaturaMongoRepository creates a new ArmaturaMongoRepository
func NewArmaturaMongoRepository(client *mongodb.Client) repositories.ArmaturaRepository {
	base := NewBaseMongoRepository[*domain.Armatura](
		client,
		"armature",
		[]string{"value.nome", "value.slug"},
	)

	return &ArmaturaMongoRepository{
		BaseMongoRepository: base,
	}
}

// FindByNome retrieves an armor by its name
func (r *ArmaturaMongoRepository) FindByNome(ctx context.Context, nome string) (*domain.Armatura, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{"value.nome": nome}

	var doc bson.M
	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("failed to find armatura by nome %s: %w", nome, err)
	}

	return extractArmaturaFromDocument(doc)
}

// FindByType retrieves armor by type (e.g., leggera, media, pesante)
func (r *ArmaturaMongoRepository) FindByType(ctx context.Context, armorType string, limit int) ([]*domain.Armatura, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"value.tipo": primitive.Regex{Pattern: armorType, Options: "i"}},
			{"value.type": primitive.Regex{Pattern: armorType, Options: "i"}},
			{"value.categoria": primitive.Regex{Pattern: armorType, Options: "i"}},
			{"value.category": primitive.Regex{Pattern: armorType, Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find armature by type %s: %w", armorType, err)
	}
	defer cursor.Close(ctx)

	var armature []*domain.Armatura
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		armatura, err := extractArmaturaFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract armatura: %w", err)
		}
		armature = append(armature, armatura)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return armature, nil
}

// FindByACRange retrieves armor within an AC range
func (r *ArmaturaMongoRepository) FindByACRange(ctx context.Context, minAC, maxAC int, limit int) ([]*domain.Armatura, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"value.classe_armatura.base": bson.M{"$gte": minAC, "$lte": maxAC}},
			{"value.ca.base": bson.M{"$gte": minAC, "$lte": maxAC}},
			{"value.ac.base": bson.M{"$gte": minAC, "$lte": maxAC}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find armature by AC range %d-%d: %w", minAC, maxAC, err)
	}
	defer cursor.Close(ctx)

	var armature []*domain.Armatura
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		armatura, err := extractArmaturaFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract armatura: %w", err)
		}
		armature = append(armature, armatura)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return armature, nil
}

// FindStealthDisadvantage retrieves armor that imposes stealth disadvantage
func (r *ArmaturaMongoRepository) FindStealthDisadvantage(ctx context.Context, limit int) ([]*domain.Armatura, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"value.svantaggio_furtivita": true},
			{"value.svantaggioFurtivita": true},
			{"value.stealth_disadvantage": true},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find armature by stealth disadvantage: %w", err)
	}
	defer cursor.Close(ctx)

	var armature []*domain.Armatura
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		armatura, err := extractArmaturaFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract armatura: %w", err)
		}
		armature = append(armature, armatura)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return armature, nil
}

// FindByStrengthRequirement retrieves armor by strength requirement
func (r *ArmaturaMongoRepository) FindByStrengthRequirement(ctx context.Context, minStr int, limit int) ([]*domain.Armatura, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"value.forza_richiesta": bson.M{"$gte": minStr}},
			{"value.requisitiForza": bson.M{"$gte": minStr}},
			{"value.strength_requirement": bson.M{"$gte": minStr}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find armature by strength requirement %d: %w", minStr, err)
	}
	defer cursor.Close(ctx)

	var armature []*domain.Armatura
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		armatura, err := extractArmaturaFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract armatura: %w", err)
		}
		armature = append(armature, armatura)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return armature, nil
}

// FindByCostRange retrieves armor within a cost range
func (r *ArmaturaMongoRepository) FindByCostRange(ctx context.Context, minCost, maxCost int, limit int) ([]*domain.Armatura, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"value.costo.valore": bson.M{"$gte": minCost, "$lte": maxCost}},
			{"value.cost.value": bson.M{"$gte": minCost, "$lte": maxCost}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find armature by cost range %d-%d: %w", minCost, maxCost, err)
	}
	defer cursor.Close(ctx)

	var armature []*domain.Armatura
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		armatura, err := extractArmaturaFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract armatura: %w", err)
		}
		armature = append(armature, armatura)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return armature, nil
}
