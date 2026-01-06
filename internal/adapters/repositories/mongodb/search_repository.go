package mongodb

import (
	"context"

	domainsearch "github.com/emiliopalmerini/quintaedizione.online/internal/domain/search"
	pkgMongodb "github.com/emiliopalmerini/quintaedizione.online/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type searchRepository struct {
	client *pkgMongodb.Client
}

func NewSearchRepository(client *pkgMongodb.Client) domainsearch.SearchRepository {
	return &searchRepository{client: client}
}

func (r *searchRepository) GetSearchableItems(ctx context.Context, collection string) ([]domainsearch.SearchableItem, error) {
	coll := r.client.GetDatabase().Collection(collection)

	projection := bson.M{
		"_id":     1,
		"title":   1,
		"filters": 1,
	}

	opts := options.Find().SetProjection(projection)
	cursor, err := coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []domainsearch.SearchableItem
	for cursor.Next(ctx) {
		var doc struct {
			ID      string         `bson:"_id"`
			Title   string         `bson:"title"`
			Filters map[string]any `bson:"filters"`
		}
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		keywords := extractKeywords(doc.Filters)
		items = append(items, domainsearch.SearchableItem{
			ID:         doc.ID,
			Collection: collection,
			Title:      doc.Title,
			Keywords:   keywords,
		})
	}

	return items, nil
}

func extractKeywords(filters map[string]any) []string {
	if filters == nil {
		return nil
	}

	keywords := make([]string, 0)

	keywordFields := []string{
		"scuola", "classe", "tipo", "categoria",
		"rarita", "livello", "taglia", "allineamento",
		"tipo_danno", "proprieta", "ambiente",
	}

	for _, field := range keywordFields {
		val, ok := filters[field]
		if !ok {
			continue
		}

		switch v := val.(type) {
		case string:
			if v != "" {
				keywords = append(keywords, v)
			}
		case []any:
			for _, item := range v {
				if s, ok := item.(string); ok && s != "" {
					keywords = append(keywords, s)
				}
			}
		}
	}

	return keywords
}
