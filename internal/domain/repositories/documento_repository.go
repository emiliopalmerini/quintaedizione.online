package repositories

import (
	"context"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// DocumentoRepository defines operations specific to Documento entities
type DocumentoRepository interface {
	BaseRepository[*domain.Documento]

	// FindBySlug retrieves a document by its slug
	FindBySlug(ctx context.Context, slug string) (*domain.Documento, error)

	// FindByTitle searches documents by title (partial match)
	FindByTitle(ctx context.Context, title string, limit int) ([]*domain.Documento, error)

	// FindByContent searches documents by content (text search)
	FindByContent(ctx context.Context, searchText string, limit int) ([]*domain.Documento, error)
}
