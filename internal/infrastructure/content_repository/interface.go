package content_repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
)

type Repository interface {

	FindMaps(ctx context.Context, collection string, filter bson.M, skip, limit int64) ([]map[string]any, error)

	FindOneMap(ctx context.Context, collection string, filter bson.M) (map[string]any, error)

	Count(ctx context.Context, collection string, filter bson.M) (int64, error)

	GetCollectionStats(ctx context.Context) ([]map[string]any, error)
}
