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
		[]string{"value.nome", "value.slug"},
	)
	
	return &CavalcaturaVeicoloMongoRepository{
		BaseMongoRepository: base,
	}
}

// extractCavalcaturaVeicoloFromDocument extracts CavalcaturaVeicolo from the nested value field
func extractCavalcaturaVeicoloFromDocument(doc bson.M) (*domain.CavalcaturaVeicolo, error) {
	valueData, exists := doc["value"]
	if !exists {
		return nil, fmt.Errorf("cavalcatura_veicolo document missing value field")
	}

	valueBytes, err := bson.Marshal(valueData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value data: %w", err)
	}

	var cavalcaturaVeicolo domain.CavalcaturaVeicolo
	err = bson.Unmarshal(valueBytes, &cavalcaturaVeicolo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cavalcatura_veicolo: %w", err)
	}

	return &cavalcaturaVeicolo, nil
}

// FindByNome retrieves a mount/vehicle by its name
func (r *CavalcaturaVeicoloMongoRepository) FindByNome(ctx context.Context, nome string) (*domain.CavalcaturaVeicolo, error) {
	collection := r.client.GetCollection(r.collectionName)

	filter := bson.M{"value.nome": nome}

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
			{"value.tipo": primitive.Regex{Pattern: vehicleType, Options: "i"}},
			{"value.type": primitive.Regex{Pattern: vehicleType, Options: "i"}},
			{"value.categoria": primitive.Regex{Pattern: vehicleType, Options: "i"}},
			{"value.category": primitive.Regex{Pattern: vehicleType, Options: "i"}},
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
			{"value.velocita": bson.M{"$gte": minSpeed}},
			{"value.speed": bson.M{"$gte": minSpeed}},
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
			{"value.capacita": bson.M{"$gte": minCapacity}},
			{"value.capacity": bson.M{"$gte": minCapacity}},
			{"value.carico": bson.M{"$gte": minCapacity}},
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