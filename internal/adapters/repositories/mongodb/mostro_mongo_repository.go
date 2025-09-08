package mongodb

import (
	"context"
	"fmt"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain/repositories"
	"github.com/emiliopalmerini/due-draghi-5e-srd/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MostroMongoRepository implements MostroRepository for MongoDB
type MostroMongoRepository struct {
	*BaseMongoRepository[*domain.Mostro]
}

// NewMostroMongoRepository creates a new MostroMongoRepository
func NewMostroMongoRepository(client *mongodb.Client) repositories.MostroRepository {
	base := NewBaseMongoRepository[*domain.Mostro](
		client,
		"mostri",
		[]string{"nome", "slug"},
	)

	return &MostroMongoRepository{
		BaseMongoRepository: base,
	}
}

// FindByNome retrieves a monster by its name
func (r *MostroMongoRepository) FindByNome(ctx context.Context, nome string) (*domain.Mostro, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{"nome": nome}

	var mostro domain.Mostro
	err := collection.FindOne(ctx, filter).Decode(&mostro)
	if err != nil {
		return nil, fmt.Errorf("failed to find mostro by nome %s: %w", nome, err)
	}

	return &mostro, nil
}

// FindByChallengeRating retrieves monsters by challenge rating
func (r *MostroMongoRepository) FindByChallengeRating(ctx context.Context, cr float64, limit int) ([]*domain.Mostro, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{"grado_sfida": cr}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	return r.findWithFilter(ctx, collection, filter, opts)
}

// FindByChallengeRatingRange retrieves monsters within CR range
func (r *MostroMongoRepository) FindByChallengeRatingRange(ctx context.Context, minCR, maxCR float64, limit int) ([]*domain.Mostro, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"grado_sfida": bson.M{
			"$gte": minCR,
			"$lte": maxCR,
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	return r.findWithFilter(ctx, collection, filter, opts)
}

// FindByType retrieves monsters by type (Aberrazione, Bestia, etc.)
func (r *MostroMongoRepository) FindByType(ctx context.Context, tipoMostro domain.TipoMostro, limit int) ([]*domain.Mostro, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{"tipo": string(tipoMostro)}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	return r.findWithFilter(ctx, collection, filter, opts)
}

// FindBySize retrieves monsters by size
func (r *MostroMongoRepository) FindBySize(ctx context.Context, size string, limit int) ([]*domain.Mostro, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{"taglia": size}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	return r.findWithFilter(ctx, collection, filter, opts)
}

// FindByEnvironment retrieves monsters by habitat/environment
func (r *MostroMongoRepository) FindByEnvironment(ctx context.Context, environment string, limit int) ([]*domain.Mostro, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"ambiente": environment},
			{"habitat": environment},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	return r.findWithFilter(ctx, collection, filter, opts)
}

// FindByAlignment retrieves monsters by alignment
func (r *MostroMongoRepository) FindByAlignment(ctx context.Context, alignment string, limit int) ([]*domain.Mostro, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{"allineamento": alignment}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	return r.findWithFilter(ctx, collection, filter, opts)
}

// FindSpellcasters retrieves monsters that can cast spells
func (r *MostroMongoRepository) FindSpellcasters(ctx context.Context, limit int) ([]*domain.Mostro, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"incantesimi_innati": bson.M{"$exists": true, "$ne": nil}},
			{"incantesimi": bson.M{"$exists": true, "$ne": nil}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	return r.findWithFilter(ctx, collection, filter, opts)
}

// FindLegendaryMonsters retrieves monsters with legendary actions
func (r *MostroMongoRepository) FindLegendaryMonsters(ctx context.Context, limit int) ([]*domain.Mostro, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"azioni_leggendarie": bson.M{
			"$exists": true,
			"$ne":     nil,
			"$not":    bson.M{"$size": 0},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	return r.findWithFilter(ctx, collection, filter, opts)
}

// findWithFilter is a helper method to execute queries and return results
func (r *MostroMongoRepository) findWithFilter(ctx context.Context, collection *mongo.Collection, filter bson.M, opts *options.FindOptions) ([]*domain.Mostro, error) {
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find mostri: %w", err)
	}
	defer cursor.Close(ctx)

	var mostri []*domain.Mostro
	for cursor.Next(ctx) {
		var mostro domain.Mostro
		if err := cursor.Decode(&mostro); err != nil {
			return nil, fmt.Errorf("failed to decode mostro: %w", err)
		}
		mostri = append(mostri, &mostro)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return mostri, nil
}
