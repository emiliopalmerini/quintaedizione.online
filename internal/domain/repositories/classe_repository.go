package repositories

import (
	"context"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/domain"
)

// ClasseRepository defines operations specific to Classe entities
type ClasseRepository interface {
	BaseRepository[*domain.Classe]

	// FindByNome retrieves a class by its name
	FindByNome(ctx context.Context, nome string) (*domain.Classe, error)

	// FindSpellcasterClasses retrieves classes that can cast spells
	FindSpellcasterClasses(ctx context.Context, limit int) ([]*domain.Classe, error)

	// FindByHitDie retrieves classes by hit die size
	FindByHitDie(ctx context.Context, hitDie int, limit int) ([]*domain.Classe, error)

	// FindByPrimaryAbility retrieves classes by primary ability score
	FindByPrimaryAbility(ctx context.Context, ability string, limit int) ([]*domain.Classe, error)

	// FindBySavingThrowProficiency retrieves classes by saving throw proficiencies
	FindBySavingThrowProficiency(ctx context.Context, savingThrow string, limit int) ([]*domain.Classe, error)

	// FindMulticlassEligible retrieves classes with multiclass prerequisites
	FindMulticlassEligible(ctx context.Context, limit int) ([]*domain.Classe, error)
}
