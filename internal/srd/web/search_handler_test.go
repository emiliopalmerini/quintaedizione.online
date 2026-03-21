package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/search"
	"github.com/emiliopalmerini/quintaedizione.online/pkg/templates"
)

type mockSearchService struct {
	searchCalled           bool
	searchCollectionCalled bool
	searchCollectionArg    string
}

func (m *mockSearchService) Search(_ context.Context, _ string, _ int) ([]search.SearchResultSet, error) {
	m.searchCalled = true
	return nil, nil
}

func (m *mockSearchService) SearchCollection(_ context.Context, collection, _ string, _ int) ([]search.SearchResult, error) {
	m.searchCollectionCalled = true
	m.searchCollectionArg = collection
	return nil, nil
}

func (m *mockSearchService) RefreshIndex(_ context.Context) error {
	return nil
}

func newTestSearchHandler(svc search.SearchService) *SearchHandler {
	return &SearchHandler{
		baseHandler: &baseHandler{
			templateEngine: templates.NewTemplEngine(),
		},
		searchService: svc,
	}
}

func TestSearchDropdown_AllCollections(t *testing.T) {
	mock := &mockSearchService{}
	handler := newTestSearchHandler(mock)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/search/dropdown?q=fireball", nil)
	handler.handleSearchDropdown(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !mock.searchCalled {
		t.Error("expected Search() to be called for all-collection query")
	}
	if mock.searchCollectionCalled {
		t.Error("SearchCollection() should not be called without collection param")
	}
}

func TestSearchDropdown_SingleCollection(t *testing.T) {
	mock := &mockSearchService{}
	handler := newTestSearchHandler(mock)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/search/dropdown?q=fireball&collection=incantesimi", nil)
	handler.handleSearchDropdown(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if mock.searchCalled {
		t.Error("Search() should not be called when collection param is set")
	}
	if !mock.searchCollectionCalled {
		t.Error("expected SearchCollection() to be called with collection param")
	}
	if mock.searchCollectionArg != "incantesimi" {
		t.Errorf("SearchCollection collection = %q, want %q", mock.searchCollectionArg, "incantesimi")
	}
}

func TestSearchDropdown_EmptyQuery(t *testing.T) {
	mock := &mockSearchService{}
	handler := newTestSearchHandler(mock)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/search/dropdown?q=", nil)
	handler.handleSearchDropdown(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if mock.searchCalled || mock.searchCollectionCalled {
		t.Error("no search method should be called for empty query")
	}
	if w.Body.String() != "" {
		t.Errorf("expected empty body for empty query, got %q", w.Body.String())
	}
}
