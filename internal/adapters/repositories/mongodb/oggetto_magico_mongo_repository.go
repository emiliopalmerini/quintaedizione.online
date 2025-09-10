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

// OggettoMagicoMongoRepository implements repositories.OggettoMagicoRepository using MongoDB
type OggettoMagicoMongoRepository struct {
	*BaseMongoRepository[*domain.OggettoMagico]
}

// NewOggettoMagicoMongoRepository creates a new MongoDB OggettoMagico repository
func NewOggettoMagicoMongoRepository(client *mongodb.Client) repositories.OggettoMagicoRepository {
	base := NewBaseMongoRepository[*domain.OggettoMagico](
		client,
		"oggetti_magici",
		[]string{"value.nome", "value.slug"},
	)
	
	return &OggettoMagicoMongoRepository{
		BaseMongoRepository: base,
	}
}

// extractOggettoMagicoFromDocument extracts OggettoMagico from the nested value field
func extractOggettoMagicoFromDocument(doc bson.M) (*domain.OggettoMagico, error) {
	valueData, exists := doc["value"]
	if !exists {
		return nil, fmt.Errorf("oggetto_magico document missing value field")
	}

	valueBytes, err := bson.Marshal(valueData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value data: %w", err)
	}

	var oggettoMagico domain.OggettoMagico
	err = bson.Unmarshal(valueBytes, &oggettoMagico)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal oggetto_magico: %w", err)
	}

	return &oggettoMagico, nil
}

// FindByNome retrieves a magic item by its name
func (r *OggettoMagicoMongoRepository) FindByNome(ctx context.Context, nome string) (*domain.OggettoMagico, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{"value.nome": nome}

	var doc bson.M
	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("failed to find oggetto_magico by nome %s: %w", nome, err)
	}

	return extractOggettoMagicoFromDocument(doc)
}

// FindByRarity retrieves magic items by rarity
func (r *OggettoMagicoMongoRepository) FindByRarity(ctx context.Context, rarity string, limit int) ([]*domain.OggettoMagico, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"value.rarita": primitive.Regex{Pattern: rarity, Options: "i"}},
			{"value.rarity": primitive.Regex{Pattern: rarity, Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find oggetti_magici by rarity %s: %w", rarity, err)
	}
	defer cursor.Close(ctx)

	var oggettiMagici []*domain.OggettoMagico
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		oggettoMagico, err := extractOggettoMagicoFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract oggetto_magico: %w", err)
		}
		oggettiMagici = append(oggettiMagici, oggettoMagico)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return oggettiMagici, nil
}

// FindByType retrieves magic items by type
func (r *OggettoMagicoMongoRepository) FindByType(ctx context.Context, itemType string, limit int) ([]*domain.OggettoMagico, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"value.tipo": primitive.Regex{Pattern: itemType, Options: "i"}},
			{"value.type": primitive.Regex{Pattern: itemType, Options: "i"}},
			{"value.categoria": primitive.Regex{Pattern: itemType, Options: "i"}},
			{"value.category": primitive.Regex{Pattern: itemType, Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find oggetti_magici by type %s: %w", itemType, err)
	}
	defer cursor.Close(ctx)

	var oggettiMagici []*domain.OggettoMagico
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		oggettoMagico, err := extractOggettoMagicoFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract oggetto_magico: %w", err)
		}
		oggettiMagici = append(oggettiMagici, oggettoMagico)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return oggettiMagici, nil
}

// FindByAttunement retrieves items that require attunement
func (r *OggettoMagicoMongoRepository) FindByAttunement(ctx context.Context, requiresAttunement bool, limit int) ([]*domain.OggettoMagico, error) {
	collection := r.client.GetCollection(r.collectionName)

	var filter bson.M
	if requiresAttunement {
		filter = bson.M{
			"$or": []bson.M{
				{"value.attunement": true},
				{"value.sintonia": true},
				{"contenuto": primitive.Regex{Pattern: "attunement", Options: "i"}},
				{"contenuto": primitive.Regex{Pattern: "sintonia", Options: "i"}},
				{"contenuto": primitive.Regex{Pattern: "requires attunement", Options: "i"}},
				{"contenuto": primitive.Regex{Pattern: "richiede sintonia", Options: "i"}},
			},
		}
	} else {
		filter = bson.M{
			"$and": []bson.M{
				{"value.attunement": bson.M{"$ne": true}},
				{"value.sintonia": bson.M{"$ne": true}},
				{"contenuto": bson.M{"$not": primitive.Regex{Pattern: "attunement|sintonia", Options: "i"}}},
			},
		}
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find oggetti_magici by attunement %t: %w", requiresAttunement, err)
	}
	defer cursor.Close(ctx)

	var oggettiMagici []*domain.OggettoMagico
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		oggettoMagico, err := extractOggettoMagicoFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract oggetto_magico: %w", err)
		}
		oggettiMagici = append(oggettiMagici, oggettoMagico)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return oggettiMagici, nil
}

// FindConsumableItems retrieves consumable magic items
func (r *OggettoMagicoMongoRepository) FindConsumableItems(ctx context.Context, limit int) ([]*domain.OggettoMagico, error) {
	collection := r.client.GetCollection(r.collectionName)

	// Search for consumable item types
	consumableTerms := []string{"pozione", "potion", "scrollo", "scroll", "pergamena", "parchment", 
		"consumabile", "consumable", "usa e getta", "single use"}

	orFilters := make([]bson.M, 0, len(consumableTerms)*3)
	for _, term := range consumableTerms {
		orFilters = append(orFilters,
			bson.M{"value.tipo": primitive.Regex{Pattern: term, Options: "i"}},
			bson.M{"value.categoria": primitive.Regex{Pattern: term, Options: "i"}},
			bson.M{"contenuto": primitive.Regex{Pattern: term, Options: "i"}},
		)
	}

	filter := bson.M{"$or": orFilters}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find consumable oggetti_magici: %w", err)
	}
	defer cursor.Close(ctx)

	var oggettiMagici []*domain.OggettoMagico
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		oggettoMagico, err := extractOggettoMagicoFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract oggetto_magico: %w", err)
		}
		oggettiMagici = append(oggettiMagici, oggettoMagico)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return oggettiMagici, nil
}