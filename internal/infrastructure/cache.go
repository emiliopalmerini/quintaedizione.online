package infrastructure

import (
	"sync"
	"time"
)

// SimpleCache provides basic in-memory caching with TTL
type SimpleCache struct {
	mu    sync.RWMutex
	items map[string]cacheItem
}

type cacheItem struct {
	value      interface{}
	expiration time.Time
}

// NewSimpleCache creates a new simple cache
func NewSimpleCache() *SimpleCache {
	cache := &SimpleCache{
		items: make(map[string]cacheItem),
	}

	// Start cleanup goroutine
	go cache.cleanupExpired()

	return cache
}

// Set adds an item to the cache with TTL
func (sc *SimpleCache) Set(key string, value interface{}, ttl time.Duration) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.items[key] = cacheItem{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
}

// Get retrieves an item from the cache
func (sc *SimpleCache) Get(key string) (interface{}, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	item, exists := sc.items[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(item.expiration) {
		return nil, false
	}

	return item.value, true
}

// Delete removes an item from the cache
func (sc *SimpleCache) Delete(key string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	delete(sc.items, key)
}

// Clear removes all items from the cache
func (sc *SimpleCache) Clear() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.items = make(map[string]cacheItem)
}

// cleanupExpired removes expired items from cache
func (sc *SimpleCache) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sc.mu.Lock()
		now := time.Now()
		for key, item := range sc.items {
			if now.After(item.expiration) {
				delete(sc.items, key)
			}
		}
		sc.mu.Unlock()
	}
}

// GetStats returns cache statistics
func (sc *SimpleCache) GetStats() map[string]any {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	return map[string]any{
		"item_count": len(sc.items),
	}
}

// Global cache instance
var globalCache = NewSimpleCache()

// GetGlobalCache returns the global cache instance
func GetGlobalCache() *SimpleCache {
	return globalCache
}
