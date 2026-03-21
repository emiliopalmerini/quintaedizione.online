package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/collections"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/search"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/models"
	"github.com/emiliopalmerini/quintaedizione.online/pkg/mappers"
	"github.com/emiliopalmerini/quintaedizione.online/web/templates"
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
		http.Redirect(w, r, "/srd", http.StatusFound)
		return
	}

	fuzzyResults, err := h.searchService.Search(r.Context(), query, 5)
	if err != nil {
		h.ErrorResponse(w, r, err, "Errore durante la ricerca")
		return
	}

	results, totalResults := h.transformSearchResults(r.Context(), fuzzyResults, query)

	data := models.SearchPageData{
		PageData: models.PageData{
			Title:       fmt.Sprintf("Risultati per: %s", query),
			Description: "Risultati della ricerca globale",
			QueryString: r.URL.RawQuery,
		},
		Query:       query,
		Results:     results,
		Total:       totalResults,
		Collections: h.getPopularCollections(),
	}

	content, err := h.templateEngine.RenderSearch(data)
	if err != nil {
		h.ErrorResponse(w, r, err, "Errore nel rendering della pagina di ricerca")
		return
	}

	h.renderHTML(w, content, "search")
}

// handleSearchDropdown renders the search dropdown for autocomplete.
func (h *SearchHandler) handleSearchDropdown(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	collection := r.URL.Query().Get("collection")

	if query == "" {
		h.renderHTML(w, "", "search")
		return
	}

	fuzzyResults, err := h.searchFuzzy(r.Context(), collection, query)
	if err != nil {
		h.renderHTML(w, "", "search")
		return
	}

	results, _ := h.transformSearchResults(r.Context(), fuzzyResults, query)

	content, err := h.templateEngine.RenderSearchDropdown(results, query)
	if err != nil {
		h.renderHTML(w, "", "search")
		return
	}

	h.renderHTML(w, content, "search")
}

// searchFuzzy performs a fuzzy search, scoped to a collection if specified.
func (h *SearchHandler) searchFuzzy(ctx context.Context, collection, query string) ([]search.SearchResultSet, error) {
	if collection == "" {
		return h.searchService.Search(ctx, query, 3)
	}

	collResults, err := h.searchService.SearchCollection(ctx, collection, query, 3)
	if err != nil {
		return nil, err
	}
	if len(collResults) == 0 {
		return nil, nil
	}
	return []search.SearchResultSet{{
		Collection: collection,
		Results:    collResults,
		Total:      int64(len(collResults)),
	}}, nil
}

// transformSearchResults converts fuzzy search results to CollectionSearchResult models.
func (h *SearchHandler) transformSearchResults(ctx context.Context, fuzzyResults []search.SearchResultSet, query string) ([]models.CollectionSearchResult, int64) {
	results := make([]models.CollectionSearchResult, 0, len(fuzzyResults))
	totalResults := int64(0)

	for _, sr := range fuzzyResults {
		documents := make([]models.Document, 0, len(sr.Results))
		for _, r := range sr.Results {
			item, err := h.contentService.GetItem(ctx, sr.Collection, r.ID)
			if err == nil {
				doc := h.documentMapper.ToModel(sr.Collection, item)
				if rawContent := mappers.GetString(item, "raw_content", ""); rawContent != "" {
					doc.Snippet = templates.ExtractSnippet(rawContent, query, 120)
				}
				documents = append(documents, doc)
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

// getPopularCollections returns a list of popular collections for the empty state.
func (h *SearchHandler) getPopularCollections() []models.Collection {
	popular := []string{"incantesimi", "mostri", "classi", "oggetti_magici", "equipaggiamenti"}
	result := make([]models.Collection, 0, len(popular))
	for _, name := range popular {
		if _, ok := collections.FromString(name); ok {
			result = append(result, models.Collection{
				Name:  name,
				Label: h.getCollectionTitle(name),
			})
		}
	}
	return result
}
