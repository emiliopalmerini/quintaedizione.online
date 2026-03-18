package web

import (
	"fmt"
	"net/http"

	"github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web/display"
	webmappers "github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web/mappers"
	"github.com/emiliopalmerini/quintaedizione.online/internal/application/services"
	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/search"
	infraconfig "github.com/emiliopalmerini/quintaedizione.online/internal/infrastructure/config"
	"github.com/emiliopalmerini/quintaedizione.online/pkg/templates"
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

// RegisterRoutes registers all SRD HTTP routes on a chi-compatible http.ServeMux or router.
// The caller is expected to mount these under the appropriate prefix (e.g., /srd).
func (h *Handlers) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", h.home.handleHome)

	// Specific routes must be registered before wildcard routes
	mux.HandleFunc("GET /search", h.search.handleGlobalSearch)
	mux.HandleFunc("GET /search/dropdown", h.search.handleSearchDropdown)
	mux.HandleFunc("GET /robots.txt", h.seo.handleRobotsTxt)
	mux.HandleFunc("GET /sitemap.xml", h.seo.handleSitemap)

	// Collection routes with validation middleware
	mux.Handle("GET /{collection}", CollectionValidationMiddleware(http.HandlerFunc(h.collection.handleCollectionList)))
	mux.Handle("GET /{collection}/rows", CollectionValidationMiddleware(http.HandlerFunc(h.collection.handleCollectionRows)))
	mux.Handle("GET /{collection}/{slug}", CollectionValidationMiddleware(http.HandlerFunc(h.collection.handleItemDetail)))
}

// BaseHandler returns the base handler for middleware use.
func (h *Handlers) BaseHandler() *baseHandler {
	return h.home.baseHandler
}
