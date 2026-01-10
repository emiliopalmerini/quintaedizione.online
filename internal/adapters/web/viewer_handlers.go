package web

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web/config"
	"github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web/display"
	webmappers "github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web/mappers"
	"github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web/models"
	"github.com/emiliopalmerini/quintaedizione.online/internal/application/services"
	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/collections"
	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/search"
	infraconfig "github.com/emiliopalmerini/quintaedizione.online/internal/infrastructure/config"
	"github.com/emiliopalmerini/quintaedizione.online/pkg/mappers"
	"github.com/emiliopalmerini/quintaedizione.online/pkg/templates"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	contentService     *services.ContentService
	searchService      search.SearchService
	templateEngine     *templates.TemplEngine
	documentMapper     webmappers.DocumentMapper
	collectionMetadata infraconfig.CollectionMetadata
}

func NewHandlers(contentService *services.ContentService, searchService search.SearchService, templateEngine *templates.TemplEngine) *Handlers {
	displayFactory := display.NewDisplayElementFactory()
	documentMapper := webmappers.NewDocumentMapper(displayFactory)

	collectionMetadata, err := infraconfig.NewCollectionMetadata()
	if err != nil {
		fmt.Printf("Warning: Failed to load collection metadata: %v\n", err)
	}

	return &Handlers{
		contentService:     contentService,
		searchService:      searchService,
		templateEngine:     templateEngine,
		documentMapper:     documentMapper,
		collectionMetadata: collectionMetadata,
	}
}

func (h *Handlers) RegisterRoutes(router *gin.Engine) {

	router.GET("/", h.handleHome)

	// Specific routes must be registered before wildcard routes
	router.GET("/search", h.handleGlobalSearch)
	router.GET("/search/dropdown", h.handleSearchDropdown)
	router.GET("/robots.txt", h.handleRobotsTxt)
	router.GET("/sitemap.xml", h.handleSitemap)

	// Wildcard routes (must come last)
	router.GET("/:collection", h.handleCollectionList)
	router.GET("/:collection/rows", h.handleCollectionRows)
	router.GET("/:collection/:slug", h.handleItemDetail)
}

func (h *Handlers) handleHome(c *gin.Context) {

	collections, err := h.contentService.GetCollectionStats(c.Request.Context())
	if err != nil {

		collections = h.getDefaultCollections()
	}

	typedCollections := make([]models.Collection, 0, len(collections))
	total := int64(0)

	for _, col := range collections {
		name := mappers.GetString(col, "collection", "")
		count := mappers.GetInt64(col, "count", 0)

		collection := models.Collection{
			Name:  name,
			Count: count,
		}

		if count > 0 {
			total += count
		}

		collection.Label = h.getCollectionTitle(name)

		typedCollections = append(typedCollections, collection)
	}

	data := models.HomePageData{
		PageData: models.PageData{
			Title:       "quintaedizione.online",
			Description: "Il Fantastico Visualizzatore di SRD (5e 2024)",
		},
		Collections: typedCollections,
		Total:       total,
	}

	content, err := h.templateEngine.RenderHome(data)
	if err != nil {
		h.ErrorResponse(c, err, "Errore nel rendering della pagina home")
		return
	}

	h.setCacheHeaders(c, "home")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
}

func (h *Handlers) handleCollectionList(c *gin.Context) {
	collection := c.Param("collection")
	page := c.DefaultQuery("page", "1")
	q := c.Query("q")
	pageSize := c.DefaultQuery("page_size", "20")

	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	pageSizeNum, err := strconv.Atoi(pageSize)
	if err != nil || pageSizeNum < 1 || pageSizeNum > 100 {
		pageSizeNum = 20
	}

	filters := h.extractFilters(c)

	rawItems, totalCount, err := h.contentService.GetCollectionItems(c.Request.Context(), collection, q, filters, pageNum, pageSizeNum)
	if err != nil {
		h.ErrorResponse(c, err, fmt.Sprintf("Errore nel caricamento della collezione %s", collection))
		return
	}

	documents := h.documentMapper.ToModels(collection, rawItems)

	pagination := CalculatePaginationData(pageNum, pageSizeNum, totalCount)

	data := models.CollectionPageData{
		PageData: models.PageData{
			Title:       h.getCollectionTitle(collection),
			Collection:  collection,
			QueryString: c.Request.URL.RawQuery,
		},
		Documents:  documents,
		Query:      q,
		Page:       pageNum,
		PageSize:   pageSizeNum,
		Total:      totalCount,
		TotalPages: pagination.TotalPages,
		HasNext:    pagination.HasNext,
		HasPrev:    pagination.HasPrev,
		StartItem:  pagination.StartItem,
		EndItem:    pagination.EndItem,
	}

	content, err := h.templateEngine.RenderCollection(data)
	if err != nil {
		h.ErrorResponse(c, err, "Errore nel rendering della pagina collezione")
		return
	}

	h.setCacheHeaders(c, "collection")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
}

func (h *Handlers) handleItemDetail(c *gin.Context) {
	collection := c.Param("collection")
	slug := c.Param("slug")

	item, err := h.contentService.GetItem(c.Request.Context(), collection, slug)
	if err != nil {
		h.ErrorResponse(c, err, "Elemento non trovato")
		return
	}

	bodyHTML := mappers.GetString(item, "content", "")
	bodyRaw := mappers.GetString(item, "raw_content", "")

	prevSlug, nextSlug, err := h.contentService.GetAdjacentItems(c.Request.Context(), collection, slug)
	if err != nil {

		fmt.Printf("Warning: Could not get adjacent items for %s/%s: %v\n", collection, slug, err)
	}

	docTitle := mappers.GetString(item, "title", "")

	prevID := ""
	if prevSlug != nil {
		prevID = *prevSlug
	}
	nextID := ""
	if nextSlug != nil {
		nextID = *nextSlug
	}

	data := models.ItemPageData{
		PageData: models.PageData{
			Title:       docTitle,
			DocTitle:    docTitle,
			DocID:       slug,
			Collection:  collection,
			QueryString: c.Request.URL.RawQuery,
		},
		BodyRaw:         bodyRaw,
		BodyHTML:        bodyHTML,
		PrevID:          prevID,
		NextID:          nextID,
		CollectionLabel: h.getCollectionTitle(collection),
	}

	content, err := h.templateEngine.RenderItem(data)
	if err != nil {
		h.ErrorResponse(c, err, "Errore nel rendering della pagina elemento")
		return
	}

	h.setCacheHeaders(c, "item")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
}

func (h *Handlers) handleCollectionRows(c *gin.Context) {
	collection := c.Param("collection")
	page := c.DefaultQuery("page", "1")
	q := c.Query("q")
	pageSize := c.DefaultQuery("page_size", "20")

	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	pageSizeNum, err := strconv.Atoi(pageSize)
	if err != nil || pageSizeNum < 1 || pageSizeNum > 100 {
		pageSizeNum = 20
	}

	filters := h.extractFilters(c)

	rawItems, totalCount, err := h.contentService.GetCollectionItems(c.Request.Context(), collection, q, filters, pageNum, pageSizeNum)
	if err != nil {
		h.ErrorResponse(c, err, fmt.Sprintf("Errore nel caricamento righe per %s", collection))
		return
	}

	documents := h.documentMapper.ToModels(collection, rawItems)

	pagination := CalculatePaginationData(pageNum, pageSizeNum, totalCount)

	data := models.CollectionPageData{
		PageData: models.PageData{
			Collection:  collection,
			QueryString: c.Request.URL.RawQuery,
		},
		Documents:  documents,
		Query:      q,
		Page:       pageNum,
		PageSize:   pageSizeNum,
		Total:      totalCount,
		TotalPages: pagination.TotalPages,
		HasNext:    pagination.HasNext,
		HasPrev:    pagination.HasPrev,
		StartItem:  pagination.StartItem,
		EndItem:    pagination.EndItem,
	}

	content, err := h.templateEngine.RenderRows(data)
	if err != nil {
		h.ErrorResponse(c, err, "Errore nel rendering delle righe")
		return
	}

	h.setCacheHeaders(c, "collection")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
}

func formatTraitContent(content string) string {

	formatted := content

	formatted = strings.ReplaceAll(formatted, "### Talenti Generali", "")
	formatted = strings.ReplaceAll(formatted, "### Talenti Razziali", "")
	formatted = strings.ReplaceAll(formatted, "### Categoria Background", "")

	formatted = regexp.MustCompile(`(\s)(\*\*\*[^*]+\.\*\*\*)`).ReplaceAllString(formatted, "\n\n$2")
	formatted = regexp.MustCompile(`(\s)(\*\*[^*]+\.\*\*)`).ReplaceAllString(formatted, "\n\n$2")

	formatted = regexp.MustCompile(`\n{3,}`).ReplaceAllString(formatted, "\n\n")
	formatted = strings.TrimSpace(formatted)

	return formatted
}

func (h *Handlers) setCacheHeaders(c *gin.Context, cacheTypeStr string) {

	var cacheType config.CacheType
	switch cacheTypeStr {
	case "home":
		cacheType = config.CacheTypeHome
	case "collection":
		cacheType = config.CacheTypeCollection
	case "item":
		cacheType = config.CacheTypeItem
	case "search":
		cacheType = config.CacheTypeSearch
	default:
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

func (h *Handlers) getCollectionTitle(collection string) string {

	if h.collectionMetadata != nil {
		return h.collectionMetadata.GetTitle(collection)
	}

	return config.GetCollectionTitle(collection)
}

func (h *Handlers) extractFilters(c *gin.Context) map[string]string {
	filters := make(map[string]string)

	skipParams := map[string]bool{
		"page":      true,
		"page_size": true,
		"q":         true,
	}

	for param, values := range c.Request.URL.Query() {
		if !skipParams[param] && len(values) > 0 && values[0] != "" {
			filters[param] = values[0]
		}
	}

	return filters
}

func (h *Handlers) getDefaultCollections() []map[string]any {
	allCollections := collections.GetAllWithInfo()
	result := make([]map[string]any, 0, len(allCollections))

	for _, info := range allCollections {
		result = append(result, map[string]any{
			"name":  info.Name.String(),
			"label": info.Title,
			"count": 0,
		})
	}

	return result
}

func (h *Handlers) handleGlobalSearch(c *gin.Context) {
	query := c.Query("q")

	if query == "" {
		c.Redirect(http.StatusFound, "/")
		return
	}

	fuzzyResults, err := h.searchService.Search(c.Request.Context(), query, 5)
	if err != nil {
		h.ErrorResponse(c, err, "Errore durante la ricerca")
		return
	}

	results := make([]models.CollectionSearchResult, 0, len(fuzzyResults))
	totalResults := int64(0)

	for _, sr := range fuzzyResults {
		documents := make([]models.Document, 0, len(sr.Results))
		for _, r := range sr.Results {
			documents = append(documents, models.Document{
				ID:    r.ID,
				Title: r.Title,
			})
		}

		results = append(results, models.CollectionSearchResult{
			CollectionName:  sr.Collection,
			CollectionLabel: h.getCollectionTitle(sr.Collection),
			Documents:       documents,
			Total:           sr.Total,
			HasMore:         sr.Total > int64(len(sr.Results)),
		})

		totalResults += sr.Total
	}

	data := models.SearchPageData{
		PageData: models.PageData{
			Title:       fmt.Sprintf("Risultati per: %s", query),
			Description: "Risultati della ricerca globale",
			QueryString: c.Request.URL.RawQuery,
		},
		Query:   query,
		Results: results,
		Total:   totalResults,
	}

	content, err := h.templateEngine.RenderSearch(data)
	if err != nil {
		h.ErrorResponse(c, err, "Errore nel rendering della pagina di ricerca")
		return
	}

	h.setCacheHeaders(c, "search")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
}

func (h *Handlers) handleSearchDropdown(c *gin.Context) {
	query := c.Query("q")

	if query == "" {
		h.setCacheHeaders(c, "search")
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(""))
		return
	}

	fuzzyResults, err := h.searchService.Search(c.Request.Context(), query, 3)
	if err != nil {
		h.setCacheHeaders(c, "search")
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(""))
		return
	}

	results := make([]models.CollectionSearchResult, 0, len(fuzzyResults))

	for _, sr := range fuzzyResults {
		documents := make([]models.Document, 0, len(sr.Results))
		for _, r := range sr.Results {
			documents = append(documents, models.Document{
				ID:    r.ID,
				Title: r.Title,
			})
		}

		results = append(results, models.CollectionSearchResult{
			CollectionName:  sr.Collection,
			CollectionLabel: h.getCollectionTitle(sr.Collection),
			Documents:       documents,
			Total:           sr.Total,
			HasMore:         sr.Total > int64(len(sr.Results)),
		})
	}

	content, err := h.templateEngine.RenderSearchDropdown(results, query)
	if err != nil {
		h.setCacheHeaders(c, "search")
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(""))
		return
	}

	h.setCacheHeaders(c, "search")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
}

func (h *Handlers) handleRobotsTxt(c *gin.Context) {
	c.Header("Cache-Control", "max-age=86400, public")
	c.File("./web/static/robots.txt")
}

func (h *Handlers) handleSitemap(c *gin.Context) {
	sitemap := h.generateSitemap()
	c.Header("Cache-Control", "max-age=86400, public")
	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(sitemap))
}

func (h *Handlers) generateSitemap() string {
	baseURL := "https://quintaedizione.online"
	sitemap := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>` + baseURL + `/</loc>
    <changefreq>weekly</changefreq>
    <priority>1.0</priority>
  </url>`

	collections := getValidCollections()
	for _, collection := range collections {
		sitemap += `
  <url>
    <loc>` + baseURL + `/` + collection + `</loc>
    <changefreq>weekly</changefreq>
    <priority>0.8</priority>
  </url>`
	}

	sitemap += `
</urlset>`
	return sitemap
}
