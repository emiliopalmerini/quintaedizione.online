package parsers

import (
	"errors"
	"fmt"
	"sync"
)


// CreateDefaultRegistry creates and configures the default parser registry
func CreateDefaultRegistry() (*Registry, error) {
	registry := NewRegistry()

	strategies := []ParsingStrategy{
		NewArmatureStrategy(),
		NewArmiStrategy(),
		NewAnimaliStrategy(),
		NewMostriStrategy(),
		NewClassiStrategy(),
		NewBackgroundsStrategy(),
		NewIncantesimiStrategy(),
		NewTalentiStrategy(),
		NewEquipaggiamentiStrategy(),
		NewServiziStrategy(),
		NewStrumentiStrategy(),
		NewRegoleStrategy(),
		NewCavalcatureVeicoliStrategy(),
		NewOggettiMagiciStrategy(),
	}

	for _, strategy := range strategies {
		key := fmt.Sprintf("%s_%s", strategy.ContentType(), Italian)
		if err := registry.Register(key, strategy); err != nil {
			return nil, fmt.Errorf("failed to register %s: %w", strategy.Name(), err)
		}
	}

	return registry, nil
}

// GetAvailableParsers returns a list of all available parser information
func GetAvailableParsers(registry *Registry) []ParserInfo {
	keys := registry.ListKeys()
	var parsers []ParserInfo

	for _, key := range keys {
		if strategy, err := registry.GetStrategyByKey(key); err == nil {
			parsers = append(parsers, ParserInfo{
				Key:         key,
				ContentType: strategy.ContentType(),
				Name:        strategy.Name(),
				Description: strategy.Description(),
			})
		}
	}

	return parsers
}

// ParserInfo holds information about a registered parser
type ParserInfo struct {
	Key         string      `json:"key"`
	ContentType ContentType `json:"content_type"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
}

// Registry manages multi-language parsing strategies using the Registry pattern
type Registry struct {
	strategies map[string]ParsingStrategy // key: "contentType_language" or just "contentType"
	mu         sync.RWMutex
}

// NewRegistry creates a new multi-language parser registry
func NewRegistry() *Registry {
	return &Registry{
		strategies: make(map[string]ParsingStrategy),
	}
}

// Register adds a parsing strategy with a specific key
func (r *Registry) Register(key string, strategy ParsingStrategy) error {
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

// GetStrategy retrieves a parsing strategy by key
func (r *Registry) GetStrategyByKey(key string) (ParsingStrategy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if strategy, exists := r.strategies[key]; exists {
		return strategy, nil
	}

	return nil, fmt.Errorf("parser not found for key: %s", key)
}

// GetStrategy retrieves a parsing strategy for specific content type and language
func (r *Registry) GetStrategy(contentType ContentType, language LanguageCode) (ParsingStrategy, error) {
	key := fmt.Sprintf("%s_%s", contentType, language)
	return r.GetStrategyByKey(key)
}

// GetStrategyByContentType retrieves a parsing strategy by content type only (backward compatibility)
// Defaults to Italian for backward compatibility
func (r *Registry) GetStrategyByContentType(contentType ContentType) (ParsingStrategy, error) {
	// First try Italian (default)
	if strategy, err := r.GetStrategy(contentType, Italian); err == nil {
		return strategy, nil
	}

	// Fallback to old-style key (just contentType)
	key := string(contentType)
	return r.GetStrategyByKey(key)
}

// HasStrategy checks if a strategy exists for the given key
func (r *Registry) HasStrategy(key string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.strategies[key]
	return exists
}

// ListKeys returns all registered strategy keys
func (r *Registry) ListKeys() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	keys := make([]string, 0, len(r.strategies))
	for key := range r.strategies {
		keys = append(keys, key)
	}

	return keys
}

// Count returns the number of registered strategies
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.strategies)
}

// Unregister removes a strategy from the registry
func (r *Registry) Unregister(key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.strategies[key]; !exists {
		return fmt.Errorf("strategy not found for key: %s", key)
	}

	delete(r.strategies, key)
	return nil
}

// Clear removes all strategies from the registry
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.strategies = make(map[string]ParsingStrategy)
}
