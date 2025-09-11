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

// TalentoMongoRepository implements repositories.TalentoRepository using MongoDB
type TalentoMongoRepository struct {
	*BaseMongoRepository[*domain.Talento]
}

// NewTalentoMongoRepository creates a new MongoDB Talento repository
func NewTalentoMongoRepository(client *mongodb.Client) repositories.TalentoRepository {
	base := NewBaseMongoRepository[*domain.Talento](
		client,
		"talenti",
		[]string{"value.nome", "value.slug"},
	)
	
	return &TalentoMongoRepository{
		BaseMongoRepository: base,
	}
}

// extractTalentoFromDocument extracts Talento from the nested value field
func extractTalentoFromDocument(doc bson.M) (*domain.Talento, error) {
	valueData, exists := doc["value"]
	if !exists {
		return nil, fmt.Errorf("talento document missing value field")
	}

	valueBytes, err := bson.Marshal(valueData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value data: %w", err)
	}

	var talento domain.Talento
	err = bson.Unmarshal(valueBytes, &talento)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal talento: %w", err)
	}

	return &talento, nil
}

// FindByNome retrieves a feat by its name
func (r *TalentoMongoRepository) FindByNome(ctx context.Context, nome string) (*domain.Talento, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{"value.nome": nome}

	var doc bson.M
	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("failed to find talento by nome %s: %w", nome, err)
	}

	return extractTalentoFromDocument(doc)
}

// FindByPrerequisite retrieves feats with specific prerequisites
func (r *TalentoMongoRepository) FindByPrerequisite(ctx context.Context, prerequisite string, limit int) ([]*domain.Talento, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"value.prerequisiti": primitive.Regex{Pattern: prerequisite, Options: "i"}},
			{"value.prerequisites": primitive.Regex{Pattern: prerequisite, Options: "i"}},
			{"contenuto": primitive.Regex{Pattern: prerequisite, Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find talenti by prerequisite %s: %w", prerequisite, err)
	}
	defer cursor.Close(ctx)

	var talenti []*domain.Talento
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		talento, err := extractTalentoFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract talento: %w", err)
		}
		talenti = append(talenti, talento)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return talenti, nil
}

// FindByAbilityScoreIncrease retrieves feats that increase ability scores
func (r *TalentoMongoRepository) FindByAbilityScoreIncrease(ctx context.Context, limit int) ([]*domain.Talento, error) {
	collection := r.client.GetCollection(r.collectionName)

	// Search for common ability score terms in Italian and English
	abilityTerms := []string{
		"Forza", "Destrezza", "Costituzione", "Intelligenza", "Saggezza", "Carisma",
		"FOR", "DES", "COS", "INT", "SAG", "CAR",
		"Strength", "Dexterity", "Constitution", "Intelligence", "Wisdom", "Charisma",
		"STR", "DEX", "CON", "WIS", "CHA",
		"punteggio", "score", "aumenta", "increase",
	}

	// Build regex filters for ability score increases
	orFilters := make([]bson.M, 0, len(abilityTerms)*2)
	for _, term := range abilityTerms {
		orFilters = append(orFilters,
			bson.M{"value.descrizione": primitive.Regex{Pattern: term, Options: "i"}},
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
		return nil, fmt.Errorf("failed to find talenti with ability score increases: %w", err)
	}
	defer cursor.Close(ctx)

	var talenti []*domain.Talento
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		talento, err := extractTalentoFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract talento: %w", err)
		}
		talenti = append(talenti, talento)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return talenti, nil
}