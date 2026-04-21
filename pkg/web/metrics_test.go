package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/", "/"},
		{"/health", "/health"},
		{"/srd/search", "/srd/search"},
		{"/srd/incantesimi", "/srd/{collection}"},
		{"/srd/mostri", "/srd/{collection}"},
		{"/srd/rows/incantesimi", "/srd/rows/{collection}"},
		{"/srd/area/personaggi", "/srd/area/{slug}"},
		{"/srd/incantesimi/srd55/fireball", "/srd/{collection}/{source}/{slug}"},
		{"/srd/mostri/srd55/dragon", "/srd/{collection}/{source}/{slug}"},
		{"/mappe/some-map", "/mappe/{slug}"},
		{"/generatori/tavern-names", "/generatori/{slug}"},
		{"/generatori/tavern-names/roll", "/generatori/{slug}/roll"},
		{"/combattimenti/", "/combattimenti/"},
		{"/combattimenti/calculate", "/combattimenti/calculate"},
		{"/combattimenti/api/monsters", "/combattimenti/api/monsters"},
		{"/combattimenti/api/difficulties", "/combattimenti/api/difficulties"},
		{"/combattimenti/party-input", "/combattimenti/party-input"},
		{"/static/css/style.css", "/static/*"},
		{"/static/js/app.js", "/static/*"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := NormalizePath(tt.path)
			if got != tt.want {
				t.Errorf("NormalizePath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestMetricsMiddleware_IncrementsRequestTotal(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	handler := m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	count := testutil.ToFloat64(m.RequestsTotal.WithLabelValues("GET", "/health", "200"))
	if count != 1 {
		t.Errorf("expected request count 1, got %v", count)
	}
}

func TestMetricsMiddleware_RecordsDuration(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	handler := m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Verify histogram was observed by checking the metric count via the registry
	metrics, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}
	var found bool
	for _, mf := range metrics {
		if mf.GetName() == "http_request_duration_seconds" {
			found = true
			for _, m := range mf.GetMetric() {
				if m.GetHistogram().GetSampleCount() != 1 {
					t.Errorf("expected 1 observation, got %d", m.GetHistogram().GetSampleCount())
				}
			}
		}
	}
	if !found {
		t.Error("expected http_request_duration_seconds metric to be present")
	}
}

func TestMetricsMiddleware_TracksInFlight(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	var inFlightDuringRequest float64
	handler := m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inFlightDuringRequest = testutil.ToFloat64(m.InFlight)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if inFlightDuringRequest != 1 {
		t.Errorf("expected in-flight 1 during request, got %v", inFlightDuringRequest)
	}

	afterRequest := testutil.ToFloat64(m.InFlight)
	if afterRequest != 0 {
		t.Errorf("expected in-flight 0 after request, got %v", afterRequest)
	}
}

func TestMetricsMiddleware_NormalizesPathLabels(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	handler := m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Two different slugs should map to the same label
	req1 := httptest.NewRequest(http.MethodGet, "/mappe/map-one", nil)
	req2 := httptest.NewRequest(http.MethodGet, "/mappe/map-two", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req1)
	handler.ServeHTTP(httptest.NewRecorder(), req2)

	count := testutil.ToFloat64(m.RequestsTotal.WithLabelValues("GET", "/mappe/{slug}", "200"))
	if count != 2 {
		t.Errorf("expected 2 requests for normalized path, got %v", count)
	}
}

func TestSearchMetrics_IncrementsCounter(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	m.RecordSearch("global")
	m.RecordSearch("global")
	m.RecordSearch("collection")

	globalCount := testutil.ToFloat64(m.SearchQueries.WithLabelValues("global"))
	if globalCount != 2 {
		t.Errorf("expected global search count 2, got %v", globalCount)
	}

	collCount := testutil.ToFloat64(m.SearchQueries.WithLabelValues("collection"))
	if collCount != 1 {
		t.Errorf("expected collection search count 1, got %v", collCount)
	}
}

func TestStatusRecorder_CapturesStatusCode(t *testing.T) {
	rec := httptest.NewRecorder()
	sr := &statusRecorder{ResponseWriter: rec, statusCode: http.StatusOK}

	sr.WriteHeader(http.StatusNotFound)

	if sr.statusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", sr.statusCode)
	}
}
