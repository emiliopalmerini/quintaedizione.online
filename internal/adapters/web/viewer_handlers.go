package web

import (
	"fmt"

	"github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web/display"
	webmappers "github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web/mappers"
	"github.com/emiliopalmerini/quintaedizione.online/internal/application/services"
	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/search"
	infraconfig "github.com/emiliopalmerini/quintaedizione.online/internal/infrastructure/config"
	"github.com/emiliopalmerini/quintaedizione.online/pkg/templates"
	"github.com/gin-gonic/gin"
)

// Handlers coordinates all specialized handlers.
type Handlers struct {
	home       *HomeHandler
	collection *CollectionHandler
	search     *SearchHandler
	seo        *SEOHandler
}

// NewHandlers creates all specialized handlers with shared dependencies.
func NewHandlers(contentService *services.ContentService, searchService search.SearchService, templateEngine *templates.TemplEngine) *Handlers {
	displayFactory := display.NewDisplayElementFactory()
	documentMapper := webmappers.NewDocumentMapper(displayFactory)

	collectionMetadata, err := infraconfig.NewCollectionMetadata()
	if err != nil {
		fmt.Printf("Warning: Failed to load collection metadata: %v\n", err)
	}

	base := &baseHandler{
		contentService:     contentService,
		templateEngine:     templateEngine,
		documentMapper:     documentMapper,
		collectionMetadata: collectionMetadata,
	}

	return &Handlers{
		home: &HomeHandler{
			baseHandler: base,
		},
		collection: &CollectionHandler{
			baseHandler: base,
		},
		search: &SearchHandler{
			baseHandler:   base,
			searchService: searchService,
		},
		seo: &SEOHandler{
			baseHandler: base,
		},
	}
}

// RegisterRoutes registers all HTTP routes with their respective handlers.
func (h *Handlers) RegisterRoutes(router *gin.Engine) {
	router.GET("/", h.home.handleHome)

	// Specific routes must be registered before wildcard routes
	router.GET("/search", h.search.handleGlobalSearch)
	router.GET("/search/dropdown", h.search.handleSearchDropdown)
	router.GET("/robots.txt", h.seo.handleRobotsTxt)
	router.GET("/sitemap.xml", h.seo.handleSitemap)

	// Wildcard routes (must come last)
	router.GET("/:collection", h.collection.handleCollectionList)
	router.GET("/:collection/rows", h.collection.handleCollectionRows)
	router.GET("/:collection/:slug", h.collection.handleItemDetail)
}

// baseHandlerForMiddleware returns a baseHandler for middleware use.
func (h *Handlers) baseHandlerForMiddleware() *baseHandler {
	return h.home.baseHandler
}
