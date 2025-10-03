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
	contentRepo            repositories.ContentRepository
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

// IncantesimoRepository returns the Incantesimo repository
func (f *RepositoryFactory) IncantesimoRepository() repositories.IncantesimoRepository {
	if f.incantesimoRepo == nil {
		f.incantesimoRepo = mongodb.NewIncantesimoMongoRepository(f.client)
	}
	return f.incantesimoRepo
}

// TalentoRepository returns the Talento repository
func (f *RepositoryFactory) TalentoRepository() repositories.TalentoRepository {
	if f.talentoRepo == nil {
		f.talentoRepo = mongodb.NewTalentoMongoRepository(f.client)
	}
	return f.talentoRepo
}

// EquipaggiamentoRepository returns the Equipaggiamento repository
func (f *RepositoryFactory) EquipaggiamentoRepository() repositories.EquipaggiamentoRepository {
	if f.equipaggiamentoRepo == nil {
		f.equipaggiamentoRepo = mongodb.NewEquipaggiamentoMongoRepository(f.client)
	}
	return f.equipaggiamentoRepo
}

// ServizioRepository returns the Servizio repository
func (f *RepositoryFactory) ServizioRepository() repositories.ServizioRepository {
	if f.servizioRepo == nil {
		f.servizioRepo = mongodb.NewServizioMongoRepository(f.client)
	}
	return f.servizioRepo
}

// StrumentoRepository returns the Strumento repository
func (f *RepositoryFactory) StrumentoRepository() repositories.StrumentoRepository {
	if f.strumentoRepo == nil {
		f.strumentoRepo = mongodb.NewStrumentoMongoRepository(f.client)
	}
	return f.strumentoRepo
}

// RegolaRepository returns the Regola repository
func (f *RepositoryFactory) RegolaRepository() repositories.RegolaRepository {
	if f.regolaRepo == nil {
		f.regolaRepo = mongodb.NewRegolaMongoRepository(f.client)
	}
	return f.regolaRepo
}

// CavalcaturaVeicoloRepository returns the CavalcaturaVeicolo repository
func (f *RepositoryFactory) CavalcaturaVeicoloRepository() repositories.CavalcaturaVeicoloRepository {
	if f.cavalcaturaVeicoloRepo == nil {
		f.cavalcaturaVeicoloRepo = mongodb.NewCavalcaturaVeicoloMongoRepository(f.client)
	}
	return f.cavalcaturaVeicoloRepo
}

// OggettoMagicoRepository returns the OggettoMagico repository
func (f *RepositoryFactory) OggettoMagicoRepository() repositories.OggettoMagicoRepository {
	if f.oggettoMagicoRepo == nil {
		f.oggettoMagicoRepo = mongodb.NewOggettoMagicoMongoRepository(f.client)
	}
	return f.oggettoMagicoRepo
}

// MostroRepository returns the Mostro repository
func (f *RepositoryFactory) MostroRepository() repositories.MostroRepository {
	if f.mostroRepo == nil {
		f.mostroRepo = mongodb.NewMostroMongoRepository(f.client)
	}
	return f.mostroRepo
}

// ContentRepository returns the unified content repository
func (f *RepositoryFactory) ContentRepository() repositories.ContentRepository {
	if f.contentRepo == nil {
		f.contentRepo = mongodb.NewContentMongoRepository(f.client)
	}
	return f.contentRepo
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
	case "classe":
		return f.ClasseRepository()
	case "background":
		return f.BackgroundRepository()
	case "incantesimo":
		return f.IncantesimoRepository()
	case "talento":
		return f.TalentoRepository()
	case "equipaggiamento":
		return f.EquipaggiamentoRepository()
	case "servizio":
		return f.ServizioRepository()
	case "strumento":
		return f.StrumentoRepository()
	case "regola":
		return f.RegolaRepository()
	case "cavalcatura_veicolo":
		return f.CavalcaturaVeicoloRepository()
	case "oggetto_magico":
		return f.OggettoMagicoRepository()
	case "mostro":
		return f.MostroRepository()
	default:
		return nil
	}
}

// RepositoryOperations defines the minimal interface needed for parser operations
type RepositoryOperations interface {
	UpsertManyMaps(ctx context.Context, uniqueFields []string, docs []map[string]any) (int, error)
	Count(ctx context.Context) (int64, error)
}

// Legacy ParserRepository interface compatibility
type ParserRepositoryWrapper struct {
	factory *RepositoryFactory
	// Collection name to repository getter mapping
	repoGetters map[string]func() RepositoryOperations
}

// NewParserRepositoryWrapper creates a wrapper that implements the legacy ParserRepository interface
func NewParserRepositoryWrapper(factory *RepositoryFactory) *ParserRepositoryWrapper {
	w := &ParserRepositoryWrapper{
		factory:     factory,
		repoGetters: make(map[string]func() RepositoryOperations),
	}

	// Initialize collection-to-repository mapping
	w.repoGetters["armature"] = func() RepositoryOperations { return factory.ArmaturaRepository() }
	w.repoGetters["classi"] = func() RepositoryOperations { return factory.ClasseRepository() }
	w.repoGetters["armi"] = func() RepositoryOperations { return factory.ArmaRepository() }
	w.repoGetters["animali"] = func() RepositoryOperations { return factory.AnimaleRepository() }
	w.repoGetters["backgrounds"] = func() RepositoryOperations { return factory.BackgroundRepository() }
	w.repoGetters["incantesimi"] = func() RepositoryOperations { return factory.IncantesimoRepository() }
	w.repoGetters["talenti"] = func() RepositoryOperations { return factory.TalentoRepository() }
	w.repoGetters["equipaggiamenti"] = func() RepositoryOperations { return factory.EquipaggiamentoRepository() }
	w.repoGetters["servizi"] = func() RepositoryOperations { return factory.ServizioRepository() }
	w.repoGetters["strumenti"] = func() RepositoryOperations { return factory.StrumentoRepository() }
	w.repoGetters["regole"] = func() RepositoryOperations { return factory.RegolaRepository() }
	w.repoGetters["cavalcature_veicoli"] = func() RepositoryOperations { return factory.CavalcaturaVeicoloRepository() }
	w.repoGetters["oggetti_magici"] = func() RepositoryOperations { return factory.OggettoMagicoRepository() }
	w.repoGetters["mostri"] = func() RepositoryOperations { return factory.MostroRepository() }

	return w
}

// getRepository retrieves the repository for a given collection name
func (w *ParserRepositoryWrapper) getRepository(collection string) (RepositoryOperations, error) {
	getter, exists := w.repoGetters[collection]
	if !exists {
		return nil, fmt.Errorf("unsupported collection: %s", collection)
	}
	return getter(), nil
}

// UpsertMany routes to the appropriate repository based on collection name
func (w *ParserRepositoryWrapper) UpsertMany(collection string, uniqueFields []string, docs []map[string]any) (int, error) {
	repo, err := w.getRepository(collection)
	if err != nil {
		return 0, err
	}
	return repo.UpsertManyMaps(context.Background(), uniqueFields, docs)
}

// Count routes to the appropriate repository based on collection name
func (w *ParserRepositoryWrapper) Count(collection string) (int64, error) {
	repo, err := w.getRepository(collection)
	if err != nil {
		return 0, err
	}
	return repo.Count(context.Background())
}
