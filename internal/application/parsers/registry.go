package parsers

import (
	"errors"
	"fmt"
	"sync"
)

func CreateDocumentRegistry() (*DocumentRegistry, error) {
	registry := NewDocumentRegistry()

	strategies := []DocumentParsingStrategy{
		NewArmatureDocumentStrategy(),
		NewArmiDocumentStrategy(),
		NewAnimaliDocumentStrategy(),
		NewMostriDocumentStrategy(),
		NewClassiDocumentStrategy(),
		NewBackgroundsDocumentStrategy(),
		NewIncantesimiDocumentStrategy(),
		NewTalentiDocumentStrategy(),
		NewEquipaggiamentiDocumentStrategy(),
		NewServiziDocumentStrategy(),
		NewStrumentiDocumentStrategy(),
		NewRegoleDocumentStrategy(),
		NewCavalcatureVeicoliDocumentStrategy(),
		NewOggettiMagiciDocumentStrategy(),
	}

	for _, strategy := range strategies {
		key := fmt.Sprintf("%s_%s", strategy.ContentType(), Italian)
		if err := registry.Register(key, strategy); err != nil {
			return nil, fmt.Errorf("failed to register %s: %w", strategy.Name(), err)
		}
	}

	return registry, nil
}

type DocumentRegistry struct {
	strategies map[string]DocumentParsingStrategy
	mu         sync.RWMutex
}

func NewDocumentRegistry() *DocumentRegistry {
	return &DocumentRegistry{
		strategies: make(map[string]DocumentParsingStrategy),
	}
}

func (r *DocumentRegistry) Register(key string, strategy DocumentParsingStrategy) error {
	if strategy == nil {
		return errors.New("strategy cannot be nil")
	}

	if key == "" {
		return errors.New("key cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.strategies[key]; exists {
		return fmt.Errorf("parser already exists for key: %s", key)
	}

	r.strategies[key] = strategy
	return nil
}

func (r *DocumentRegistry) GetStrategyByKey(key string) (DocumentParsingStrategy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if strategy, exists := r.strategies[key]; exists {
		return strategy, nil
	}

	return nil, fmt.Errorf("parser not found for key: %s", key)
}

func (r *DocumentRegistry) GetStrategy(contentType ContentType, language LanguageCode) (DocumentParsingStrategy, error) {
	key := fmt.Sprintf("%s_%s", contentType, language)
	return r.GetStrategyByKey(key)
}

func (r *DocumentRegistry) ListKeys() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	keys := make([]string, 0, len(r.strategies))
	for key := range r.strategies {
		keys = append(keys, key)
	}

	return keys
}

func (r *DocumentRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.strategies)
}
