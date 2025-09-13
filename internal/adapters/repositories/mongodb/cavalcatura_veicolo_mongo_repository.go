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

// CavalcaturaVeicoloMongoRepository implements repositories.CavalcaturaVeicoloRepository using MongoDB
type CavalcaturaVeicoloMongoRepository struct {
	*BaseMongoRepository[*domain.CavalcaturaVeicolo]
}

// NewCavalcaturaVeicoloMongoRepository creates a new MongoDB CavalcaturaVeicolo repository
func NewCavalcaturaVeicoloMongoRepository(client *mongodb.Client) repositories.CavalcaturaVeicoloRepository {
	base := NewBaseMongoRepository[*domain.CavalcaturaVeicolo](
		client,
		"cavalcature_veicoli",
		[]string{"nome", "slug"},
	)
	
	return &CavalcaturaVeicoloMongoRepository{
		BaseMongoRepository: base,
	}
}

// extractCavalcaturaVeicoloFromDocument extracts CavalcaturaVeicolo from the flattened document
func extractCavalcaturaVeicoloFromDocument(doc bson.M) (*domain.CavalcaturaVeicolo, error) {
	docBytes, err := bson.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal document: %w", err)
	}

	var cavalcaturaVeicolo domain.CavalcaturaVeicolo
	err = bson.Unmarshal(docBytes, &cavalcaturaVeicolo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cavalcatura_veicolo: %w", err)
	}

	return &cavalcaturaVeicolo, nil
}

// FindByNome retrieves a mount/vehicle by its name
func (r *CavalcaturaVeicoloMongoRepository) FindByNome(ctx context.Context, nome string) (*domain.CavalcaturaVeicolo, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{"nome": nome}

	var doc bson.M
	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("failed to find cavalcatura_veicolo by nome %s: %w", nome, err)
	}

	return extractCavalcaturaVeicoloFromDocument(doc)
}

// FindByType retrieves mounts/vehicles by type
func (r *CavalcaturaVeicoloMongoRepository) FindByType(ctx context.Context, vehicleType string, limit int) ([]*domain.CavalcaturaVeicolo, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"tipo": primitive.Regex{Pattern: vehicleType, Options: "i"}},
			{"type": primitive.Regex{Pattern: vehicleType, Options: "i"}},
			{"categoria": primitive.Regex{Pattern: vehicleType, Options: "i"}},
			{"category": primitive.Regex{Pattern: vehicleType, Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find cavalcature_veicoli by type %s: %w", vehicleType, err)
	}
	defer cursor.Close(ctx)

	var cavalcatureVeicoli []*domain.CavalcaturaVeicolo
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		cavalcaturaVeicolo, err := extractCavalcaturaVeicoloFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract cavalcatura_veicolo: %w", err)
		}
		cavalcatureVeicoli = append(cavalcatureVeicoli, cavalcaturaVeicolo)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return cavalcatureVeicoli, nil
}

// FindBySpeed retrieves mounts/vehicles by speed
func (r *CavalcaturaVeicoloMongoRepository) FindBySpeed(ctx context.Context, minSpeed int, limit int) ([]*domain.CavalcaturaVeicolo, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"velocita": bson.M{"$gte": minSpeed}},
			{"speed": bson.M{"$gte": minSpeed}},
			{"contenuto": primitive.Regex{Pattern: fmt.Sprintf("%d", minSpeed), Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find cavalcature_veicoli by speed %d: %w", minSpeed, err)
	}
	defer cursor.Close(ctx)

	var cavalcatureVeicoli []*domain.CavalcaturaVeicolo
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		cavalcaturaVeicolo, err := extractCavalcaturaVeicoloFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract cavalcatura_veicolo: %w", err)
		}
		cavalcatureVeicoli = append(cavalcatureVeicoli, cavalcaturaVeicolo)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return cavalcatureVeicoli, nil
}

// FindByCapacity retrieves vehicles by carrying capacity
func (r *CavalcaturaVeicoloMongoRepository) FindByCapacity(ctx context.Context, minCapacity int, limit int) ([]*domain.CavalcaturaVeicolo, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{
		"$or": []bson.M{
			{"capacita": bson.M{"$gte": minCapacity}},
			{"capacity": bson.M{"$gte": minCapacity}},
			{"carico": bson.M{"$gte": minCapacity}},
			{"contenuto": primitive.Regex{Pattern: fmt.Sprintf("%d", minCapacity), Options: "i"}},
		},
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find cavalcature_veicoli by capacity %d: %w", minCapacity, err)
	}
	defer cursor.Close(ctx)

	var cavalcatureVeicoli []*domain.CavalcaturaVeicolo
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}

		cavalcaturaVeicolo, err := extractCavalcaturaVeicoloFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to extract cavalcatura_veicolo: %w", err)
		}
		cavalcatureVeicoli = append(cavalcatureVeicoli, cavalcaturaVeicolo)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return cavalcatureVeicoli, nil
}