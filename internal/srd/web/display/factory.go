package display

import "github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/dto"

type DisplayElementFactory struct {
	strategies  map[string]DisplayElementStrategy
	multiSource bool // true when multiple sources are loaded
}

func NewDisplayElementFactory(multiSource bool) *DisplayElementFactory {
	factory := &DisplayElementFactory{
		strategies:  make(map[string]DisplayElementStrategy),
		multiSource: multiSource,
	}

	strategies := []DisplayElementStrategy{
		&IncantesimiDisplayStrategy{},
		&OggettiMagiciDisplayStrategy{},
		&MostriDisplayStrategy{},
		&EquipaggiamentiDisplayStrategy{},
		&BackgroundsDisplayStrategy{},
		&TalentiDisplayStrategy{},
		&ClassiDisplayStrategy{},
		&GlossarioDisplayStrategy{},
		&SpecieDisplayStrategy{},
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
	elements := strategy.GetElements(doc)

	// Append edition badge if source metadata is present and multiple sources loaded
	if f.multiSource {
		if source, ok := doc["_source_short"].(string); ok && source != "" {
			elements = append(elements, dto.DisplayElementDTO{
				Value: source,
				Type:  "edition",
			})
		}
	}

	return elements
}
