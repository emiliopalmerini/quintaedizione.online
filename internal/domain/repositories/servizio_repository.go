package repositories

import (
	"context"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// ServizioRepository defines operations specific to Servizio entities
type ServizioRepository interface {
	BaseRepository[*domain.Servizio]

	// FindByNome retrieves a service by its name
	FindByNome(ctx context.Context, nome string) (*domain.Servizio, error)

	// FindByCategory retrieves services by category
	FindByCategory(ctx context.Context, category string, limit int) ([]*domain.Servizio, error)

	// FindByPriceRange retrieves services within price range
	FindByPriceRange(ctx context.Context, minPrice, maxPrice float64, limit int) ([]*domain.Servizio, error)
}
