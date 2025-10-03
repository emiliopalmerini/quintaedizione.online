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

// AnimaleMongoRepository implements AnimaleRepository for MongoDB
type AnimaleMongoRepository struct {
	*BaseMongoRepository[*domain.Animale]
}

// extractAnimaleFromDocument extracts Animale from the flattened document
func extractAnimaleFromDocument(doc bson.M) (*domain.Animale, error) {
	return ExtractEntityFromDocument[domain.Animale](doc, false)
}

// NewAnimaleMongoRepository creates a new AnimaleMongoRepository
func NewAnimaleMongoRepository(client *mongodb.Client) repositories.AnimaleRepository {
	base := NewBaseMongoRepository[*domain.Animale](
		client,
		"animali",
		[]string{"nome", "slug"},
	)

	return &AnimaleMongoRepository{
		BaseMongoRepository: base,
	}
}

// FindByNome retrieves an animal by its name
func (r *AnimaleMongoRepository) FindByNome(ctx context.Context, nome string) (*domain.Animale, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{"nome": nome}

	var doc bson.M
	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("failed to find animale by nome %s: %w", nome, err)
	}

	return extractAnimaleFromDocument(doc)
}

// FindBySize retrieves animals by size
func (r *AnimaleMongoRepository) FindBySize(ctx context.Context, size string, limit int) ([]*domain.Animale, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"taglia": primitive.Regex{Pattern: size, Options: "i"}},
			{"size": primitive.Regex{Pattern: size, Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find animali by size %s: %w", size, err)
	}
	defer cursor.Close(ctx)

	var animali []*domain.Animale
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		animale, err := extractAnimaleFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract animale: %w", err)
		}
		animali = append(animali, animale)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return animali, nil
}

// FindByChallengeRating retrieves animals by challenge rating
func (r *AnimaleMongoRepository) FindByChallengeRating(ctx context.Context, cr float64, limit int) ([]*domain.Animale, error) {
	collection := r.client.GetCollection(r.collectionName)

	// Convert float to string for comparison (GS is stored as string like "1/4", "1/2", "1", "2", etc.)
	crStr := strconv.FormatFloat(cr, 'f', -1, 64)
	
	filter := bson.M{
		"$or": []bson.M{
			{"grado_sfida.valore": crStr},
			{"challenge_rating.value": crStr},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find animali by CR %f: %w", cr, err)
	}
	defer cursor.Close(ctx)

	var animali []*domain.Animale
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		animale, err := extractAnimaleFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract animale: %w", err)
		}
		animali = append(animali, animale)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return animali, nil
}

// FindByEnvironment retrieves animals by habitat
func (r *AnimaleMongoRepository) FindByEnvironment(ctx context.Context, environment string, limit int) ([]*domain.Animale, error) {
	collection := r.client.GetCollection(r.collectionName)

	// Since environment isn't a direct field in our domain model, we'll search in content and traits
	filter := bson.M{
		"$or": []bson.M{
			{"tratti.descrizione": primitive.Regex{Pattern: environment, Options: "i"}},
			{"contenuto": primitive.Regex{Pattern: environment, Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find animali by environment %s: %w", environment, err)
	}
	defer cursor.Close(ctx)

	var animali []*domain.Animale
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		animale, err := extractAnimaleFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract animale: %w", err)
		}
		animali = append(animali, animale)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return animali, nil
}