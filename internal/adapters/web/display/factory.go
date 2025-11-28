package display

import "github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web/dto"

type DisplayElementFactory struct {
	strategies map[string]DisplayElementStrategy
}

func NewDisplayElementFactory() *DisplayElementFactory {
	factory := &DisplayElementFactory{
		strategies: make(map[string]DisplayElementStrategy),
	}

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

func (f *DisplayElementFactory) GetStrategy(collection string) DisplayElementStrategy {
	if strategy, exists := f.strategies[collection]; exists {
		return strategy
	}
	return &DefaultDisplayStrategy{}
}

func (f *DisplayElementFactory) GetDisplayElements(collection string, doc map[string]any) []dto.DisplayElementDTO {
	strategy := f.GetStrategy(collection)
	return strategy.GetElements(doc)
}
