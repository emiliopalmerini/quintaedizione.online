package repositories

import (
	"fmt"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/adapters/repositories/mongodb"
	pkgMongodb "github.com/emiliopalmerini/due-draghi-5e-srd/pkg/mongodb"
)

// RepositoryFactoryFunc creates repository instances
type RepositoryFactoryFunc func(client *pkgMongodb.Client) interface{}

// RepositoryRegistry manages repository factories and creation
type RepositoryRegistry interface {
	Register(entityType string, factory RepositoryFactoryFunc)
	GetRepository(entityType string) (interface{}, error)
	GetAvailableTypes() []string
}

type repositoryRegistry struct {
	factories map[string]RepositoryFactoryFunc
	client    *pkgMongodb.Client
}

// NewRepositoryRegistry creates a new repository registry with default factories
func NewRepositoryRegistry(client *pkgMongodb.Client) RepositoryRegistry {
	registry := &repositoryRegistry{
		factories: make(map[string]RepositoryFactoryFunc),
		client:    client,
	}

	// Register all default factories
	registry.registerDefaults()
	return registry
}

// Register adds a new repository factory for an entity type
func (r *repositoryRegistry) Register(entityType string, factory RepositoryFactoryFunc) {
	r.factories[entityType] = factory
}

// GetRepository creates and returns a repository for the given entity type
func (r *repositoryRegistry) GetRepository(entityType string) (interface{}, error) {
	factory, exists := r.factories[entityType]
	if !exists {
		return nil, fmt.Errorf("unknown entity type: %s", entityType)
	}

	return factory(r.client), nil
}

// GetAvailableTypes returns all registered entity types
func (r *repositoryRegistry) GetAvailableTypes() []string {
	types := make([]string, 0, len(r.factories))
	for entityType := range r.factories {
		types = append(types, entityType)
	}
	return types
}

// registerDefaults registers all default repository factories
func (r *repositoryRegistry) registerDefaults() {
	r.Register("arma", func(client *pkgMongodb.Client) interface{} {
		return mongodb.NewArmaMongoRepository(client)
	})
	r.Register("armatura", func(client *pkgMongodb.Client) interface{} {
		return mongodb.NewArmaturaMongoRepository(client)
	})
	r.Register("animale", func(client *pkgMongodb.Client) interface{} {
		return mongodb.NewAnimaleMongoRepository(client)
	})
	r.Register("background", func(client *pkgMongodb.Client) interface{} {
		return mongodb.NewBackgroundMongoRepository(client)
	})
	r.Register("classe", func(client *pkgMongodb.Client) interface{} {
		return mongodb.NewClasseMongoRepository(client)
	})
	r.Register("equipaggiamento", func(client *pkgMongodb.Client) interface{} {
		return mongodb.NewEquipaggiamentoMongoRepository(client)
	})
	r.Register("incantesimo", func(client *pkgMongodb.Client) interface{} {
		return mongodb.NewIncantesimoMongoRepository(client)
	})
	r.Register("mostro", func(client *pkgMongodb.Client) interface{} {
		return mongodb.NewMostroMongoRepository(client)
	})
	r.Register("oggetto_magico", func(client *pkgMongodb.Client) interface{} {
		return mongodb.NewOggettoMagicoMongoRepository(client)
	})
	r.Register("regola", func(client *pkgMongodb.Client) interface{} {
		return mongodb.NewRegolaMongoRepository(client)
	})
	r.Register("servizio", func(client *pkgMongodb.Client) interface{} {
		return mongodb.NewServizioMongoRepository(client)
	})
	r.Register("strumento", func(client *pkgMongodb.Client) interface{} {
		return mongodb.NewStrumentoMongoRepository(client)
	})
	r.Register("talento", func(client *pkgMongodb.Client) interface{} {
		return mongodb.NewTalentoMongoRepository(client)
	})
	r.Register("cavalcatura_veicolo", func(client *pkgMongodb.Client) interface{} {
		return mongodb.NewCavalcaturaVeicoloMongoRepository(client)
	})
}