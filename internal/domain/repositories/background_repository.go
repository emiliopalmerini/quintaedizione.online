package repositories

import (
	"context"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// BackgroundRepository defines operations specific to Background entities
type BackgroundRepository interface {
	BaseRepository[*domain.Background]

	// FindByNome retrieves a background by its name
	FindByNome(ctx context.Context, nome string) (*domain.Background, error)

	// FindBySkillProficiency retrieves backgrounds that grant specific skill proficiencies
	FindBySkillProficiency(ctx context.Context, skill string, limit int) ([]*domain.Background, error)

	// FindByLanguage retrieves backgrounds that provide specific languages
	FindByLanguage(ctx context.Context, language string, limit int) ([]*domain.Background, error)

	// FindByToolProficiency retrieves backgrounds that grant tool proficiencies
	FindByToolProficiency(ctx context.Context, tool string, limit int) ([]*domain.Background, error)
}
