package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/application/services"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/collections"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/domain/filters"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/infrastructure/datastore"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/infrastructure/persistence"
	"github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/display"
	webmappers "github.com/emiliopalmerini/quintaedizione.online/internal/srd/web/mappers"
	"github.com/emiliopalmerini/quintaedizione.online/pkg/templates"
)

// noopFilterService is a minimal FilterService for handler tests.
type noopFilterService struct{}

func (n *noopFilterService) ParseFilters(_ collections.CollectionName, _ map[string]string) (*filters.FilterSet, error) {
	return &filters.FilterSet{}, nil
}
func (n *noopFilterService) ValidateFilterSet(_ *filters.FilterSet) error { return nil }
func (n *noopFilterService) BuildFilter(_ *filters.FilterSet) (filters.DocumentPredicate, error) {
	return nil, nil
}
func (n *noopFilterService) GetAvailableFilters(_ collections.CollectionName) ([]filters.FilterDefinition, error) {
	return nil, nil
}
func (n *noopFilterService) BuildSearchPredicate(_ collections.CollectionName, _ string) filters.DocumentPredicate {
	return nil
}
func (n *noopFilterService) CombinePredicates(preds ...filters.DocumentPredicate) filters.DocumentPredicate {
	return nil
}

// newTestCollectionHandler creates a CollectionHandler backed by a real Store for acceptance tests.
func newTestCollectionHandler(data map[string][]map[string]any) *CollectionHandler {
	store := datastore.NewStore(data)
	repo := persistence.NewDocumentRepository(store)
	contentService := services.NewContentService(repo, &noopFilterService{})
	return &CollectionHandler{
		baseHandler: &baseHandler{
			contentService: contentService,
			templateEngine: templates.NewTemplEngine(),
			documentMapper: webmappers.NewDocumentMapper(display.NewDisplayElementFactory(true)),
		},
	}
}

func TestHandleCollectionListRendersCanonicalFullAndPartialResponses(t *testing.T) {
	handler := newTestCollectionHandler(map[string][]map[string]any{
		"incantesimi": {
			{"_id": "luce", "_source_short": "5.5e", "title": "Luce"},
		},
	})

	tests := []struct {
		name       string
		hxRequest  string
		hxTarget   string
		wantLayout bool
	}{
		{name: "normal request", wantLayout: true},
		{name: "boosted navigation", hxRequest: "true", hxTarget: "page-root", wantLayout: true},
		{name: "rows update", hxRequest: "true", hxTarget: "rows", wantLayout: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/srd/incantesimi?q=luce", nil)
			req.SetPathValue("collection", "incantesimi")
			if tt.hxRequest != "" {
				req.Header.Set("HX-Request", tt.hxRequest)
				req.Header.Set("HX-Target", tt.hxTarget)
			}

			handler.handleCollectionList(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
			}
			hasLayout := strings.Contains(w.Body.String(), "<!doctype html>")
			if hasLayout != tt.wantLayout {
				t.Errorf("layout presence = %t, want %t", hasLayout, tt.wantLayout)
			}
			vary := w.Header().Values("Vary")
			if !strings.Contains(strings.Join(vary, ","), "HX-Request") || !strings.Contains(strings.Join(vary, ","), "HX-Target") {
				t.Errorf("expected HTMX request headers in Vary, got %v", vary)
			}
		})
	}
}

func TestExtractFiltersJoinsRepeatedValues(t *testing.T) {
	handler := &CollectionHandler{}
	req := httptest.NewRequest(http.MethodGet, "/srd/incantesimi?q=fuoco&page=2&livello=0&livello=1&scuola=&scuola=Invocazione", nil)

	got := handler.extractFilters(req)

	if got["livello"] != "0,1" {
		t.Errorf("expected repeated levels to be joined, got %q", got["livello"])
	}
	if got["scuola"] != "Invocazione" {
		t.Errorf("expected empty filter values to be ignored, got %q", got["scuola"])
	}
	if _, exists := got["q"]; exists {
		t.Error("search query must not be extracted as a filter")
	}
	if _, exists := got["page"]; exists {
		t.Error("page must not be extracted as a filter")
	}
}

func TestHandleItemDetail_MultiVersion_ShowsVersionTabs(t *testing.T) {
	handler := newTestCollectionHandler(map[string][]map[string]any{
		"incantesimi": {
			{"_id": "palla-di-fuoco", "_source_short": "5.5e", "_source": "srd-5.5e", "title": "Palla di Fuoco", "content": "<p>5.5e</p>", "raw_content": "5.5e"},
			{"_id": "palla-di-fuoco", "_source_short": "5e", "_source": "srd-5e", "title": "Palla di Fuoco", "content": "<p>5e</p>", "raw_content": "5e"},
		},
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/srd/incantesimi/5.5e/palla-di-fuoco", nil)
	req.SetPathValue("collection", "incantesimi")
	req.SetPathValue("source", "5.5e")
	req.SetPathValue("slug", "palla-di-fuoco")
	handler.handleItemDetail(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "version-switcher") {
		t.Error("expected version-switcher element in response for multi-version document")
	}
	if !strings.Contains(body, "5e") || !strings.Contains(body, "5.5e") {
		t.Error("expected both edition labels in version switcher")
	}
}

func TestHandleItemDetail_SingleVersion_NoVersionTabs(t *testing.T) {
	// Store has two documents but with different slugs; "luce" exists in only one source.
	handler := newTestCollectionHandler(map[string][]map[string]any{
		"incantesimi": {
			{"_id": "luce", "_source_short": "5.5e", "_source": "srd-5.5e", "title": "Luce", "content": "<p>Luce</p>", "raw_content": "Luce"},
			{"_id": "palla-di-fuoco", "_source_short": "5.5e", "_source": "srd-5.5e", "title": "Palla di Fuoco", "content": "<p>Fire</p>", "raw_content": "Fire"},
			{"_id": "palla-di-fuoco", "_source_short": "5e", "_source": "srd-5e", "title": "Palla di Fuoco", "content": "<p>Fire old</p>", "raw_content": "Fire old"},
		},
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/srd/incantesimi/5.5e/luce", nil)
	req.SetPathValue("collection", "incantesimi")
	req.SetPathValue("source", "5.5e")
	req.SetPathValue("slug", "luce")
	handler.handleItemDetail(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if strings.Contains(body, "version-switcher") {
		t.Error("expected no version-switcher for single-version document")
	}
}

func TestHandleItemDetail_VersionTabLinks(t *testing.T) {
	handler := newTestCollectionHandler(map[string][]map[string]any{
		"incantesimi": {
			{"_id": "palla-di-fuoco", "_source_short": "5.5e", "_source": "srd-5.5e", "title": "Palla di Fuoco", "content": "<p>5.5e</p>", "raw_content": "5.5e"},
			{"_id": "palla-di-fuoco", "_source_short": "5e", "_source": "srd-5e", "title": "Palla di Fuoco", "content": "<p>5e</p>", "raw_content": "5e"},
		},
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/srd/incantesimi/5.5e/palla-di-fuoco", nil)
	req.SetPathValue("collection", "incantesimi")
	req.SetPathValue("source", "5.5e")
	req.SetPathValue("slug", "palla-di-fuoco")
	handler.handleItemDetail(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()

	// Version switcher must exist for multi-version documents
	if !strings.Contains(body, "version-switcher") {
		t.Fatal("expected version-switcher element; cannot verify tab links without it")
	}

	// The active tab should show the current edition
	if !strings.Contains(body, `aria-selected="true"`) {
		t.Error("expected active tab with aria-selected=true for current edition")
	}

	// The inactive tab should link to the other edition via hx-get
	if !strings.Contains(body, `hx-get="/srd/incantesimi/5e/palla-di-fuoco"`) {
		t.Error("expected hx-get link to 5e version in version switcher")
	}
}

func TestBuildQuickFilterData_Incantesimi(t *testing.T) {
	data := buildQuickFilterData("incantesimi", map[string]string{})
	if data == nil {
		t.Fatal("expected quick filter data for incantesimi")
	}
	if data.FilterName != "livello" {
		t.Errorf("expected filter name 'livello', got %q", data.FilterName)
	}
	if len(data.Chips) != 10 {
		t.Errorf("expected 10 chips, got %d", len(data.Chips))
	}
	for _, chip := range data.Chips {
		if chip.Active {
			t.Errorf("chip %q should not be active with no filters", chip.Label)
		}
	}
}

func TestBuildQuickFilterData_ActiveChip(t *testing.T) {
	data := buildQuickFilterData("incantesimi", map[string]string{"livello": "0"})
	if data == nil {
		t.Fatal("expected quick filter data")
	}
	if !data.Chips[0].Active {
		t.Error("Trucchetto chip should be active when livello=0")
	}
	if data.Chips[1].Active {
		t.Error("1° chip should not be active when livello=0")
	}
}

func TestBuildQuickFilterData_MultiValueChipActive(t *testing.T) {
	// Mostri CR range "0-¼" maps to values "0", "1/8", "1/4"
	// All three must be in the filter for the chip to be active
	data := buildQuickFilterData("mostri", map[string]string{"grado_sfida": "0,1/8,1/4"})
	if data == nil {
		t.Fatal("expected quick filter data for mostri")
	}
	if !data.Chips[0].Active {
		t.Error("0-¼ chip should be active when all its values are selected")
	}
}

func TestBuildQuickFilterData_MultiValueChipPartial(t *testing.T) {
	// Only 2 of 3 values selected — chip should NOT be active
	data := buildQuickFilterData("mostri", map[string]string{"grado_sfida": "0,1/8"})
	if data == nil {
		t.Fatal("expected quick filter data for mostri")
	}
	if data.Chips[0].Active {
		t.Error("0-¼ chip should not be active when only 2 of 3 values selected")
	}
}

func TestBuildQuickFilterData_UnknownCollection(t *testing.T) {
	data := buildQuickFilterData("nonexistent", map[string]string{})
	if data != nil {
		t.Error("expected nil for unknown collection")
	}
}

func TestBuildQuickFilterData_CollectionWithoutQuickFilter(t *testing.T) {
	data := buildQuickFilterData("regole", map[string]string{})
	if data != nil {
		t.Error("expected nil for collection without quick filter")
	}
}

func TestChipMatchesFilter(t *testing.T) {
	tests := []struct {
		name         string
		chipValues   []string
		activeValues map[string]bool
		want         bool
	}{
		{"all match", []string{"a", "b"}, map[string]bool{"a": true, "b": true, "c": true}, true},
		{"partial match", []string{"a", "b"}, map[string]bool{"a": true}, false},
		{"no match", []string{"a"}, map[string]bool{"b": true}, false},
		{"single match", []string{"a"}, map[string]bool{"a": true}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := chipMatchesFilter(tt.chipValues, tt.activeValues)
			if got != tt.want {
				t.Errorf("chipMatchesFilter(%v, %v) = %v, want %v", tt.chipValues, tt.activeValues, got, tt.want)
			}
		})
	}
}
