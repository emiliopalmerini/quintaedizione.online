package repositories

import (
	"context"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// MostroRepository defines operations specific to Mostro entities
type MostroRepository interface {
	BaseRepository[*domain.Mostro]

	// FindByNome retrieves a monster by its name
	FindByNome(ctx context.Context, nome string) (*domain.Mostro, error)

	// FindByChallengeRating retrieves monsters by challenge rating
	FindByChallengeRating(ctx context.Context, cr float64, limit int) ([]*domain.Mostro, error)

	// FindByChallengeRatingRange retrieves monsters within CR range
	FindByChallengeRatingRange(ctx context.Context, minCR, maxCR float64, limit int) ([]*domain.Mostro, error)

	// FindByType retrieves monsters by type (Aberrazione, Bestia, etc.)
	FindByType(ctx context.Context, tipoMostro domain.TipoMostro, limit int) ([]*domain.Mostro, error)

	// FindBySize retrieves monsters by size
	FindBySize(ctx context.Context, size string, limit int) ([]*domain.Mostro, error)

	// FindByEnvironment retrieves monsters by habitat/environment
	FindByEnvironment(ctx context.Context, environment string, limit int) ([]*domain.Mostro, error)

	// FindByAlignment retrieves monsters by alignment
	FindByAlignment(ctx context.Context, alignment string, limit int) ([]*domain.Mostro, error)

	// FindSpellcasters retrieves monsters that can cast spells
	FindSpellcasters(ctx context.Context, limit int) ([]*domain.Mostro, error)

	// FindLegendaryMonsters retrieves monsters with legendary actions
	FindLegendaryMonsters(ctx context.Context, limit int) ([]*domain.Mostro, error)
}
