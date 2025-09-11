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

// ClasseMongoRepository implements ClasseRepository for MongoDB
type ClasseMongoRepository struct {
	*BaseMongoRepository[*domain.Classe]
}

// extractClasseFromDocument extracts Classe from the nested value field
func extractClasseFromDocument(doc bson.M) (*domain.Classe, error) {
	valueData, exists := doc["value"]
	if !exists {
		return nil, fmt.Errorf("classe document missing value field")
	}

	valueBytes, err := bson.Marshal(valueData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value data: %w", err)
	}

	var classe domain.Classe
	err = bson.Unmarshal(valueBytes, &classe)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal classe: %w", err)
	}

	return &classe, nil
}

// NewClasseMongoRepository creates a new ClasseMongoRepository
func NewClasseMongoRepository(client *mongodb.Client) repositories.ClasseRepository {
	base := NewBaseMongoRepository[*domain.Classe](
		client,
		"classi",
		[]string{"value.nome", "value.slug"},
	)

	return &ClasseMongoRepository{
		BaseMongoRepository: base,
	}
}

// FindByNome retrieves a class by its name
func (r *ClasseMongoRepository) FindByNome(ctx context.Context, nome string) (*domain.Classe, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{"value.nome": nome}

	var doc bson.M
	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("failed to find classe by nome %s: %w", nome, err)
	}

	return extractClasseFromDocument(doc)
}

// FindSpellcasterClasses retrieves classes that can cast spells
func (r *ClasseMongoRepository) FindSpellcasterClasses(ctx context.Context, limit int) ([]*domain.Classe, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"value.magia.ha_incantesimi": true},
			{"value.magia.haIncantesimi": true},
			{"value.magic.has_spells": true},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find spellcaster classes: %w", err)
	}
	defer cursor.Close(ctx)

	var classi []*domain.Classe
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}
		
		classe, err := extractClasseFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract classe: %w", err)
		}
		classi = append(classi, classe)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return classi, nil
}

// FindByHitDie retrieves classes by hit die size
func (r *ClasseMongoRepository) FindByHitDie(ctx context.Context, hitDie int, limit int) ([]*domain.Classe, error) {
	collection := r.client.GetCollection(r.collectionName)

	// Convert hit die number to string representations (D4, D6, D8, D10, D12, D20)
	hitDieStr := "D" + strconv.Itoa(hitDie)

	filter := bson.M{
		"$or": []bson.M{
			{"value.dado_vita": hitDieStr},
			{"value.dadoVita": hitDieStr},
			{"value.hit_die": hitDieStr},
			{"value.dado_vita": primitive.Regex{Pattern: strconv.Itoa(hitDie), Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find classes by hit die %d: %w", hitDie, err)
	}
	defer cursor.Close(ctx)

	var classi []*domain.Classe
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}
		
		classe, err := extractClasseFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract classe: %w", err)
		}
		classi = append(classi, classe)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return classi, nil
}

// FindByPrimaryAbility retrieves classes by primary ability score
func (r *ClasseMongoRepository) FindByPrimaryAbility(ctx context.Context, ability string, limit int) ([]*domain.Classe, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"value.caratteristica_primaria": bson.M{"$in": []string{ability}}},
			{"value.caratteristicaPrimaria": bson.M{"$in": []string{ability}}},
			{"value.primary_ability": bson.M{"$in": []string{ability}}},
			{"value.caratteristica_primaria.nome": ability},
			{"value.caratteristicaPrimaria.nome": ability},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find classes by primary ability %s: %w", ability, err)
	}
	defer cursor.Close(ctx)

	var classi []*domain.Classe
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}
		
		classe, err := extractClasseFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract classe: %w", err)
		}
		classi = append(classi, classe)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return classi, nil
}

// FindBySavingThrowProficiency retrieves classes by saving throw proficiencies
func (r *ClasseMongoRepository) FindBySavingThrowProficiency(ctx context.Context, savingThrow string, limit int) ([]*domain.Classe, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"value.salvezze_competenze": bson.M{"$in": []string{savingThrow}}},
			{"value.salvezzeCompetenze": bson.M{"$in": []string{savingThrow}}},
			{"value.saving_throw_proficiencies": bson.M{"$in": []string{savingThrow}}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find classes by saving throw %s: %w", savingThrow, err)
	}
	defer cursor.Close(ctx)

	var classi []*domain.Classe
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}
		
		classe, err := extractClasseFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract classe: %w", err)
		}
		classi = append(classi, classe)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return classi, nil
}

// FindMulticlassEligible retrieves classes with multiclass prerequisites
func (r *ClasseMongoRepository) FindMulticlassEligible(ctx context.Context, limit int) ([]*domain.Classe, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"value.multiclasse.prerequisiti": bson.M{"$exists": true, "$ne": []string{}}},
			{"value.multiclasse.prerequisiti": bson.M{"$size": bson.M{"$gt": 0}}},
			{"value.multiclass.prerequisites": bson.M{"$exists": true, "$ne": []string{}}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find multiclass eligible classes: %w", err)
	}
	defer cursor.Close(ctx)

	var classi []*domain.Classe
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}
		
		classe, err := extractClasseFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract classe: %w", err)
		}
		classi = append(classi, classe)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return classi, nil
}