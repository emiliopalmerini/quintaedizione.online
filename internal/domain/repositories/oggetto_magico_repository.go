package repositories

import (
	"context"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// OggettoMagicoRepository defines operations specific to OggettoMagico entities
type OggettoMagicoRepository interface {
	BaseRepository[*domain.OggettoMagico]

	// FindByNome retrieves a magic item by its name
	FindByNome(ctx context.Context, nome string) (*domain.OggettoMagico, error)

	// FindByRarity retrieves magic items by rarity
	FindByRarity(ctx context.Context, rarity string, limit int) ([]*domain.OggettoMagico, error)

	// FindByType retrieves magic items by type
	FindByType(ctx context.Context, itemType string, limit int) ([]*domain.OggettoMagico, error)

	// FindByAttunement retrieves items that require attunement
	FindByAttunement(ctx context.Context, requiresAttunement bool, limit int) ([]*domain.OggettoMagico, error)

	// FindConsumableItems retrieves consumable magic items
	FindConsumableItems(ctx context.Context, limit int) ([]*domain.OggettoMagico, error)
}
