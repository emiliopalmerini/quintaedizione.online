package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web/models"
	"github.com/emiliopalmerini/quintaedizione.online/internal/domain/search"
)

// SearchHandler handles search-related requests.
type SearchHandler struct {
	*baseHandler
	searchService search.SearchService
}

// handleGlobalSearch renders the global search results page.
func (h *SearchHandler) handleGlobalSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	if query == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	fuzzyResults, err := h.searchService.Search(r.Context(), query, 5)
	if err != nil {
		h.ErrorResponse(w, r, err, "Errore durante la ricerca")
		return
	}

	results, totalResults := h.transformSearchResults(r.Context(), fuzzyResults)

	data := models.SearchPageData{
		PageData: models.PageData{
			Title:       fmt.Sprintf("Risultati per: %s", query),
			Description: "Risultati della ricerca globale",
			QueryString: r.URL.RawQuery,
		},
		Query:   query,
		Results: results,
		Total:   totalResults,
	}

	content, err := h.templateEngine.RenderSearch(data)
	if err != nil {
		h.ErrorResponse(w, r, err, "Errore nel rendering della pagina di ricerca")
		return
	}

	h.setCacheHeaders(w, "search")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(content))
}

// handleSearchDropdown renders the search dropdown for autocomplete.
func (h *SearchHandler) handleSearchDropdown(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	collection := r.URL.Query().Get("collection")

	if query == "" {
		h.setCacheHeaders(w, "search")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(""))
		return
	}

	var fuzzyResults []search.SearchResultSet

	if collection != "" {
		collResults, err := h.searchService.SearchCollection(r.Context(), collection, query, 3)
		if err != nil {
			h.setCacheHeaders(w, "search")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(""))
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
		fuzzyResults, err = h.searchService.Search(r.Context(), query, 3)
		if err != nil {
			h.setCacheHeaders(w, "search")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(""))
			return
		}
	}

	results, _ := h.transformSearchResults(r.Context(), fuzzyResults)

	content, err := h.templateEngine.RenderSearchDropdown(results, query)
	if err != nil {
		h.setCacheHeaders(w, "search")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(""))
		return
	}

	h.setCacheHeaders(w, "search")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(content))
}

// transformSearchResults converts fuzzy search results to CollectionSearchResult models.
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
