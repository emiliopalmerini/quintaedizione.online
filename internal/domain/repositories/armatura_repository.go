package repositories

import (
	"context"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// ArmaturaRepository defines operations specific to Armatura entities
type ArmaturaRepository interface {
	BaseRepository[*domain.Armatura]

	// FindByNome retrieves armor by its name
	FindByNome(ctx context.Context, nome string) (*domain.Armatura, error)

	// FindByType retrieves armor by type (Leggera, Media, Pesante, Scudo)
	FindByType(ctx context.Context, armorType string, limit int) ([]*domain.Armatura, error)

	// FindByACRange retrieves armor within AC range
	FindByACRange(ctx context.Context, minAC, maxAC int, limit int) ([]*domain.Armatura, error)

	// FindStealthDisadvantage retrieves armor that imposes stealth disadvantage
	FindStealthDisadvantage(ctx context.Context, limit int) ([]*domain.Armatura, error)
}
