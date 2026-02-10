package web

import (
	"context"
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

	results, totalResults := h.transformSearchResults(c.Request.Context(), fuzzyResults)

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
	collection := c.Query("collection")

	if query == "" {
		h.setCacheHeaders(c, "search")
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(""))
		return
	}

	var fuzzyResults []search.SearchResultSet

	if collection != "" {
		collResults, err := h.searchService.SearchCollection(c.Request.Context(), collection, query, 3)
		if err != nil {
			h.setCacheHeaders(c, "search")
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(""))
			return
		}
		if len(collResults) > 0 {
			fuzzyResults = []search.SearchResultSet{{
				Collection: collection,
				Results:    collResults,
				Total:      int64(len(collResults)),
			}}
		}
	} else {
		var err error
		fuzzyResults, err = h.searchService.Search(c.Request.Context(), query, 3)
		if err != nil {
			h.setCacheHeaders(c, "search")
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(""))
			return
		}
	}

	results, _ := h.transformSearchResults(c.Request.Context(), fuzzyResults)

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
func (h *SearchHandler) transformSearchResults(ctx context.Context, fuzzyResults []search.SearchResultSet) ([]models.CollectionSearchResult, int64) {
	results := make([]models.CollectionSearchResult, 0, len(fuzzyResults))
	totalResults := int64(0)

	for _, sr := range fuzzyResults {
		documents := make([]models.Document, 0, len(sr.Results))
		for _, r := range sr.Results {
			item, err := h.contentService.GetItem(ctx, sr.Collection, r.ID)
			if err == nil {
				documents = append(documents, h.documentMapper.ToModel(sr.Collection, item))
			} else {
				documents = append(documents, models.Document{
					ID:    r.ID,
					Title: r.Title,
				})
			}
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
