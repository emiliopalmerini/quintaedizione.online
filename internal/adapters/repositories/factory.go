package repositories

import (
	"context"

	"github.com/emiliopalmerini/quintaedizione.online/internal/adapters/repositories/mongodb"
	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/repositories"
	pkgMongodb "github.com/emiliopalmerini/quintaedizione.online/pkg/mongodb"
)

// RepositoryFactory creates and manages all repository instances
type RepositoryFactory struct {
	client *pkgMongodb.Client

	// Repository instances
	documentRepo repositories.DocumentRepository
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(client *pkgMongodb.Client) *RepositoryFactory {
	return &RepositoryFactory{
		client: client,
	}
}

// DocumentRepository returns the unified document repository
func (f *RepositoryFactory) DocumentRepository() repositories.DocumentRepository {
	if f.documentRepo == nil {
		f.documentRepo = mongodb.NewDocumentMongoRepository(f.client)
	}
	return f.documentRepo
}

// ParserRepositoryWrapper provides legacy compatibility for old entity-based parsers
// Routes operations to DocumentRepository with collection-based routing
type ParserRepositoryWrapper struct {
	factory *RepositoryFactory
}

// NewParserRepositoryWrapper creates a wrapper that implements the legacy ParserRepository interface
func NewParserRepositoryWrapper(factory *RepositoryFactory) *ParserRepositoryWrapper {
	return &ParserRepositoryWrapper{
		factory: factory,
	}
}

// UpsertMany routes to DocumentRepository.UpsertManyMaps
func (w *ParserRepositoryWrapper) UpsertMany(collection string, uniqueFields []string, docs []map[string]any) (int, error) {
	return w.factory.DocumentRepository().UpsertManyMaps(context.Background(), collection, uniqueFields, docs)
}

// Count routes to DocumentRepository.Count
func (w *ParserRepositoryWrapper) Count(collection string) (int64, error) {
	return w.factory.DocumentRepository().Count(context.Background(), collection)
}
