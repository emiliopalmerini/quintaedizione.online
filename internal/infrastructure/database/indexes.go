package database

import (
	"context"
	"fmt"
	"log"
	"time"

	pkgMongodb "github.com/emiliopalmerini/quintaedizione.online/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IndexManager struct {
	client *pkgMongodb.Client
}

func NewIndexManager(client *pkgMongodb.Client) *IndexManager {
	return &IndexManager{client: client}
}

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

func (im *IndexManager) createCollectionIndexes(ctx context.Context, collectionName string) error {
	collection := im.client.GetCollection(collectionName)

	if _, err := collection.Indexes().DropOne(ctx, "text_search"); err != nil {
		log.Printf("Note: could not drop text_search index on %s (may not exist): %v", collectionName, err)
	}

	commonIndexes := []mongo.IndexModel{

		{
			Keys:    bson.D{{Key: "source_file", Value: 1}},
			Options: options.Index().SetName("source_file_1").SetBackground(true),
		},
	}

	specificIndexes := im.getCollectionSpecificIndexes(collectionName)

	allIndexes := append(commonIndexes, specificIndexes...)

	if len(allIndexes) > 0 {
		_, err := collection.Indexes().CreateMany(ctx, allIndexes)
		if err != nil {
			return fmt.Errorf("failed to create indexes: %w", err)
		}
	}

	return nil
}

func (im *IndexManager) getCollectionSpecificIndexes(collectionName string) []mongo.IndexModel {
	switch collectionName {
	case "incantesimi":
		return []mongo.IndexModel{

			{
				Keys:    bson.D{{Key: "livello", Value: 1}},
				Options: options.Index().SetName("livello_1").SetBackground(true),
			},
			{
				Keys:    bson.D{{Key: "scuola", Value: 1}},
				Options: options.Index().SetName("scuola_1").SetBackground(true),
			},
			{
				Keys: bson.D{
					{Key: "livello", Value: 1},
					{Key: "scuola", Value: 1},
				},
				Options: options.Index().SetName("livello_scuola_1").SetBackground(true),
			},
			{
				Keys:    bson.D{{Key: "classi", Value: 1}},
				Options: options.Index().SetName("classi_1").SetBackground(true),
			},

			{
				Keys: bson.D{
					{Key: "title", Value: "text"},
					{Key: "content", Value: "text"},
					{Key: "raw_content", Value: "text"},
					{Key: "filters.scuola", Value: "text"},
					{Key: "filters.livello", Value: "text"},
					{Key: "filters.classe", Value: "text"},
				},
				Options: options.Index().SetName("text_search").SetBackground(true).SetDefaultLanguage("none").
					SetWeights(bson.M{"title": 10, "content": 1, "raw_content": 1, "filters.scuola": 1, "filters.livello": 1, "filters.classe": 1}),
			},
		}

	case "mostri", "animali":
		return []mongo.IndexModel{

			{
				Keys:    bson.D{{Key: "gs", Value: 1}},
				Options: options.Index().SetName("gs_1").SetBackground(true),
			},
			{
				Keys:    bson.D{{Key: "cr", Value: 1}},
				Options: options.Index().SetName("cr_1").SetBackground(true),
			},
			{
				Keys:    bson.D{{Key: "grado_sfida", Value: 1}},
				Options: options.Index().SetName("grado_sfida_1").SetBackground(true),
			},
			{
				Keys:    bson.D{{Key: "tipo", Value: 1}},
				Options: options.Index().SetName("tipo_1").SetBackground(true),
			},
			{
				Keys:    bson.D{{Key: "taglia", Value: 1}},
				Options: options.Index().SetName("taglia_1").SetBackground(true),
			},
			{
				Keys: bson.D{
					{Key: "tipo", Value: 1},
					{Key: "taglia", Value: 1},
				},
				Options: options.Index().SetName("tipo_taglia_1").SetBackground(true),
			},

			{
				Keys: bson.D{
					{Key: "title", Value: "text"},
					{Key: "content", Value: "text"},
					{Key: "raw_content", Value: "text"},
					{Key: "filters.tipo", Value: "text"},
					{Key: "filters.taglia", Value: "text"},
					{Key: "filters.ambiente", Value: "text"},
					{Key: "filters.allineamento", Value: "text"},
				},
				Options: options.Index().SetName("text_search").SetBackground(true).SetDefaultLanguage("none").
					SetWeights(bson.M{"title": 10, "content": 1, "raw_content": 1, "filters.tipo": 1, "filters.taglia": 1, "filters.ambiente": 1, "filters.allineamento": 1}),
			},
		}

	case "armi":
		return []mongo.IndexModel{

			{
				Keys:    bson.D{{Key: "categoria", Value: 1}},
				Options: options.Index().SetName("categoria_1").SetBackground(true),
			},
			{
				Keys:    bson.D{{Key: "tipo_danno", Value: 1}},
				Options: options.Index().SetName("tipo_danno_1").SetBackground(true),
			},

			{
				Keys: bson.D{
					{Key: "title", Value: "text"},
					{Key: "content", Value: "text"},
					{Key: "raw_content", Value: "text"},
					{Key: "filters.categoria", Value: "text"},
					{Key: "filters.tipo_danno", Value: "text"},
					{Key: "filters.proprieta", Value: "text"},
				},
				Options: options.Index().SetName("text_search").SetBackground(true).SetDefaultLanguage("none").
					SetWeights(bson.M{"title": 10, "content": 1, "raw_content": 1, "filters.categoria": 1, "filters.tipo_danno": 1, "filters.proprieta": 1}),
			},
		}

	case "armature":
		return []mongo.IndexModel{

			{
				Keys:    bson.D{{Key: "categoria", Value: 1}},
				Options: options.Index().SetName("categoria_1").SetBackground(true),
			},
			{
				Keys:    bson.D{{Key: "ca_base", Value: 1}},
				Options: options.Index().SetName("ca_base_1").SetBackground(true),
			},

			{
				Keys: bson.D{
					{Key: "title", Value: "text"},
					{Key: "content", Value: "text"},
					{Key: "raw_content", Value: "text"},
					{Key: "filters.categoria", Value: "text"},
					{Key: "filters.tipo", Value: "text"},
				},
				Options: options.Index().SetName("text_search").SetBackground(true).SetDefaultLanguage("none").
					SetWeights(bson.M{"title": 10, "content": 1, "raw_content": 1, "filters.categoria": 1, "filters.tipo": 1}),
			},
		}

	case "oggetti_magici":
		return []mongo.IndexModel{

			{
				Keys:    bson.D{{Key: "rarita", Value: 1}},
				Options: options.Index().SetName("rarita_1").SetBackground(true),
			},
			{
				Keys:    bson.D{{Key: "tipo", Value: 1}},
				Options: options.Index().SetName("tipo_1").SetBackground(true),
			},

			{
				Keys: bson.D{
					{Key: "title", Value: "text"},
					{Key: "content", Value: "text"},
					{Key: "raw_content", Value: "text"},
					{Key: "filters.tipo", Value: "text"},
					{Key: "filters.rarita", Value: "text"},
					{Key: "filters.sintonia", Value: "text"},
				},
				Options: options.Index().SetName("text_search").SetBackground(true).SetDefaultLanguage("none").
					SetWeights(bson.M{"title": 10, "content": 1, "raw_content": 1, "filters.tipo": 1, "filters.rarita": 1, "filters.sintonia": 1}),
			},
		}

	case "talenti":
		return []mongo.IndexModel{

			{
				Keys:    bson.D{{Key: "categoria", Value: 1}},
				Options: options.Index().SetName("categoria_1").SetBackground(true),
			},

			{
				Keys: bson.D{
					{Key: "title", Value: "text"},
					{Key: "content", Value: "text"},
					{Key: "raw_content", Value: "text"},
					{Key: "filters.categoria", Value: "text"},
					{Key: "filters.prerequisiti", Value: "text"},
				},
				Options: options.Index().SetName("text_search").SetBackground(true).SetDefaultLanguage("none").
					SetWeights(bson.M{"title": 10, "content": 1, "raw_content": 1, "filters.categoria": 1, "filters.prerequisiti": 1}),
			},
		}

	default:

		return []mongo.IndexModel{

			{
				Keys:    bson.D{{Key: "categoria", Value: 1}},
				Options: options.Index().SetName("categoria_1").SetBackground(true),
			},

			{
				Keys: bson.D{
					{Key: "title", Value: "text"},
					{Key: "content", Value: "text"},
					{Key: "raw_content", Value: "text"},
					{Key: "filters.categoria", Value: "text"},
					{Key: "filters.tipo", Value: "text"},
				},
				Options: options.Index().SetName("text_search").SetBackground(true).SetDefaultLanguage("none").
					SetWeights(bson.M{"title": 10, "content": 1, "raw_content": 1, "filters.categoria": 1, "filters.tipo": 1}),
			},
		}
	}
}

func (im *IndexManager) DropIndexes(ctx context.Context) error {
	collections := []string{
		"incantesimi", "mostri", "classi", "backgrounds", "equipaggiamenti",
		"oggetti_magici", "armi", "armature", "talenti", "servizi",
		"strumenti", "animali", "regole", "cavalcature_veicoli",
	}

	for _, collectionName := range collections {
		collection := im.client.GetCollection(collectionName)

		indexView := collection.Indexes()
		cursor, err := indexView.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list indexes for %s: %w", collectionName, err)
		}
		defer cursor.Close(ctx)

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
			"count":   len(indexes),
			"indexes": indexes,
		}
		totalIndexes += len(indexes)
	}

	stats["total_indexes"] = totalIndexes
	stats["generated_at"] = time.Now()

	return stats, nil
}
