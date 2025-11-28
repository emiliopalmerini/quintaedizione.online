package repositories

import (
	"context"

	"github.com/emiliopalmerini/quintaedizione.online/internal/adapters/repositories/mongodb"
	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/repositories"
	pkgMongodb "github.com/emiliopalmerini/quintaedizione.online/pkg/mongodb"
)

type RepositoryFactory struct {
	client *pkgMongodb.Client

	documentRepo repositories.DocumentRepository
}

func NewRepositoryFactory(client *pkgMongodb.Client) *RepositoryFactory {
	return &RepositoryFactory{
		client: client,
	}
}

func (f *RepositoryFactory) DocumentRepository() repositories.DocumentRepository {
	if f.documentRepo == nil {
		f.documentRepo = mongodb.NewDocumentMongoRepository(f.client)
	}
	return f.documentRepo
}

type ParserRepositoryWrapper struct {
	factory *RepositoryFactory
}

func NewParserRepositoryWrapper(factory *RepositoryFactory) *ParserRepositoryWrapper {
	return &ParserRepositoryWrapper{
		factory: factory,
	}
}

func (w *ParserRepositoryWrapper) UpsertMany(collection string, uniqueFields []string, docs []map[string]any) (int, error) {
	return w.factory.DocumentRepository().UpsertManyMaps(context.Background(), collection, uniqueFields, docs)
}

func (w *ParserRepositoryWrapper) Count(collection string) (int64, error) {
	return w.factory.DocumentRepository().Count(context.Background(), collection)
}
