package domain

import "time"

// Cache defines an interface for caching key-value pairs with TTL.
type Cache interface {
	Get(key string) (any, bool)
	Set(key string, value any, ttl time.Duration)
	GetStats() map[string]any
}
