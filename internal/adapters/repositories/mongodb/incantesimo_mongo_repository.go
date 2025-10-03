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

// IncantesimoMongoRepository implements repositories.IncantesimoRepository using MongoDB
type IncantesimoMongoRepository struct {
	*BaseMongoRepository[*domain.Incantesimo]
}

// NewIncantesimoMongoRepository creates a new MongoDB Incantesimo repository
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

// extractIncantesimoFromDocument extracts Incantesimo from the flattened document
func extractIncantesimoFromDocument(doc bson.M) (*domain.Incantesimo, error) {
	return ExtractEntityFromDocument[domain.Incantesimo](doc, false)
}

// FindByNome retrieves a spell by its name
func (r *IncantesimoMongoRepository) FindByNome(ctx context.Context, nome string) (*domain.Incantesimo, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{"nome": nome}

	var doc bson.M
	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("failed to find incantesimo by nome %s: %w", nome, err)
	}

	return extractIncantesimoFromDocument(doc)
}

// FindByLevel retrieves spells by level
func (r *IncantesimoMongoRepository) FindByLevel(ctx context.Context, level int, limit int) ([]*domain.Incantesimo, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"livello": level},
			{"level": level},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find incantesimi by level %d: %w", level, err)
	}
	defer cursor.Close(ctx)

	var incantesimi []*domain.Incantesimo
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		incantesimo, err := extractIncantesimoFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract incantesimo: %w", err)
		}
		incantesimi = append(incantesimi, incantesimo)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return incantesimi, nil
}

// FindBySchool retrieves spells by school of magic
func (r *IncantesimoMongoRepository) FindBySchool(ctx context.Context, school string, limit int) ([]*domain.Incantesimo, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"scuola": primitive.Regex{Pattern: school, Options: "i"}},
			{"school": primitive.Regex{Pattern: school, Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find incantesimi by school %s: %w", school, err)
	}
	defer cursor.Close(ctx)

	var incantesimi []*domain.Incantesimo
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		incantesimo, err := extractIncantesimoFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract incantesimo: %w", err)
		}
		incantesimi = append(incantesimi, incantesimo)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return incantesimi, nil
}

// FindByClass retrieves spells available to a specific class
func (r *IncantesimoMongoRepository) FindByClass(ctx context.Context, className string, limit int) ([]*domain.Incantesimo, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"classi": primitive.Regex{Pattern: className, Options: "i"}},
			{"classes": primitive.Regex{Pattern: className, Options: "i"}},
			{"contenuto": primitive.Regex{Pattern: className, Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find incantesimi by class %s: %w", className, err)
	}
	defer cursor.Close(ctx)

	var incantesimi []*domain.Incantesimo
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		incantesimo, err := extractIncantesimoFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract incantesimo: %w", err)
		}
		incantesimi = append(incantesimi, incantesimo)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return incantesimi, nil
}

// FindByLevelAndClass retrieves spells by level and class
func (r *IncantesimoMongoRepository) FindByLevelAndClass(ctx context.Context, level int, className string, limit int) ([]*domain.Incantesimo, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$and": []bson.M{
			{
				"$or": []bson.M{
					{"livello": level},
					{"level": level},
				},
			},
			{
				"$or": []bson.M{
					{"classi": primitive.Regex{Pattern: className, Options: "i"}},
					{"classes": primitive.Regex{Pattern: className, Options: "i"}},
					{"contenuto": primitive.Regex{Pattern: className, Options: "i"}},
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
		return nil, fmt.Errorf("failed to find incantesimi by level %d and class %s: %w", level, className, err)
	}
	defer cursor.Close(ctx)

	var incantesimi []*domain.Incantesimo
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		incantesimo, err := extractIncantesimoFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract incantesimo: %w", err)
		}
		incantesimi = append(incantesimi, incantesimo)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return incantesimi, nil
}

// FindByComponents retrieves spells by required components
func (r *IncantesimoMongoRepository) FindByComponents(ctx context.Context, components []string, limit int) ([]*domain.Incantesimo, error) {
	collection := r.client.GetCollection(r.collectionName)

	// Build component filters
	componentFilters := make([]bson.M, 0, len(components)*2)
	for _, component := range components {
		componentFilters = append(componentFilters,
			bson.M{"componenti": primitive.Regex{Pattern: component, Options: "i"}},
			bson.M{"components": primitive.Regex{Pattern: component, Options: "i"}},
		)
	}

	filter := bson.M{
		"$or": componentFilters,
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find incantesimi by components %v: %w", components, err)
	}
	defer cursor.Close(ctx)

	var incantesimi []*domain.Incantesimo
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		incantesimo, err := extractIncantesimoFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract incantesimo: %w", err)
		}
		incantesimi = append(incantesimi, incantesimo)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return incantesimi, nil
}

// FindRitualSpells retrieves ritual spells
func (r *IncantesimoMongoRepository) FindRitualSpells(ctx context.Context, limit int) ([]*domain.Incantesimo, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"rituale": true},
			{"ritual": true},
			{"contenuto": primitive.Regex{Pattern: "rituale", Options: "i"}},
			{"contenuto": primitive.Regex{Pattern: "ritual", Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find ritual spells: %w", err)
	}
	defer cursor.Close(ctx)

	var incantesimi []*domain.Incantesimo
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		incantesimo, err := extractIncantesimoFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract incantesimo: %w", err)
		}
		incantesimi = append(incantesimi, incantesimo)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return incantesimi, nil
}