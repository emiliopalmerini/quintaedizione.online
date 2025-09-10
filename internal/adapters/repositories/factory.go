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

// ClasseRepository returns the Classe repository
func (f *RepositoryFactory) ClasseRepository() repositories.ClasseRepository {
	if f.classeRepo == nil {
		f.classeRepo = mongodb.NewClasseMongoRepository(f.client)
	}
	return f.classeRepo
}

// BackgroundRepository returns the Background repository
func (f *RepositoryFactory) BackgroundRepository() repositories.BackgroundRepository {
	if f.backgroundRepo == nil {
		f.backgroundRepo = mongodb.NewBackgroundMongoRepository(f.client)
	}
	return f.backgroundRepo
}

// ArmaRepository returns the Arma repository
func (f *RepositoryFactory) ArmaRepository() repositories.ArmaRepository {
	if f.armaRepo == nil {
		f.armaRepo = mongodb.NewArmaMongoRepository(f.client)
	}
	return f.armaRepo
}

// ArmaturaRepository returns the Armatura repository
func (f *RepositoryFactory) ArmaturaRepository() repositories.ArmaturaRepository {
	if f.armaturaRepo == nil {
		f.armaturaRepo = mongodb.NewArmaturaMongoRepository(f.client)
	}
	return f.armaturaRepo
}

// EquipaggiamentoRepository returns the Equipaggiamento repository
func (f *RepositoryFactory) EquipaggiamentoRepository() repositories.EquipaggiamentoRepository {
	if f.equipaggiamentoRepo == nil {
		f.equipaggiamentoRepo = mongodb.NewEquipaggiamentoMongoRepository(f.client)
	}
	return f.equipaggiamentoRepo
}

// OggettoMagicoRepository returns the OggettoMagico repository
func (f *RepositoryFactory) OggettoMagicoRepository() repositories.OggettoMagicoRepository {
	if f.oggettoMagicoRepo == nil {
		f.oggettoMagicoRepo = mongodb.NewOggettoMagicoMongoRepository(f.client)
	}
	return f.oggettoMagicoRepo
}

// TalentoRepository returns the Talento repository
func (f *RepositoryFactory) TalentoRepository() repositories.TalentoRepository {
	if f.talentoRepo == nil {
		f.talentoRepo = mongodb.NewTalentoMongoRepository(f.client)
	}
	return f.talentoRepo
}

// RegolaRepository returns the Regola repository
func (f *RepositoryFactory) RegolaRepository() repositories.RegolaRepository {
	if f.regolaRepo == nil {
		f.regolaRepo = mongodb.NewRegolaMongoRepository(f.client)
	}
	return f.regolaRepo
}

// AnimaleRepository returns the Animale repository
func (f *RepositoryFactory) AnimaleRepository() repositories.AnimaleRepository {
	if f.animaleRepo == nil {
		f.animaleRepo = mongodb.NewAnimaleMongoRepository(f.client)
	}
	return f.animaleRepo
}

// GetRepositoryByEntityType returns the appropriate repository for a given entity type
func (f *RepositoryFactory) GetRepositoryByEntityType(entityType string) any {
	switch entityType {
	case "documento":
		return f.DocumentoRepository()
	case "incantesimo":
		return f.IncantesimoRepository()
	case "mostro":
		return f.MostroRepository()
	case "classe":
		return f.ClasseRepository()
	case "background":
		return f.BackgroundRepository()
	case "arma":
		return f.ArmaRepository()
	case "armatura":
		return f.ArmaturaRepository()
	case "equipaggiamento":
		return f.EquipaggiamentoRepository()
	case "oggetto_magico":
		return f.OggettoMagicoRepository()
	case "talento":
		return f.TalentoRepository()
	case "regola":
		return f.RegolaRepository()
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
	default:
		return 0, fmt.Errorf("unsupported collection: %s", collection)
	}
}

// Count routes to the appropriate repository based on collection name
func (w *ParserRepositoryWrapper) Count(collection string) (int64, error) {
	switch collection {
	case "armature":
		return w.factory.ArmaturaRepository().Count(context.Background())
	default:
		return 0, fmt.Errorf("unsupported collection: %s", collection)
	}
}
