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
	case "backgrounds":
		return w.factory.BackgroundRepository().UpsertManyMaps(context.Background(), uniqueFields, docs)
	case "incantesimi":
		return w.factory.IncantesimoRepository().UpsertManyMaps(context.Background(), uniqueFields, docs)
	case "talenti":
		return w.factory.TalentoRepository().UpsertManyMaps(context.Background(), uniqueFields, docs)
	case "equipaggiamenti":
		return w.factory.EquipaggiamentoRepository().UpsertManyMaps(context.Background(), uniqueFields, docs)
	case "servizi":
		return w.factory.ServizioRepository().UpsertManyMaps(context.Background(), uniqueFields, docs)
	case "strumenti":
		return w.factory.StrumentoRepository().UpsertManyMaps(context.Background(), uniqueFields, docs)
	case "regole":
		return w.factory.RegolaRepository().UpsertManyMaps(context.Background(), uniqueFields, docs)
	case "cavalcature_veicoli":
		return w.factory.CavalcaturaVeicoloRepository().UpsertManyMaps(context.Background(), uniqueFields, docs)
	case "oggetti_magici":
		return w.factory.OggettoMagicoRepository().UpsertManyMaps(context.Background(), uniqueFields, docs)
	case "mostri":
		return w.factory.MostroRepository().UpsertManyMaps(context.Background(), uniqueFields, docs)
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
	case "backgrounds":
		return w.factory.BackgroundRepository().Count(context.Background())
	case "incantesimi":
		return w.factory.IncantesimoRepository().Count(context.Background())
	case "talenti":
		return w.factory.TalentoRepository().Count(context.Background())
	case "equipaggiamenti":
		return w.factory.EquipaggiamentoRepository().Count(context.Background())
	case "servizi":
		return w.factory.ServizioRepository().Count(context.Background())
	case "strumenti":
		return w.factory.StrumentoRepository().Count(context.Background())
	case "regole":
		return w.factory.RegolaRepository().Count(context.Background())
	case "cavalcature_veicoli":
		return w.factory.CavalcaturaVeicoloRepository().Count(context.Background())
	case "oggetti_magici":
		return w.factory.OggettoMagicoRepository().Count(context.Background())
	case "mostri":
		return w.factory.MostroRepository().Count(context.Background())
	default:
		return 0, fmt.Errorf("unsupported collection: %s", collection)
	}
}
