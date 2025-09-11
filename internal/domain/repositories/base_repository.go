package repositories

import (
	"context"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
)

// Entity represents any domain entity with basic identification
type Entity interface {
	domain.ParsedEntity
}

// BaseRepository defines common repository operations for any entity type
type BaseRepository[T Entity] interface {
	// Create inserts a new entity
	Create(ctx context.Context, entity T) error

	// Update modifies an existing entity
	Update(ctx context.Context, entity T) error

	// Delete removes an entity by its identifier
	Delete(ctx context.Context, id string) error

	// FindByID retrieves an entity by its identifier
	FindByID(ctx context.Context, id string) (*T, error)

	// FindAll retrieves all entities with optional limit
	FindAll(ctx context.Context, limit int) ([]*T, error)

	// FindByFilter retrieves entities matching the filter
	FindByFilter(ctx context.Context, filter bson.M, limit int) ([]*T, error)

	// Count returns the total number of entities
	Count(ctx context.Context) (int64, error)

	// UpsertMany performs bulk upsert operations
	UpsertMany(ctx context.Context, entities []T) (int, error)

	// UpsertManyMaps performs bulk upsert operations from raw maps (for parser compatibility)
	UpsertManyMaps(ctx context.Context, uniqueFields []string, docs []map[string]any) (int, error)
}
