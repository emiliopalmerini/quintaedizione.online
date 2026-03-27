package web

import (
	"net/http"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/application/services"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/collections"
	infraconfig "github.com/emiliopalmerini/quintaedizione.online/internal/srd/infrastructure/config"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/config"
	webmappers "github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/mappers"
	"github.com/emiliopalmerini/quintaedizione.online/pkg/templates"
	pkgweb "github.com/emiliopalmerini/quintaedizione.online/pkg/web"
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

// renderHTML sets cache headers, content type, and writes the HTML response.
func (h *baseHandler) renderHTML(w http.ResponseWriter, content string, cacheType string) {
	h.setCacheHeaders(w, cacheType)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(content))
}

// setCacheHeaders sets appropriate cache control headers based on the cache type.
func (h *baseHandler) setCacheHeaders(w http.ResponseWriter, cacheTypeStr string) {
	cacheType, exists := cacheTypeMap[cacheTypeStr]
	if !exists {
		cacheType = config.CacheTypeCollection
	}

	pkgweb.SetCacheHeaders(w, config.GetCacheDuration(cacheType))
}
