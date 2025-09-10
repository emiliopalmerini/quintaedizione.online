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

// ArmaturaRepository returns the Armatura repository
func (f *RepositoryFactory) ArmaturaRepository() repositories.ArmaturaRepository {
	if f.armaturaRepo == nil {
		f.armaturaRepo = mongodb.NewArmaturaMongoRepository(f.client)
	}
	return f.armaturaRepo
}

// ArmaRepository returns the Arma repository
func (f *RepositoryFactory) ArmaRepository() repositories.ArmaRepository {
	if f.armaRepo == nil {
		f.armaRepo = mongodb.NewArmaMongoRepository(f.client)
	}
	return f.armaRepo
}

// AnimaleRepository returns the Animale repository
func (f *RepositoryFactory) AnimaleRepository() repositories.AnimaleRepository {
	if f.animaleRepo == nil {
		f.animaleRepo = mongodb.NewAnimaleMongoRepository(f.client)
	}
	return f.animaleRepo
}

// ArmaturaRepository returns the Armatura repository
func (f *RepositoryFactory) ClasseRepository() repositories.ArmaturaRepository {
	if f.classeRepo == nil {
		f.classeRepo = mongodb.NewClasseMongoRepository(f.client)
	}
	return f.armaturaRepo
}

// GetRepositoryByEntityType returns the appropriate repository for a given entity type
func (f *RepositoryFactory) GetRepositoryByEntityType(entityType string) any {
	switch entityType {
	case "arma":
		return f.ArmaRepository()
	case "armatura":
		return f.ArmaturaRepository()
	case "animale":
		return f.AnimaleRepository()
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
	case "armature":
		return w.factory.ArmaturaRepository().UpsertManyMaps(context.Background(), uniqueFields, docs)
	case "classi":
		return w.factory.ClasseRepository().UpsertManyMaps(context.Background(), uniqueFields, docs)
	case "armi":
		return w.factory.ArmaRepository().UpsertManyMaps(context.Background(), uniqueFields, docs)
	case "animali":
		return w.factory.AnimaleRepository().UpsertManyMaps(context.Background(), uniqueFields, docs)
	default:
		return 0, fmt.Errorf("unsupported collection: %s", collection)
	}
}

// Count routes to the appropriate repository based on collection name
func (w *ParserRepositoryWrapper) Count(collection string) (int64, error) {
	switch collection {
	case "armature":
		return w.factory.ArmaturaRepository().Count(context.Background())
	case "classi":
		return w.factory.ClasseRepository().Count(context.Background())
	case "armi":
		return w.factory.ArmaRepository().Count(context.Background())
	case "animali":
		return w.factory.AnimaleRepository().Count(context.Background())
	default:
		return 0, fmt.Errorf("unsupported collection: %s", collection)
	}
}
