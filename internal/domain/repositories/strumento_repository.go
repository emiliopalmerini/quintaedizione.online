package repositories

import (
	"context"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// StrumentoRepository defines operations specific to Strumento entities
type StrumentoRepository interface {
	BaseRepository[*domain.Strumento]

	// FindByNome retrieves a tool by its name
	FindByNome(ctx context.Context, nome string) (*domain.Strumento, error)

	// FindByCategory retrieves tools by category
	FindByCategory(ctx context.Context, category string, limit int) ([]*domain.Strumento, error)

	// FindByUse retrieves tools by their usage type
	FindByUse(ctx context.Context, useType string, limit int) ([]*domain.Strumento, error)
}
