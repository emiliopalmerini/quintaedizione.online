package config

// CacheType defines cache policy types for different response types
type CacheType string

const (
	// CacheTypeHome is for home page responses with collection stats
	CacheTypeHome CacheType = "home"
	// CacheTypeCollection is for collection lists and rows
	CacheTypeCollection CacheType = "collection"
	// CacheTypeItem is for individual item details
	CacheTypeItem CacheType = "item"
	// CacheTypeSearch is for search results (no caching)
	CacheTypeSearch CacheType = "search"
)

// CacheDurations defines cache max-age values in seconds for different response types
var CacheDurations = map[CacheType]int{
	CacheTypeHome:       3600,  // 1 hour - home page with collection stats
	CacheTypeCollection: 1800,  // 30 minutes - collection lists and rows
	CacheTypeItem:       14400, // 4 hours - individual item details (considering D&D session prep time)
	CacheTypeSearch:     0,     // No cache for search results
}

// GetCacheDuration returns the cache duration for a cache type
func GetCacheDuration(cacheType CacheType) int {
	if duration, exists := CacheDurations[cacheType]; exists {
		return duration
	}
	return 1800 // Default to 30 minutes
}
