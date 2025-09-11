package repositories

import (
	"context"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// SpecieRepository defines operations specific to Specie entities
type SpecieRepository interface {
	BaseRepository[*domain.Specie]

	// FindByNome retrieves a species by its name
	FindByNome(ctx context.Context, nome string) (*domain.Specie, error)

	// FindBySize retrieves species by size
	FindBySize(ctx context.Context, size string, limit int) ([]*domain.Specie, error)

	// FindByAbilityScoreIncrease retrieves species by ability score bonuses
	FindByAbilityScoreIncrease(ctx context.Context, ability string, limit int) ([]*domain.Specie, error)

	// FindBySpeed retrieves species by movement speed
	FindBySpeed(ctx context.Context, minSpeed int, limit int) ([]*domain.Specie, error)
}
