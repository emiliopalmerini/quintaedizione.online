package config

type CacheType string

const (

	CacheTypeHome CacheType = "home"

	CacheTypeCollection CacheType = "collection"

	CacheTypeItem CacheType = "item"

	CacheTypeSearch CacheType = "search"
)

var CacheDurations = map[CacheType]int{
	CacheTypeHome:       3600,
	CacheTypeCollection: 1800,
	CacheTypeItem:       14400,
	CacheTypeSearch:     0,
}

func GetCacheDuration(cacheType CacheType) int {
	if duration, exists := CacheDurations[cacheType]; exists {
		return duration
	}
	return 1800
}
