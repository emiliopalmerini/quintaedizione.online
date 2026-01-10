package web

import (
	"fmt"
	"net/http"

	"github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web/models"
	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/search"
	"github.com/gin-gonic/gin"
)

// SearchHandler handles search-related requests.
type SearchHandler struct {
	*baseHandler
	searchService search.SearchService
}

// handleGlobalSearch renders the global search results page.
func (h *SearchHandler) handleGlobalSearch(c *gin.Context) {
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

	results, totalResults := h.transformSearchResults(fuzzyResults)

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

// handleSearchDropdown renders the search dropdown for autocomplete.
func (h *SearchHandler) handleSearchDropdown(c *gin.Context) {
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

	results, _ := h.transformSearchResults(fuzzyResults)

	content, err := h.templateEngine.RenderSearchDropdown(results, query)
	if err != nil {
		h.setCacheHeaders(c, "search")
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(""))
		return
	}

	h.setCacheHeaders(c, "search")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
}

// transformSearchResults converts fuzzy search results to CollectionSearchResult models.
// Applies collection title resolution and calculates hasMore flag.
// Returns the transformed results and total result count across all collections.
func (h *SearchHandler) transformSearchResults(fuzzyResults []search.SearchResultSet) ([]models.CollectionSearchResult, int64) {
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

	return results, totalResults
}
