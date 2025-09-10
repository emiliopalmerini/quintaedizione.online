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

// EquipaggiamentoMongoRepository implements repositories.EquipaggiamentoRepository using MongoDB
type EquipaggiamentoMongoRepository struct {
	*BaseMongoRepository[*domain.Equipaggiamento]
}

// NewEquipaggiamentoMongoRepository creates a new MongoDB Equipaggiamento repository
func NewEquipaggiamentoMongoRepository(client *mongodb.Client) repositories.EquipaggiamentoRepository {
	base := NewBaseMongoRepository[*domain.Equipaggiamento](
		client,
		"equipaggiamenti",
		[]string{"value.nome", "value.slug"},
	)
	
	return &EquipaggiamentoMongoRepository{
		BaseMongoRepository: base,
	}
}

// extractEquipaggiamentoFromDocument extracts Equipaggiamento from the nested value field
func extractEquipaggiamentoFromDocument(doc bson.M) (*domain.Equipaggiamento, error) {
	valueData, exists := doc["value"]
	if !exists {
		return nil, fmt.Errorf("equipaggiamento document missing value field")
	}

	valueBytes, err := bson.Marshal(valueData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value data: %w", err)
	}

	var equipaggiamento domain.Equipaggiamento
	err = bson.Unmarshal(valueBytes, &equipaggiamento)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal equipaggiamento: %w", err)
	}

	return &equipaggiamento, nil
}

// FindByNome retrieves equipment by its name
func (r *EquipaggiamentoMongoRepository) FindByNome(ctx context.Context, nome string) (*domain.Equipaggiamento, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{"value.nome": nome}

	var doc bson.M
	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("failed to find equipaggiamento by nome %s: %w", nome, err)
	}

	return extractEquipaggiamentoFromDocument(doc)
}

// FindByCategory retrieves equipment by category
func (r *EquipaggiamentoMongoRepository) FindByCategory(ctx context.Context, category string, limit int) ([]*domain.Equipaggiamento, error) {
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
		return nil, fmt.Errorf("failed to find equipaggiamento by category %s: %w", category, err)
	}
	defer cursor.Close(ctx)

	var equipaggiamenti []*domain.Equipaggiamento
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		equipaggiamento, err := extractEquipaggiamentoFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract equipaggiamento: %w", err)
		}
		equipaggiamenti = append(equipaggiamenti, equipaggiamento)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return equipaggiamenti, nil
}

// FindByPriceRange retrieves equipment within price range
func (r *EquipaggiamentoMongoRepository) FindByPriceRange(ctx context.Context, minPrice, maxPrice float64, limit int) ([]*domain.Equipaggiamento, error) {
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
		return nil, fmt.Errorf("failed to find equipaggiamento by price range %.2f-%.2f: %w", minPrice, maxPrice, err)
	}
	defer cursor.Close(ctx)

	var equipaggiamenti []*domain.Equipaggiamento
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		equipaggiamento, err := extractEquipaggiamentoFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract equipaggiamento: %w", err)
		}
		equipaggiamenti = append(equipaggiamenti, equipaggiamento)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return equipaggiamenti, nil
}

// FindByWeight retrieves equipment by weight
func (r *EquipaggiamentoMongoRepository) FindByWeight(ctx context.Context, maxWeight float64, limit int) ([]*domain.Equipaggiamento, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{
				"value.peso.valore": bson.M{
					"$lte": maxWeight,
				},
			},
			{
				"value.weight.value": bson.M{
					"$lte": maxWeight,
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
		return nil, fmt.Errorf("failed to find equipaggiamento by weight %.2f: %w", maxWeight, err)
	}
	defer cursor.Close(ctx)

	var equipaggiamenti []*domain.Equipaggiamento
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		equipaggiamento, err := extractEquipaggiamentoFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract equipaggiamento: %w", err)
		}
		equipaggiamenti = append(equipaggiamenti, equipaggiamento)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return equipaggiamenti, nil
}