package display

import (
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/dto"
)

// VersionResolver returns the available source short names for a document slug.
type VersionResolver func(collection, slug string) []string

type DisplayElementFactory struct {
	strategies      map[string]DisplayElementStrategy
	multiSource     bool // true when multiple sources are loaded
	versionResolver VersionResolver
}

func NewDisplayElementFactory(multiSource bool, opts ...func(*DisplayElementFactory)) *DisplayElementFactory {
	factory := &DisplayElementFactory{
		strategies:  make(map[string]DisplayElementStrategy),
		multiSource: multiSource,
	}
	for _, opt := range opts {
		opt(factory)
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

// WithVersionResolver sets a function that resolves all available sources for a slug.
func WithVersionResolver(fn VersionResolver) func(*DisplayElementFactory) {
	return func(f *DisplayElementFactory) { f.versionResolver = fn }
}

func (f *DisplayElementFactory) GetStrategy(collection string) DisplayElementStrategy {
	if strategy, exists := f.strategies[collection]; exists {
		return strategy
	}
	return &DefaultDisplayStrategy{}
}

func (f *DisplayElementFactory) GetDisplayElements(collection string, doc *domain.Document) []dto.DisplayElementDTO {
	strategy := f.GetStrategy(collection)
	elements := strategy.GetElements(doc)

	// Append edition badges for all available sources when multiple sources loaded
	if f.multiSource {
		if f.versionResolver != nil {
			for _, src := range f.versionResolver(collection, string(doc.ID)) {
				elements = append(elements, dto.DisplayElementDTO{Value: src, Type: "edition"})
			}
		} else if doc.Source != "" {
			elements = append(elements, dto.DisplayElementDTO{Value: doc.Source, Type: "edition"})
		}
	}

	return elements
}
