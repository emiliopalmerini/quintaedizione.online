package web

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/collections"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/search"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/models"
	"github.com/emiliopalmerini/quintaedizione.online/pkg/mappers"
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

	results, totalResults := h.transformSearchResults(r.Context(), fuzzyResults)

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

	content, err := h.templateEngine.RenderSearch(r.Context(), data)
	if err != nil {
		h.ErrorResponse(w, r, err, "Errore nel rendering della pagina di ricerca")
		return
	}

	h.renderHTML(w, content, "search")
}

// handleSearchDropdown renders the two-panel search dropdown.
// Without a query it shows browse mode; with a query it shows search results.
func (h *SearchHandler) handleSearchDropdown(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	collection := r.URL.Query().Get("collection")

	if query == "" {
		h.handleSearchBrowse(w, r)
		return
	}

	// Search mode: get results across all collections (or filtered)
	fuzzyResults, err := h.searchFuzzy(r.Context(), collection, query)
	if err != nil {
		h.renderHTML(w, "", "search")
		return
	}

	searchResults, _ := h.transformSearchResults(r.Context(), fuzzyResults)

	// Build sidebar collections from search results (only those with matches)
	allCollections := h.getAllCollectionsWithCounts(r.Context())

	// Active collection: only set if user explicitly chose one
	activeCollection := collection

	// Get documents: aggregated from all collections, or filtered to one
	var documents []models.Document
	var collectionLabel string
	if activeCollection != "" {
		for _, sr := range searchResults {
			if sr.CollectionName == activeCollection {
				documents = sr.Documents
				collectionLabel = sr.CollectionLabel
				break
			}
		}
	} else {
		// Aggregate results from all collections, tagging each doc with its collection
		for _, sr := range searchResults {
			for _, doc := range sr.Documents {
				doc.Collection = sr.CollectionName
				doc.CollectionLabel = sr.CollectionLabel
				documents = append(documents, doc)
			}
		}
	}

	// Show all collections in sidebar, with result counts for those that matched
	resultCountMap := make(map[string]int64)
	for _, sr := range searchResults {
		resultCountMap[sr.CollectionName] = sr.Total
	}
	sidebarCollections := make([]models.Collection, 0, len(allCollections))
	for _, col := range allCollections {
		sidebarCollections = append(sidebarCollections, models.Collection{
			Name:  col.Name,
			Label: col.Label,
			Count: resultCountMap[col.Name],
		})
	}

	data := models.SearchBrowseData{
		Collections:      sidebarCollections,
		ActiveCollection: activeCollection,
		Documents:        documents,
		CollectionName:   activeCollection,
		CollectionLabel:  collectionLabel,
		Query:            query,
	}

	content, err := h.templateEngine.RenderSearchBrowse(r.Context(), data)
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

// handleSearchBrowse renders the two-panel browse dropdown (on focus, before typing).
func (h *SearchHandler) handleSearchBrowse(w http.ResponseWriter, r *http.Request) {
	collection := r.URL.Query().Get("collection")

	allCollections := h.getAllCollectionsWithCounts(r.Context())

	// Default to incantesimi if none selected
	activeCollection := collection
	if activeCollection == "" {
		activeCollection = "incantesimi"
	}

	// Fetch items from the active collection (limited for the browse panel)
	var documents []models.Document
	var collectionLabel string
	if activeCollection != "" && h.contentService != nil {
		items, _, err := h.contentService.GetCollectionItems(r.Context(), activeCollection, "", nil, 1, 8)
		if err == nil {
			documents = make([]models.Document, 0, len(items))
			for _, item := range items {
				documents = append(documents, h.documentMapper.ToModel(activeCollection, item))
			}
		}
		collectionLabel = h.getCollectionTitle(activeCollection)
	}

	data := models.SearchBrowseData{
		Collections:      allCollections,
		ActiveCollection: activeCollection,
		Documents:        documents,
		CollectionName:   activeCollection,
		CollectionLabel:  collectionLabel,
	}

	content, err := h.templateEngine.RenderSearchBrowse(r.Context(), data)
	if err != nil {
		h.renderHTML(w, "", "search")
		return
	}

	h.renderHTML(w, content, "search")
}

// getAllCollectionsWithCounts returns all collections with their item counts.
func (h *SearchHandler) getAllCollectionsWithCounts(ctx context.Context) []models.Collection {
	if h.contentService == nil {
		return nil
	}
	stats, err := h.contentService.GetCollectionStats(ctx)
	if err != nil {
		return nil
	}

	result := make([]models.Collection, 0, len(stats))
	for _, col := range stats {
		name := mappers.GetString(col, "collection", "")
		count := mappers.GetInt64(col, "count", 0)
		if count > 0 {
			result = append(result, models.Collection{
				Name:  name,
				Label: h.getCollectionTitle(name),
				Count: count,
			})
		}
	}

	// Sort by count descending (most popular first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})

	return result
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
