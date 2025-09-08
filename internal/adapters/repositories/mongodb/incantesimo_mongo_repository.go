package mongodb

import (
	"context"
	"fmt"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain/repositories"
	"github.com/emiliopalmerini/due-draghi-5e-srd/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// IncantesimoMongoRepository implements IncantesimoRepository for MongoDB
type IncantesimoMongoRepository struct {
	*BaseMongoRepository[*domain.Incantesimo]
}

// NewIncantesimoMongoRepository creates a new IncantesimoMongoRepository
func NewIncantesimoMongoRepository(client *mongodb.Client) repositories.IncantesimoRepository {
	base := NewBaseMongoRepository[*domain.Incantesimo](
		client,
		"incantesimi",
		[]string{"nome", "slug"},
	)

	return &IncantesimoMongoRepository{
		BaseMongoRepository: base,
	}
}

// FindByNome retrieves a spell by its name
func (r *IncantesimoMongoRepository) FindByNome(ctx context.Context, nome string) (*domain.Incantesimo, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{"nome": nome}

	var incantesimo domain.Incantesimo
	err := collection.FindOne(ctx, filter).Decode(&incantesimo)
	if err != nil {
		return nil, fmt.Errorf("failed to find incantesimo by nome %s: %w", nome, err)
	}

	return &incantesimo, nil
}

// FindByLevel retrieves spells by level
func (r *IncantesimoMongoRepository) FindByLevel(ctx context.Context, level int, limit int) ([]*domain.Incantesimo, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{"livello": level}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	return r.findWithFilter(ctx, collection, filter, opts)
}

// FindBySchool retrieves spells by school of magic
func (r *IncantesimoMongoRepository) FindBySchool(ctx context.Context, school string, limit int) ([]*domain.Incantesimo, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{"scuola": primitive.Regex{Pattern: school, Options: "i"}}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	return r.findWithFilter(ctx, collection, filter, opts)
}

// FindByClass retrieves spells available to a specific class
func (r *IncantesimoMongoRepository) FindByClass(ctx context.Context, className string, limit int) ([]*domain.Incantesimo, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"classi": bson.M{"$in": []string{className}}},
			{"classes": bson.M{"$in": []string{className}}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	return r.findWithFilter(ctx, collection, filter, opts)
}

// FindByLevelAndClass retrieves spells by level and class
func (r *IncantesimoMongoRepository) FindByLevelAndClass(ctx context.Context, level int, className string, limit int) ([]*domain.Incantesimo, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"livello": level,
		"$or": []bson.M{
			{"classi": bson.M{"$in": []string{className}}},
			{"classes": bson.M{"$in": []string{className}}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	return r.findWithFilter(ctx, collection, filter, opts)
}

// FindByComponents retrieves spells by required components
func (r *IncantesimoMongoRepository) FindByComponents(ctx context.Context, components []string, limit int) ([]*domain.Incantesimo, error) {
	collection := r.client.GetCollection(r.collectionName)

	// Build filter for spells that have ALL specified components
	componentFilters := []bson.M{}
	for _, component := range components {
		componentFilters = append(componentFilters, bson.M{
			"$or": []bson.M{
				{"componenti.componenti": bson.M{"$in": []string{component}}},
				{"components.components": bson.M{"$in": []string{component}}},
			},
		})
	}

	filter := bson.M{"$and": componentFilters}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	return r.findWithFilter(ctx, collection, filter, opts)
}

// FindRitualSpells retrieves ritual spells
func (r *IncantesimoMongoRepository) FindRitualSpells(ctx context.Context, limit int) ([]*domain.Incantesimo, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"rituale": true},
			{"ritual": true},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	return r.findWithFilter(ctx, collection, filter, opts)
}

// findWithFilter is a helper method to execute queries and return results
func (r *IncantesimoMongoRepository) findWithFilter(ctx context.Context, collection *mongo.Collection, filter bson.M, opts *options.FindOptions) ([]*domain.Incantesimo, error) {
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find incantesimi: %w", err)
	}
	defer cursor.Close(ctx)

	var incantesimi []*domain.Incantesimo
	for cursor.Next(ctx) {
		var incantesimo domain.Incantesimo
		if err := cursor.Decode(&incantesimo); err != nil {
			return nil, fmt.Errorf("failed to decode incantesimo: %w", err)
		}
		incantesimi = append(incantesimi, &incantesimo)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return incantesimi, nil
}
