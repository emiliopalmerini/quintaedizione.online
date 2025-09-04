package web

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/services"
	"github.com/emiliopalmerini/due-draghi-5e-srd/pkg/templates"
)

// Handlers contains web handlers for the editor
type Handlers struct {
	contentService *services.ContentService
	templateEngine *templates.Engine
}

// NewHandlers creates a new Handlers instance
func NewHandlers(contentService *services.ContentService, templateEngine *templates.Engine) *Handlers {
	return &Handlers{
		contentService: contentService,
		templateEngine: templateEngine,
	}
}

// RegisterRoutes registers all web routes
func (h *Handlers) RegisterRoutes(router *gin.Engine) {
	// Main page
	router.GET("/", h.handleHome)
	
	// Collection pages
	router.GET("/c/:collection", h.handleCollectionList)
	router.GET("/c/:collection/:slug", h.handleItemDetail)
	
	// Search functionality
	router.GET("/search", h.handleSearch)
	router.POST("/search", h.handleSearchPost)
	
	// Admin routes
	admin := router.Group("/admin")
	{
		admin.GET("/", h.handleAdminHome)
		admin.GET("/collections", h.handleAdminCollections)
		admin.POST("/sync", h.handleAdminSync)
	}
}

// handleHome renders the home page
func (h *Handlers) handleHome(c *gin.Context) {
	data := gin.H{
		"title":       "D&D 5e SRD Italiano",
		"description": "System Reference Document di Dungeons & Dragons 5a Edizione in italiano",
		"collections": []string{"incantesimi", "mostri", "classi", "backgrounds", "equipaggiamento"},
	}
	
	h.renderTemplate(c, "home.html", data)
}

// handleCollectionList renders a collection list page
func (h *Handlers) handleCollectionList(c *gin.Context) {
	collection := c.Param("collection")
	page := c.DefaultQuery("page", "1")
	search := c.Query("search")
	
	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}
	
	// Get items from service
	items, totalCount, err := h.contentService.GetCollectionItems(c.Request.Context(), collection, search, pageNum, 20)
	if err != nil {
		h.renderError(c, "Errore nel caricamento della collezione", http.StatusInternalServerError)
		return
	}
	
	// Calculate pagination
	totalPages := int((totalCount + 19) / 20) // Ceiling division
	
	data := gin.H{
		"title":       getCollectionTitle(collection),
		"collection":  collection,
		"items":       items,
		"search":      search,
		"page":        pageNum,
		"totalPages":  totalPages,
		"totalCount":  totalCount,
		"hasNext":     pageNum < totalPages,
		"hasPrev":     pageNum > 1,
	}
	
	h.renderTemplate(c, "collection.html", data)
}

// handleItemDetail renders an item detail page
func (h *Handlers) handleItemDetail(c *gin.Context) {
	collection := c.Param("collection")
	slug := c.Param("slug")
	
	item, err := h.contentService.GetItem(c.Request.Context(), collection, slug)
	if err != nil {
		h.renderError(c, "Elemento non trovato", http.StatusNotFound)
		return
	}
	
	data := gin.H{
		"title":      item["nome"],
		"collection": collection,
		"item":       item,
	}
	
	h.renderTemplate(c, "item.html", data)
}

// handleSearch handles search requests
func (h *Handlers) handleSearch(c *gin.Context) {
	query := c.Query("q")
	collections := c.QueryArray("collections")
	
	if query == "" {
		h.renderTemplate(c, "search.html", gin.H{"title": "Cerca"})
		return
	}
	
	if len(collections) == 0 {
		collections = []string{"incantesimi", "mostri", "classi", "backgrounds", "equipaggiamento"}
	}
	
	results, err := h.contentService.Search(c.Request.Context(), query, collections, 50)
	if err != nil {
		h.renderError(c, "Errore nella ricerca", http.StatusInternalServerError)
		return
	}
	
	data := gin.H{
		"title":       "Risultati ricerca",
		"query":       query,
		"collections": collections,
		"results":     results,
		"resultCount": len(results),
	}
	
	h.renderTemplate(c, "search.html", data)
}

// handleSearchPost handles POST search requests (HTMX)
func (h *Handlers) handleSearchPost(c *gin.Context) {
	var form struct {
		Query       string   `form:"q" binding:"required"`
		Collections []string `form:"collections"`
	}
	
	if err := c.ShouldBind(&form); err != nil {
		h.renderError(c, "Parametri di ricerca non validi", http.StatusBadRequest)
		return
	}
	
	if len(form.Collections) == 0 {
		form.Collections = []string{"incantesimi", "mostri", "classi", "backgrounds", "equipaggiamento"}
	}
	
	results, err := h.contentService.Search(c.Request.Context(), form.Query, form.Collections, 50)
	if err != nil {
		h.renderError(c, "Errore nella ricerca", http.StatusInternalServerError)
		return
	}
	
	data := gin.H{
		"query":       form.Query,
		"results":     results,
		"resultCount": len(results),
	}
	
	h.renderTemplate(c, "search_results.html", data)
}

// Admin handlers
func (h *Handlers) handleAdminHome(c *gin.Context) {
	stats, err := h.contentService.GetStats(c.Request.Context())
	if err != nil {
		h.renderError(c, "Errore nel caricamento delle statistiche", http.StatusInternalServerError)
		return
	}
	
	data := gin.H{
		"title": "Amministrazione",
		"stats": stats,
	}
	
	h.renderTemplate(c, "admin/home.html", data)
}

func (h *Handlers) handleAdminCollections(c *gin.Context) {
	collections, err := h.contentService.GetCollectionStats(c.Request.Context())
	if err != nil {
		h.renderError(c, "Errore nel caricamento delle collezioni", http.StatusInternalServerError)
		return
	}
	
	data := gin.H{
		"title":       "Collezioni",
		"collections": collections,
	}
	
	h.renderTemplate(c, "admin/collections.html", data)
}

func (h *Handlers) handleAdminSync(c *gin.Context) {
	// TODO: Implement sync functionality
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Sincronizzazione completata",
	})
}

// Helper methods
func (h *Handlers) renderTemplate(c *gin.Context, template string, data gin.H) {
	content, err := h.templateEngine.Render(template, data)
	if err != nil {
		h.renderError(c, "Errore nel rendering della pagina", http.StatusInternalServerError)
		return
	}
	
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
}

func (h *Handlers) renderError(c *gin.Context, message string, statusCode int) {
	data := gin.H{
		"title":        "Errore",
		"error":        message,
		"status_code":  statusCode,
	}
	
	content, err := h.templateEngine.Render("error.html", data)
	if err != nil {
		// Fallback to simple error response
		c.String(statusCode, "Errore: %s", message)
		return
	}
	
	c.Data(statusCode, "text/html; charset=utf-8", []byte(content))
}

func getCollectionTitle(collection string) string {
	titles := map[string]string{
		"incantesimi":     "Incantesimi",
		"mostri":          "Mostri",
		"classi":          "Classi",
		"backgrounds":     "Background",
		"equipaggiamento": "Equipaggiamento",
		"armi":            "Armi",
		"armature":        "Armature",
		"oggetti_magici":  "Oggetti Magici",
		"talenti":         "Talenti",
		"servizi":         "Servizi",
		"strumenti":       "Strumenti",
		"animali":         "Animali",
	}
	
	if title, exists := titles[collection]; exists {
		return title
	}
	
	return collection
}