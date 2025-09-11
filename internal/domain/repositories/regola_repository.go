package repositories

import (
	"context"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// RegolaRepository defines operations specific to Regola entities
type RegolaRepository interface {
	BaseRepository[*domain.Regola]

	// FindByNome retrieves a rule by its name
	FindByNome(ctx context.Context, nome string) (*domain.Regola, error)

	// FindByCategory retrieves rules by category
	FindByCategory(ctx context.Context, category string, limit int) ([]*domain.Regola, error)

	// FindByKeyword searches rules by keyword in content
	FindByKeyword(ctx context.Context, keyword string, limit int) ([]*domain.Regola, error)
}
