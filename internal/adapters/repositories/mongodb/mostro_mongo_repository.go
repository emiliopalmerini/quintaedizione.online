package mongodb

import (
	"context"
	"fmt"
	"strconv"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain/repositories"
	"github.com/emiliopalmerini/due-draghi-5e-srd/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MostroMongoRepository implements repositories.MostroRepository using MongoDB
type MostroMongoRepository struct {
	*BaseMongoRepository[*domain.Mostro]
}

// NewMostroMongoRepository creates a new MongoDB Mostro repository
func NewMostroMongoRepository(client *mongodb.Client) repositories.MostroRepository {
	base := NewBaseMongoRepository[*domain.Mostro](
		client,
		"mostri",
		[]string{"value.nome", "value.slug"},
	)
	
	return &MostroMongoRepository{
		BaseMongoRepository: base,
	}
}

// extractMostroFromDocument extracts Mostro from the nested value field
func extractMostroFromDocument(doc bson.M) (*domain.Mostro, error) {
	valueData, exists := doc["value"]
	if !exists {
		return nil, fmt.Errorf("mostro document missing value field")
	}

	valueBytes, err := bson.Marshal(valueData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value data: %w", err)
	}

	var mostro domain.Mostro
	err = bson.Unmarshal(valueBytes, &mostro)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal mostro: %w", err)
	}

	return &mostro, nil
}

// FindByNome retrieves a monster by its name
func (r *MostroMongoRepository) FindByNome(ctx context.Context, nome string) (*domain.Mostro, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{"value.nome": nome}

	var doc bson.M
	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("failed to find mostro by nome %s: %w", nome, err)
	}

	return extractMostroFromDocument(doc)
}

// FindByChallengeRating retrieves monsters by challenge rating
func (r *MostroMongoRepository) FindByChallengeRating(ctx context.Context, cr float64, limit int) ([]*domain.Mostro, error) {
	collection := r.client.GetCollection(r.collectionName)

	// Convert float to string for comparison (GS is stored as string like "1/4", "1/2", "1", "2", etc.)
	crStr := strconv.FormatFloat(cr, 'f', -1, 64)
	
	filter := bson.M{
		"$or": []bson.M{
			{"value.grado_sfida.valore": crStr},
			{"value.challenge_rating.value": crStr},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find mostri by CR %f: %w", cr, err)
	}
	defer cursor.Close(ctx)

	var mostri []*domain.Mostro
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		mostro, err := extractMostroFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract mostro: %w", err)
		}
		mostri = append(mostri, mostro)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return mostri, nil
}

// FindByChallengeRatingRange retrieves monsters within CR range
func (r *MostroMongoRepository) FindByChallengeRatingRange(ctx context.Context, minCR, maxCR float64, limit int) ([]*domain.Mostro, error) {
	collection := r.client.GetCollection(r.collectionName)

	// For simplicity, use content search for CR ranges
	filter := bson.M{
		"$or": []bson.M{
			{"value.grado_sfida.valore": bson.M{"$gte": minCR, "$lte": maxCR}},
			{"value.challenge_rating.value": bson.M{"$gte": minCR, "$lte": maxCR}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find mostri by CR range %.2f-%.2f: %w", minCR, maxCR, err)
	}
	defer cursor.Close(ctx)

	var mostri []*domain.Mostro
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		mostro, err := extractMostroFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract mostro: %w", err)
		}
		mostri = append(mostri, mostro)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return mostri, nil
}

// FindByType retrieves monsters by type
func (r *MostroMongoRepository) FindByType(ctx context.Context, tipoMostro domain.TipoMostro, limit int) ([]*domain.Mostro, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"value.tipo": string(tipoMostro)},
			{"value.type": string(tipoMostro)},
			{"value.tipo": primitive.Regex{Pattern: string(tipoMostro), Options: "i"}},
			{"value.type": primitive.Regex{Pattern: string(tipoMostro), Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find mostri by type %s: %w", tipoMostro, err)
	}
	defer cursor.Close(ctx)

	var mostri []*domain.Mostro
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		mostro, err := extractMostroFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract mostro: %w", err)
		}
		mostri = append(mostri, mostro)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return mostri, nil
}

// FindBySize retrieves monsters by size
func (r *MostroMongoRepository) FindBySize(ctx context.Context, size string, limit int) ([]*domain.Mostro, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"value.taglia": primitive.Regex{Pattern: size, Options: "i"}},
			{"value.size": primitive.Regex{Pattern: size, Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find mostri by size %s: %w", size, err)
	}
	defer cursor.Close(ctx)

	var mostri []*domain.Mostro
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		mostro, err := extractMostroFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract mostro: %w", err)
		}
		mostri = append(mostri, mostro)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return mostri, nil
}

// FindByEnvironment retrieves monsters by habitat/environment
func (r *MostroMongoRepository) FindByEnvironment(ctx context.Context, environment string, limit int) ([]*domain.Mostro, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"value.ambiente": primitive.Regex{Pattern: environment, Options: "i"}},
			{"value.environment": primitive.Regex{Pattern: environment, Options: "i"}},
			{"contenuto": primitive.Regex{Pattern: environment, Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find mostri by environment %s: %w", environment, err)
	}
	defer cursor.Close(ctx)

	var mostri []*domain.Mostro
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		mostro, err := extractMostroFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract mostro: %w", err)
		}
		mostri = append(mostri, mostro)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return mostri, nil
}

// FindByAlignment retrieves monsters by alignment
func (r *MostroMongoRepository) FindByAlignment(ctx context.Context, alignment string, limit int) ([]*domain.Mostro, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"value.allineamento": primitive.Regex{Pattern: alignment, Options: "i"}},
			{"value.alignment": primitive.Regex{Pattern: alignment, Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find mostri by alignment %s: %w", alignment, err)
	}
	defer cursor.Close(ctx)

	var mostri []*domain.Mostro
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		mostro, err := extractMostroFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract mostro: %w", err)
		}
		mostri = append(mostri, mostro)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return mostri, nil
}

// FindSpellcasters retrieves monsters that can cast spells
func (r *MostroMongoRepository) FindSpellcasters(ctx context.Context, limit int) ([]*domain.Mostro, error) {
	collection := r.client.GetCollection(r.collectionName)

	spellTerms := []string{"incantesimi", "spells", "spell", "incantesimo", "magia", "magic", "casting", "caster"}

	orFilters := make([]bson.M, 0, len(spellTerms)*2)
	for _, term := range spellTerms {
		orFilters = append(orFilters,
			bson.M{"value.tratti.nome": primitive.Regex{Pattern: term, Options: "i"}},
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
		return nil, fmt.Errorf("failed to find spellcaster mostri: %w", err)
	}
	defer cursor.Close(ctx)

	var mostri []*domain.Mostro
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		mostro, err := extractMostroFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract mostro: %w", err)
		}
		mostri = append(mostri, mostro)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return mostri, nil
}

// FindLegendaryMonsters retrieves monsters with legendary actions
func (r *MostroMongoRepository) FindLegendaryMonsters(ctx context.Context, limit int) ([]*domain.Mostro, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"value.azioni_leggendarie": bson.M{"$exists": true, "$ne": nil}},
			{"value.legendary_actions": bson.M{"$exists": true, "$ne": nil}},
			{"contenuto": primitive.Regex{Pattern: "legendary|leggendari", Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find legendary mostri: %w", err)
	}
	defer cursor.Close(ctx)

	var mostri []*domain.Mostro
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		mostro, err := extractMostroFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract mostro: %w", err)
		}
		mostri = append(mostri, mostro)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return mostri, nil
}