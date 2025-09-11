package content_repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
)

// Repository provides map-based access to content for web services
// This interface bridges the gap between typed domain repositories and web service needs
type Repository interface {
	// FindMaps retrieves items as maps with pagination and search
	FindMaps(ctx context.Context, collection string, filter bson.M, skip, limit int64) ([]map[string]interface{}, error)
	
	// FindOneMap retrieves a single item as a map
	FindOneMap(ctx context.Context, collection string, filter bson.M) (map[string]interface{}, error)
	
	// Count returns the total number of items matching the filter
	Count(ctx context.Context, collection string, filter bson.M) (int64, error)
	
	// GetCollectionStats returns statistics for all collections
	GetCollectionStats(ctx context.Context) ([]map[string]interface{}, error)
}