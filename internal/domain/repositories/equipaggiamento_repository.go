package repositories

import (
	"context"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// EquipaggiamentoRepository defines operations specific to Equipaggiamento entities
type EquipaggiamentoRepository interface {
	BaseRepository[*domain.Equipaggiamento]

	// FindByNome retrieves equipment by its name
	FindByNome(ctx context.Context, nome string) (*domain.Equipaggiamento, error)

	// FindByCategory retrieves equipment by category
	FindByCategory(ctx context.Context, category string, limit int) ([]*domain.Equipaggiamento, error)

	// FindByPriceRange retrieves equipment within price range
	FindByPriceRange(ctx context.Context, minPrice, maxPrice float64, limit int) ([]*domain.Equipaggiamento, error)

	// FindByWeight retrieves equipment by weight
	FindByWeight(ctx context.Context, maxWeight float64, limit int) ([]*domain.Equipaggiamento, error)
}
