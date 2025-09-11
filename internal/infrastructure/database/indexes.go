package database

import (
	"context"
	"fmt"
	"log"
	"time"

	pkgMongodb "github.com/emiliopalmerini/due-draghi-5e-srd/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// IndexManager handles database index creation and management
type IndexManager struct {
	client *pkgMongodb.Client
}

// NewIndexManager creates a new index manager
func NewIndexManager(client *pkgMongodb.Client) *IndexManager {
	return &IndexManager{client: client}
}

// EnsureIndexes creates all necessary indexes for optimal performance
func (im *IndexManager) EnsureIndexes(ctx context.Context) error {
	collections := []string{
		"incantesimi", "mostri", "classi", "backgrounds", "equipaggiamenti",
		"oggetti_magici", "armi", "armature", "talenti", "servizi",
		"strumenti", "animali", "regole", "cavalcature_veicoli",
	}

	for _, collectionName := range collections {
		if err := im.createCollectionIndexes(ctx, collectionName); err != nil {
			return fmt.Errorf("failed to create indexes for %s: %w", collectionName, err)
		}
		log.Printf("✅ Indexes created for collection: %s", collectionName)
	}

	return nil
}

// createCollectionIndexes creates indexes for a specific collection
func (im *IndexManager) createCollectionIndexes(ctx context.Context, collectionName string) error {
	collection := im.client.GetCollection(collectionName)
	
	// Common indexes for all collections
	commonIndexes := []mongo.IndexModel{
		// Primary search fields
		{
			Keys: bson.D{{Key: "value.nome", Value: 1}},
			Options: options.Index().SetName("nome_1").SetBackground(true),
		},
		{
			Keys: bson.D{{Key: "value.slug", Value: 1}},
			Options: options.Index().SetName("slug_1").SetUnique(true).SetBackground(true),
		},
		// Text search index for full-text search
		{
			Keys: bson.D{
				{Key: "value.nome", Value: "text"},
				{Key: "contenuto", Value: "text"},
				{Key: "value.descrizione", Value: "text"},
			},
			Options: options.Index().SetName("text_search").SetBackground(true),
		},
		// Source file index for administrative queries
		{
			Keys: bson.D{{Key: "source_file", Value: 1}},
			Options: options.Index().SetName("source_file_1").SetBackground(true),
		},
	}

	// Collection-specific indexes
	specificIndexes := im.getCollectionSpecificIndexes(collectionName)
	
	// Combine all indexes
	allIndexes := append(commonIndexes, specificIndexes...)
	
	if len(allIndexes) > 0 {
		_, err := collection.Indexes().CreateMany(ctx, allIndexes)
		if err != nil {
			return fmt.Errorf("failed to create indexes: %w", err)
		}
	}

	return nil
}

// getCollectionSpecificIndexes returns indexes specific to each collection type
func (im *IndexManager) getCollectionSpecificIndexes(collectionName string) []mongo.IndexModel {
	switch collectionName {
	case "incantesimi":
		return []mongo.IndexModel{
			{
				Keys: bson.D{{Key: "value.livello", Value: 1}},
				Options: options.Index().SetName("livello_1").SetBackground(true),
			},
			{
				Keys: bson.D{{Key: "value.scuola", Value: 1}},
				Options: options.Index().SetName("scuola_1").SetBackground(true),
			},
			{
				Keys: bson.D{
					{Key: "value.livello", Value: 1},
					{Key: "value.scuola", Value: 1},
				},
				Options: options.Index().SetName("livello_scuola_1").SetBackground(true),
			},
			{
				Keys: bson.D{{Key: "value.classi", Value: 1}},
				Options: options.Index().SetName("classi_1").SetBackground(true),
			},
		}

	case "mostri", "animali":
		return []mongo.IndexModel{
			{
				Keys: bson.D{{Key: "value.gs", Value: 1}},
				Options: options.Index().SetName("gs_1").SetBackground(true),
			},
			{
				Keys: bson.D{{Key: "value.cr", Value: 1}},
				Options: options.Index().SetName("cr_1").SetBackground(true),
			},
			{
				Keys: bson.D{{Key: "value.grado_sfida", Value: 1}},
				Options: options.Index().SetName("grado_sfida_1").SetBackground(true),
			},
			{
				Keys: bson.D{{Key: "value.tipo", Value: 1}},
				Options: options.Index().SetName("tipo_1").SetBackground(true),
			},
			{
				Keys: bson.D{{Key: "value.taglia", Value: 1}},
				Options: options.Index().SetName("taglia_1").SetBackground(true),
			},
			{
				Keys: bson.D{
					{Key: "value.tipo", Value: 1},
					{Key: "value.taglia", Value: 1},
				},
				Options: options.Index().SetName("tipo_taglia_1").SetBackground(true),
			},
		}

	case "armi":
		return []mongo.IndexModel{
			{
				Keys: bson.D{{Key: "value.categoria", Value: 1}},
				Options: options.Index().SetName("categoria_1").SetBackground(true),
			},
			{
				Keys: bson.D{{Key: "value.tipo_danno", Value: 1}},
				Options: options.Index().SetName("tipo_danno_1").SetBackground(true),
			},
		}

	case "armature":
		return []mongo.IndexModel{
			{
				Keys: bson.D{{Key: "value.categoria", Value: 1}},
				Options: options.Index().SetName("categoria_1").SetBackground(true),
			},
			{
				Keys: bson.D{{Key: "value.ca_base", Value: 1}},
				Options: options.Index().SetName("ca_base_1").SetBackground(true),
			},
		}

	case "oggetti_magici":
		return []mongo.IndexModel{
			{
				Keys: bson.D{{Key: "value.rarita", Value: 1}},
				Options: options.Index().SetName("rarita_1").SetBackground(true),
			},
			{
				Keys: bson.D{{Key: "value.tipo", Value: 1}},
				Options: options.Index().SetName("tipo_1").SetBackground(true),
			},
		}

	case "talenti":
		return []mongo.IndexModel{
			{
				Keys: bson.D{{Key: "value.categoria", Value: 1}},
				Options: options.Index().SetName("categoria_1").SetBackground(true),
			},
		}

	default:
		// Default indexes for other collections
		return []mongo.IndexModel{
			{
				Keys: bson.D{{Key: "value.categoria", Value: 1}},
				Options: options.Index().SetName("categoria_1").SetBackground(true),
			},
		}
	}
}

// DropIndexes removes all custom indexes (useful for testing/migration)
func (im *IndexManager) DropIndexes(ctx context.Context) error {
	collections := []string{
		"incantesimi", "mostri", "classi", "backgrounds", "equipaggiamenti",
		"oggetti_magici", "armi", "armature", "talenti", "servizi",
		"strumenti", "animali", "regole", "cavalcature_veicoli",
	}

	for _, collectionName := range collections {
		collection := im.client.GetCollection(collectionName)
		
		// List all indexes
		indexView := collection.Indexes()
		cursor, err := indexView.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list indexes for %s: %w", collectionName, err)
		}
		defer cursor.Close(ctx)

		// Drop all indexes except _id_
		var indexes []bson.M
		if err = cursor.All(ctx, &indexes); err != nil {
			return fmt.Errorf("failed to decode indexes for %s: %w", collectionName, err)
		}

		for _, index := range indexes {
			if name, ok := index["name"].(string); ok && name != "_id_" {
				if _, err := indexView.DropOne(ctx, name); err != nil {
					log.Printf("Warning: failed to drop index %s on %s: %v", name, collectionName, err)
				}
			}
		}

		log.Printf("✅ Indexes dropped for collection: %s", collectionName)
	}

	return nil
}

// GetIndexStats returns index usage statistics
func (im *IndexManager) GetIndexStats(ctx context.Context) (map[string]any, error) {
	collections := []string{
		"incantesimi", "mostri", "classi", "backgrounds", "equipaggiamenti",
		"oggetti_magici", "armi", "armature", "talenti", "servizi",
		"strumenti", "animali", "regole", "cavalcature_veicoli",
	}

	stats := make(map[string]any)
	totalIndexes := 0

	for _, collectionName := range collections {
		collection := im.client.GetCollection(collectionName)
		
		cursor, err := collection.Indexes().List(ctx)
		if err != nil {
			continue
		}
		defer cursor.Close(ctx)

		var indexes []bson.M
		if err = cursor.All(ctx, &indexes); err != nil {
			continue
		}

		stats[collectionName] = map[string]any{
			"count": len(indexes),
			"indexes": indexes,
		}
		totalIndexes += len(indexes)
	}

	stats["total_indexes"] = totalIndexes
	stats["generated_at"] = time.Now()

	return stats, nil
}