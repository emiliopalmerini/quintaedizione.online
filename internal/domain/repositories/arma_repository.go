package repositories

import (
	"context"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// ArmaRepository defines operations specific to Arma entities
type ArmaRepository interface {
	BaseRepository[*domain.Arma]

	// FindByNome retrieves a weapon by its name
	FindByNome(ctx context.Context, nome string) (*domain.Arma, error)

	// FindByCategory retrieves weapons by category (Semplici, Militari, etc.)
	FindByCategory(ctx context.Context, category string, limit int) ([]*domain.Arma, error)

	// FindByDamageType retrieves weapons by damage type
	FindByDamageType(ctx context.Context, damageType string, limit int) ([]*domain.Arma, error)

	// FindRangedWeapons retrieves ranged weapons
	FindRangedWeapons(ctx context.Context, limit int) ([]*domain.Arma, error)

	// FindMeleeWeapons retrieves melee weapons
	FindMeleeWeapons(ctx context.Context, limit int) ([]*domain.Arma, error)

	// FindByProperty retrieves weapons with specific properties (Finezza, Pesante, etc.)
	FindByProperty(ctx context.Context, property string, limit int) ([]*domain.Arma, error)
}
