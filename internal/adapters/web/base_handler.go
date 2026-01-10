package web

import (
	"fmt"

	"github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web/config"
	webmappers "github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web/mappers"
	"github.com/emiliopalmerini/quintaedizione.online/internal/application/services"
	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/collections"
	infraconfig "github.com/emiliopalmerini/quintaedizione.online/internal/infrastructure/config"
	"github.com/emiliopalmerini/quintaedizione.online/pkg/templates"
	"github.com/gin-gonic/gin"
)

// baseHandler contains shared dependencies for all handlers.
type baseHandler struct {
	contentService     *services.ContentService
	templateEngine     *templates.TemplEngine
	documentMapper     webmappers.DocumentMapper
	collectionMetadata infraconfig.CollectionMetadata
}

// getCollectionTitle returns the display title for a collection.
func (h *baseHandler) getCollectionTitle(collection string) string {
	if h.collectionMetadata != nil {
		return h.collectionMetadata.GetTitle(collection)
	}
	return collections.GetTitle(collection)
}

// cacheTypeMap maps string identifiers to CacheType constants.
var cacheTypeMap = map[string]config.CacheType{
	"home":       config.CacheTypeHome,
	"collection": config.CacheTypeCollection,
	"item":       config.CacheTypeItem,
	"search":     config.CacheTypeSearch,
}

// setCacheHeaders sets appropriate cache control headers based on the cache type.
func (h *baseHandler) setCacheHeaders(c *gin.Context, cacheTypeStr string) {
	cacheType, exists := cacheTypeMap[cacheTypeStr]
	if !exists {
		cacheType = config.CacheTypeCollection
	}

	maxAge := config.GetCacheDuration(cacheType)

	if maxAge > 0 {
		c.Header("Cache-Control", fmt.Sprintf("max-age=%d, public", maxAge))
	} else {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
	}
}
