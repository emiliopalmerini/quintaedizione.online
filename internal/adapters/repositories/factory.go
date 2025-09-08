package repositories

import (
	"context"
	"fmt"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/adapters/repositories/mongodb"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain/repositories"
	pkgMongodb "github.com/emiliopalmerini/due-draghi-5e-srd/pkg/mongodb"
)

// RepositoryFactory creates and manages all repository instances
type RepositoryFactory struct {
	client *pkgMongodb.Client

	// Repository instances
	documentoRepo          repositories.DocumentoRepository
	incantesimoRepo        repositories.IncantesimoRepository
	mostroRepo             repositories.MostroRepository
	classeRepo             repositories.ClasseRepository
	backgroundRepo         repositories.BackgroundRepository
	armaRepo               repositories.ArmaRepository
	armaturaRepo           repositories.ArmaturaRepository
	equipaggiamentoRepo    repositories.EquipaggiamentoRepository
	oggettoMagicoRepo      repositories.OggettoMagicoRepository
	talentoRepo            repositories.TalentoRepository
	regolaRepo             repositories.RegolaRepository
	specieRepo             repositories.SpecieRepository
	animaleRepo            repositories.AnimaleRepository
	cavalcaturaVeicoloRepo repositories.CavalcaturaVeicoloRepository
	strumentoRepo          repositories.StrumentoRepository
	servizioRepo           repositories.ServizioRepository
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(client *pkgMongodb.Client) *RepositoryFactory {
	return &RepositoryFactory{
		client: client,
	}
}

// DocumentoRepository returns the Documento repository
func (f *RepositoryFactory) DocumentoRepository() repositories.DocumentoRepository {
	if f.documentoRepo == nil {
		f.documentoRepo = mongodb.NewDocumentoMongoRepository(f.client)
	}
	return f.documentoRepo
}

// IncantesimoRepository returns the Incantesimo repository
func (f *RepositoryFactory) IncantesimoRepository() repositories.IncantesimoRepository {
	if f.incantesimoRepo == nil {
		f.incantesimoRepo = mongodb.NewIncantesimoMongoRepository(f.client)
	}
	return f.incantesimoRepo
}

// MostroRepository returns the Mostro repository
func (f *RepositoryFactory) MostroRepository() repositories.MostroRepository {
	if f.mostroRepo == nil {
		f.mostroRepo = mongodb.NewMostroMongoRepository(f.client)
	}
	return f.mostroRepo
}

// GetRepositoryByEntityType returns the appropriate repository for a given entity type
func (f *RepositoryFactory) GetRepositoryByEntityType(entityType string) interface{} {
	switch entityType {
	case "documento":
		return f.DocumentoRepository()
	case "incantesimo":
		return f.IncantesimoRepository()
	case "mostro":
		return f.MostroRepository()
	default:
		return nil
	}
}

// Legacy ParserRepository interface compatibility
type ParserRepositoryWrapper struct {
	factory *RepositoryFactory
}

// NewParserRepositoryWrapper creates a wrapper that implements the legacy ParserRepository interface
func NewParserRepositoryWrapper(factory *RepositoryFactory) *ParserRepositoryWrapper {
	return &ParserRepositoryWrapper{factory: factory}
}

// UpsertMany routes to the appropriate repository based on collection name
func (w *ParserRepositoryWrapper) UpsertMany(collection string, uniqueFields []string, docs []map[string]any) (int, error) {
	switch collection {
	case "documenti":
		return w.factory.DocumentoRepository().UpsertManyMaps(context.Background(), uniqueFields, docs)
	case "incantesimi":
		return w.factory.IncantesimoRepository().UpsertManyMaps(context.Background(), uniqueFields, docs)
	case "mostri":
		return w.factory.MostroRepository().UpsertManyMaps(context.Background(), uniqueFields, docs)
	default:
		return 0, fmt.Errorf("unsupported collection: %s", collection)
	}
}

// Count routes to the appropriate repository based on collection name
func (w *ParserRepositoryWrapper) Count(collection string) (int64, error) {
	switch collection {
	case "documenti":
		return w.factory.DocumentoRepository().Count(context.Background())
	case "incantesimi":
		return w.factory.IncantesimoRepository().Count(context.Background())
	case "mostri":
		return w.factory.MostroRepository().Count(context.Background())
	default:
		return 0, fmt.Errorf("unsupported collection: %s", collection)
	}
}
