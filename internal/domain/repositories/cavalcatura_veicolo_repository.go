package repositories

import (
	"context"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// CavalcaturaVeicoloRepository defines operations specific to CavalcaturaVeicolo entities
type CavalcaturaVeicoloRepository interface {
	BaseRepository[*domain.CavalcaturaVeicolo]

	// FindByNome retrieves a mount/vehicle by its name
	FindByNome(ctx context.Context, nome string) (*domain.CavalcaturaVeicolo, error)

	// FindByType retrieves mounts/vehicles by type
	FindByType(ctx context.Context, vehicleType string, limit int) ([]*domain.CavalcaturaVeicolo, error)

	// FindBySpeed retrieves mounts/vehicles by speed
	FindBySpeed(ctx context.Context, minSpeed int, limit int) ([]*domain.CavalcaturaVeicolo, error)

	// FindByCapacity retrieves vehicles by carrying capacity
	FindByCapacity(ctx context.Context, minCapacity int, limit int) ([]*domain.CavalcaturaVeicolo, error)
}
