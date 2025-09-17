package repositories

import (
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/adapters/repositories/mongodb"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain/repositories"
	pkgMongodb "github.com/emiliopalmerini/due-draghi-5e-srd/pkg/mongodb"
)

// RegistryBasedRepositoryFactory implements the same interface as the old factory
// but uses the registry pattern internally
type RegistryBasedRepositoryFactory struct {
	client   *pkgMongodb.Client
	registry RepositoryRegistry
}

// NewRegistryBasedRepositoryFactory creates a new registry-based factory
func NewRegistryBasedRepositoryFactory(client *pkgMongodb.Client) *RegistryBasedRepositoryFactory {
	return &RegistryBasedRepositoryFactory{
		client:   client,
		registry: NewRepositoryRegistry(client),
	}
}

// Repository interface methods using registry

func (f *RegistryBasedRepositoryFactory) ArmaRepository() repositories.ArmaRepository {
	repo, _ := f.registry.GetRepository("arma")
	return repo.(repositories.ArmaRepository)
}

func (f *RegistryBasedRepositoryFactory) ArmaturaRepository() repositories.ArmaturaRepository {
	repo, _ := f.registry.GetRepository("armatura")
	return repo.(repositories.ArmaturaRepository)
}

func (f *RegistryBasedRepositoryFactory) AnimaleRepository() repositories.AnimaleRepository {
	repo, _ := f.registry.GetRepository("animale")
	return repo.(repositories.AnimaleRepository)
}

func (f *RegistryBasedRepositoryFactory) BackgroundRepository() repositories.BackgroundRepository {
	repo, _ := f.registry.GetRepository("background")
	return repo.(repositories.BackgroundRepository)
}

func (f *RegistryBasedRepositoryFactory) ClasseRepository() repositories.ClasseRepository {
	repo, _ := f.registry.GetRepository("classe")
	return repo.(repositories.ClasseRepository)
}

func (f *RegistryBasedRepositoryFactory) EquipaggiamentoRepository() repositories.EquipaggiamentoRepository {
	repo, _ := f.registry.GetRepository("equipaggiamento")
	return repo.(repositories.EquipaggiamentoRepository)
}

func (f *RegistryBasedRepositoryFactory) IncantesimoRepository() repositories.IncantesimoRepository {
	repo, _ := f.registry.GetRepository("incantesimo")
	return repo.(repositories.IncantesimoRepository)
}

func (f *RegistryBasedRepositoryFactory) MostroRepository() repositories.MostroRepository {
	repo, _ := f.registry.GetRepository("mostro")
	return repo.(repositories.MostroRepository)
}

func (f *RegistryBasedRepositoryFactory) OggettoMagicoRepository() repositories.OggettoMagicoRepository {
	repo, _ := f.registry.GetRepository("oggetto_magico")
	return repo.(repositories.OggettoMagicoRepository)
}

func (f *RegistryBasedRepositoryFactory) RegolaRepository() repositories.RegolaRepository {
	repo, _ := f.registry.GetRepository("regola")
	return repo.(repositories.RegolaRepository)
}

func (f *RegistryBasedRepositoryFactory) ServizioRepository() repositories.ServizioRepository {
	repo, _ := f.registry.GetRepository("servizio")
	return repo.(repositories.ServizioRepository)
}

func (f *RegistryBasedRepositoryFactory) StrumentoRepository() repositories.StrumentoRepository {
	repo, _ := f.registry.GetRepository("strumento")
	return repo.(repositories.StrumentoRepository)
}

func (f *RegistryBasedRepositoryFactory) TalentoRepository() repositories.TalentoRepository {
	repo, _ := f.registry.GetRepository("talento")
	return repo.(repositories.TalentoRepository)
}

func (f *RegistryBasedRepositoryFactory) CavalcaturaVeicoloRepository() repositories.CavalcaturaVeicoloRepository {
	repo, _ := f.registry.GetRepository("cavalcatura_veicolo")
	return repo.(repositories.CavalcaturaVeicoloRepository)
}

// Content repository - kept as singleton
func (f *RegistryBasedRepositoryFactory) ContentRepository() repositories.ContentRepository {
	return mongodb.NewContentMongoRepository(f.client)
}

// GetRepositoryByEntityType uses the registry instead of switch statements
func (f *RegistryBasedRepositoryFactory) GetRepositoryByEntityType(entityType string) any {
	repo, err := f.registry.GetRepository(entityType)
	if err != nil {
		// Return nil if entity type not found - maintains compatibility
		return nil
	}
	return repo
}