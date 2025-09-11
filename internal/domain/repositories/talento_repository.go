package repositories

import (
	"context"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// TalentoRepository defines operations specific to Talento entities
type TalentoRepository interface {
	BaseRepository[*domain.Talento]

	// FindByNome retrieves a feat by its name
	FindByNome(ctx context.Context, nome string) (*domain.Talento, error)

	// FindByPrerequisite retrieves feats with specific prerequisites
	FindByPrerequisite(ctx context.Context, prerequisite string, limit int) ([]*domain.Talento, error)

	// FindByAbilityScoreIncrease retrieves feats that increase ability scores
	FindByAbilityScoreIncrease(ctx context.Context, limit int) ([]*domain.Talento, error)
}
