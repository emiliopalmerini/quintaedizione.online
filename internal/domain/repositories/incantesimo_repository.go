package repositories

import (
	"context"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// IncantesimoRepository defines operations specific to Incantesimo entities
type IncantesimoRepository interface {
	BaseRepository[*domain.Incantesimo]

	// FindByNome retrieves a spell by its name
	FindByNome(ctx context.Context, nome string) (*domain.Incantesimo, error)

	// FindByLevel retrieves spells by level
	FindByLevel(ctx context.Context, level int, limit int) ([]*domain.Incantesimo, error)

	// FindBySchool retrieves spells by school of magic
	FindBySchool(ctx context.Context, school string, limit int) ([]*domain.Incantesimo, error)

	// FindByClass retrieves spells available to a specific class
	FindByClass(ctx context.Context, className string, limit int) ([]*domain.Incantesimo, error)

	// FindByLevelAndClass retrieves spells by level and class
	FindByLevelAndClass(ctx context.Context, level int, className string, limit int) ([]*domain.Incantesimo, error)

	// FindByComponents retrieves spells by required components
	FindByComponents(ctx context.Context, components []string, limit int) ([]*domain.Incantesimo, error)

	// FindRitualSpells retrieves ritual spells
	FindRitualSpells(ctx context.Context, limit int) ([]*domain.Incantesimo, error)
}
