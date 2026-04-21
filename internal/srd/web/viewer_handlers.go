package web

import (
	"fmt"
	"net/http"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/application/services"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/search"
	infraconfig "github.com/emiliopalmerini/quintaedizione.online/internal/srd/infrastructure/config"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/display"
	webmappers "github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/mappers"
	"github.com/emiliopalmerini/quintaedizione.online/pkg/templates"
)

// Handlers coordinates all specialized handlers.
type Handlers struct {
	home          *HomeHandler
	area          *AreaHandler
	collection    *CollectionHandler
	search        *SearchHandler
	seo           *SEOHandler
	defaultSource string // default source ID for legacy URL redirects
}

// NewHandlers creates all specialized handlers with shared dependencies.
// defaultSource is the source ID used for legacy URL redirects (e.g. "srd-5.5e").
func NewHandlers(contentService *services.ContentService, searchService search.SearchService, templateEngine *templates.TemplEngine, defaultSource string, multiSource bool, versionResolver display.VersionResolver, opts ...HandlersOption) *Handlers {
	var displayOpts []func(*display.DisplayElementFactory)
	if versionResolver != nil {
		displayOpts = append(displayOpts, display.WithVersionResolver(versionResolver))
	}
	displayFactory := display.NewDisplayElementFactory(multiSource, displayOpts...)
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

	h := &Handlers{
		home: &HomeHandler{
			baseHandler: base,
		},
		area: &AreaHandler{
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
		defaultSource: defaultSource,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// HandlersOption configures optional dependencies for Handlers.
type HandlersOption func(*Handlers)

// WithSearchRecorder sets a SearchRecorder for tracking search metrics.
func WithSearchRecorder(r SearchRecorder) HandlersOption {
	return func(h *Handlers) {
		h.search.searchRecorder = r
	}
}

// RegisterRoutes registers all SRD HTTP routes on a chi-compatible http.ServeMux or router.
// The caller is expected to mount these under the appropriate prefix (e.g., /srd).
func (h *Handlers) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", h.home.handleHome)

	// Specific routes must be registered before wildcard routes
	mux.HandleFunc("GET /search", h.search.handleGlobalSearch)
	mux.HandleFunc("GET /search/dropdown", h.search.handleSearchDropdown)
	mux.HandleFunc("GET /area/{slug}", h.area.handleArea)
	mux.HandleFunc("GET /robots.txt", h.seo.handleRobotsTxt)
	mux.HandleFunc("GET /sitemap.xml", h.seo.handleSitemap)

	// Collection routes with validation middleware.
	// Note: rows endpoint lives at /rows/{collection} (instead of /{collection}/rows) to avoid
	// ambiguity with /area/{slug}; Go's ServeMux refuses to register overlapping patterns.
	mux.Handle("GET /rows/{collection}", CollectionValidationMiddleware(http.HandlerFunc(h.collection.handleCollectionRows)))
	mux.Handle("GET /{collection}", CollectionValidationMiddleware(http.HandlerFunc(h.collection.handleCollectionList)))
	mux.Handle("GET /{collection}/{source}/{slug}", CollectionValidationMiddleware(http.HandlerFunc(h.collection.handleItemDetail)))

	// Legacy redirect: /{collection}/{slug} → /{collection}/{defaultSource}/{slug}
	if h.defaultSource != "" {
		mux.Handle("GET /{collection}/{slug}", CollectionValidationMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			collection := r.PathValue("collection")
			slug := r.PathValue("slug")
			http.Redirect(w, r, "/srd/"+collection+"/"+h.defaultSource+"/"+slug, http.StatusMovedPermanently)
		})))
	}
}
