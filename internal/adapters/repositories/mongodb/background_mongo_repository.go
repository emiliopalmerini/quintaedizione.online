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

// BackgroundMongoRepository implements repositories.BackgroundRepository using MongoDB
type BackgroundMongoRepository struct {
	*BaseMongoRepository[*domain.Background]
}

// NewBackgroundMongoRepository creates a new MongoDB Background repository
func NewBackgroundMongoRepository(client *mongodb.Client) repositories.BackgroundRepository {
	base := NewBaseMongoRepository[*domain.Background](
		client,
		"backgrounds",
		[]string{"value.nome", "value.slug"},
	)
	
	return &BackgroundMongoRepository{
		BaseMongoRepository: base,
	}
}

// FindByNome retrieves a background by its name
func (r *BackgroundMongoRepository) FindByNome(ctx context.Context, nome string) (*domain.Background, error) {
	collection := r.client.GetCollection(r.collectionName)
	
	filter := bson.M{"value.nome": nome}
	
	var doc bson.M
	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("failed to find background by nome %s: %w", nome, err)
	}
	
	return extractBackgroundFromDocument(doc)
}

// FindBySkillProficiency retrieves backgrounds that grant specific skill proficiencies
func (r *BackgroundMongoRepository) FindBySkillProficiency(ctx context.Context, skill string, limit int) ([]*domain.Background, error) {
	collection := r.client.GetCollection(r.collectionName)
	
	filter := bson.M{
		"value.competenze_abilita": bson.M{
			"$elemMatch": bson.M{
				"$regex": primitive.Regex{Pattern: skill, Options: "i"},
			},
		},
	}
	
	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find backgrounds by skill %s: %w", skill, err)
	}
	defer cursor.Close(ctx)
	
	var backgrounds []*domain.Background
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}
		
		background, err := extractBackgroundFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract background: %w", err)
		}
		backgrounds = append(backgrounds, background)
	}
	
	return backgrounds, nil
}

// FindByLanguage retrieves backgrounds that provide specific languages
func (r *BackgroundMongoRepository) FindByLanguage(ctx context.Context, language string, limit int) ([]*domain.Background, error) {
	collection := r.client.GetCollection(r.collectionName)
	
	filter := bson.M{
		"value.linguaggi": bson.M{
			"$elemMatch": bson.M{
				"$regex": primitive.Regex{Pattern: language, Options: "i"},
			},
		},
	}
	
	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find backgrounds by language %s: %w", language, err)
	}
	defer cursor.Close(ctx)
	
	var backgrounds []*domain.Background
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}
		
		background, err := extractBackgroundFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract background: %w", err)
		}
		backgrounds = append(backgrounds, background)
	}
	
	return backgrounds, nil
}

// FindByToolProficiency retrieves backgrounds that grant tool proficiencies
func (r *BackgroundMongoRepository) FindByToolProficiency(ctx context.Context, tool string, limit int) ([]*domain.Background, error) {
	collection := r.client.GetCollection(r.collectionName)
	
	filter := bson.M{
		"value.competenze_strumenti": bson.M{
			"$elemMatch": bson.M{
				"$regex": primitive.Regex{Pattern: tool, Options: "i"},
			},
		},
	}
	
	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find backgrounds by tool %s: %w", tool, err)
	}
	defer cursor.Close(ctx)
	
	var backgrounds []*domain.Background
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}
		
		background, err := extractBackgroundFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract background: %w", err)
		}
		backgrounds = append(backgrounds, background)
	}
	
	return backgrounds, nil
}

// extractBackgroundFromDocument extracts Background from the nested value field
func extractBackgroundFromDocument(doc bson.M) (*domain.Background, error) {
	return ExtractEntityFromDocument[domain.Background](doc, true)
}