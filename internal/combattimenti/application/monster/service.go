package monster

import (
	"github.com/emiliopalmerini/quintaedizione.online/internal/combattimenti/domain/monster"
)

// Service provides monster search use cases.
type Service struct {
	repo monster.Repository
}

// NewService creates a new monster application service.
func NewService(repo monster.Repository) *Service {
	return &Service{repo: repo}
}

// SearchMonsters returns monsters matching the query with XP up to maxXP.
func (s *Service) SearchMonsters(query string, maxXP int) []monster.Monster {
	return s.repo.Search(query, maxXP)
}

// SearchMonstersWithFilters returns monsters matching the given filters.
func (s *Service) SearchMonstersWithFilters(filters monster.SearchFilters) []monster.Monster {
	return s.repo.SearchWithFilters(filters)
}

// AvailableTypes returns all distinct monster types.
func (s *Service) AvailableTypes() []string {
	return s.repo.AvailableTypes()
}

// AvailableSizes returns all distinct normalized monster sizes.
func (s *Service) AvailableSizes() []string {
	return s.repo.AvailableSizes()
}

// AvailableCRs returns all distinct CRs sorted numerically.
func (s *Service) AvailableCRs() []string {
	return s.repo.AvailableCRs()
}
