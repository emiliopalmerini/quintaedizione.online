package repositories

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

// ContentRepository provides unified operations across all content collections
type ContentRepository interface {
	// GetCollectionItems retrieves items from any collection with pagination and search
	GetCollectionItems(ctx context.Context, collection, search string, skip int64, limit int64) ([]map[string]any, int64, error)
	
	// GetItemBySlug retrieves a specific item by slug from any collection
	GetItemBySlug(ctx context.Context, collection, slug string) (map[string]any, error)
	
	// GetCollectionStats retrieves statistics for all collections
	GetCollectionStats(ctx context.Context) ([]map[string]any, error)
	
	// CountCollection counts items in a specific collection with optional filter
	CountCollection(ctx context.Context, collection string, filter bson.M) (int64, error)
	
	// FindCollectionMaps finds documents as maps in a collection
	FindCollectionMaps(ctx context.Context, collection string, filter bson.M, skip, limit int64) ([]map[string]any, error)
	
	// GetAdjacentItems gets the previous and next items for navigation
	GetAdjacentItems(ctx context.Context, collection, currentSlug string) (prevSlug, nextSlug *string, err error)
}