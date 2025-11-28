package infrastructure

import (
	"sync"
	"time"
)

type SimpleCache struct {
	mu    sync.RWMutex
	items map[string]cacheItem
}

type cacheItem struct {
	value      interface{}
	expiration time.Time
}

func NewSimpleCache() *SimpleCache {
	cache := &SimpleCache{
		items: make(map[string]cacheItem),
	}

	go cache.cleanupExpired()

	return cache
}

func (sc *SimpleCache) Set(key string, value interface{}, ttl time.Duration) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.items[key] = cacheItem{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
}

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

func (sc *SimpleCache) Delete(key string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	delete(sc.items, key)
}

func (sc *SimpleCache) Clear() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.items = make(map[string]cacheItem)
}

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

func (sc *SimpleCache) GetStats() map[string]any {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	return map[string]any{
		"item_count": len(sc.items),
	}
}

var globalCache = NewSimpleCache()

func GetGlobalCache() *SimpleCache {
	return globalCache
}
