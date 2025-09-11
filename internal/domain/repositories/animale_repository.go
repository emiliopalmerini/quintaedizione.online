package repositories

import (
	"context"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// AnimaleRepository defines operations specific to Animale entities
type AnimaleRepository interface {
	BaseRepository[*domain.Animale]

	// FindByNome retrieves an animal by its name
	FindByNome(ctx context.Context, nome string) (*domain.Animale, error)

	// FindBySize retrieves animals by size
	FindBySize(ctx context.Context, size string, limit int) ([]*domain.Animale, error)

	// FindByChallengeRating retrieves animals by challenge rating
	FindByChallengeRating(ctx context.Context, cr float64, limit int) ([]*domain.Animale, error)

	// FindByEnvironment retrieves animals by habitat
	FindByEnvironment(ctx context.Context, environment string, limit int) ([]*domain.Animale, error)
}
