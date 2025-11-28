package display

import "github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web/dto"

// DisplayElementFactory creates the appropriate display strategy for a collection
type DisplayElementFactory struct {
	strategies map[string]DisplayElementStrategy
}

// NewDisplayElementFactory creates a new factory with all strategies registered
func NewDisplayElementFactory() *DisplayElementFactory {
	factory := &DisplayElementFactory{
		strategies: make(map[string]DisplayElementStrategy),
	}

	// Register all strategies
	strategies := []DisplayElementStrategy{
		&IncantesimiDisplayStrategy{},
		&OggettiMagiciDisplayStrategy{},
		&MostriDisplayStrategy{},
		&ArmiDisplayStrategy{},
		&ArmatureDisplayStrategy{},
	}

	for _, strategy := range strategies {
		factory.strategies[strategy.GetCollectionType()] = strategy
	}

	return factory
}

// GetStrategy returns the appropriate strategy for a collection
func (f *DisplayElementFactory) GetStrategy(collection string) DisplayElementStrategy {
	if strategy, exists := f.strategies[collection]; exists {
		return strategy
	}
	return &DefaultDisplayStrategy{}
}

// GetDisplayElements extracts display elements for a document using the appropriate strategy
func (f *DisplayElementFactory) GetDisplayElements(collection string, doc map[string]any) []dto.DisplayElementDTO {
	strategy := f.GetStrategy(collection)
	return strategy.GetElements(doc)
}
