package web

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Metrics holds all Prometheus metrics for the application.
type Metrics struct {
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	InFlight        prometheus.Gauge
	SearchQueries   *prometheus.CounterVec
}

// NewMetrics creates and registers all application metrics with the given registry.
func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests.",
			},
			[]string{"method", "path", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		InFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Number of HTTP requests currently being served.",
			},
		),
		SearchQueries: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "search_queries_total",
				Help: "Total number of search queries.",
			},
			[]string{"type"},
		),
	}

	reg.MustRegister(m.RequestsTotal, m.RequestDuration, m.InFlight, m.SearchQueries)
	return m
}

// RecordSearch increments the search query counter for the given search type.
func (m *Metrics) RecordSearch(searchType string) {
	m.SearchQueries.WithLabelValues(searchType).Inc()
}

// Middleware returns an HTTP middleware that records request metrics.
func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := NormalizePath(r.URL.Path)

		m.InFlight.Inc()
		defer m.InFlight.Dec()

		start := time.Now()
		sr := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(sr, r)

		duration := time.Since(start).Seconds()
		status := fmt.Sprintf("%d", sr.statusCode)

		m.RequestsTotal.WithLabelValues(r.Method, path, status).Inc()
		m.RequestDuration.WithLabelValues(r.Method, path).Observe(duration)
	})
}

// statusRecorder wraps http.ResponseWriter to capture the status code.
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.statusCode = code
	sr.ResponseWriter.WriteHeader(code)
}

// knownSRDCollections is the set of valid SRD collection names for path normalization.
var knownSRDCollections = map[string]bool{
	"incantesimi":         true,
	"mostri":              true,
	"animali":             true,
	"classi":              true,
	"backgrounds":         true,
	"equipaggiamenti":     true,
	"oggetti_magici":      true,
	"armi":                true,
	"armature":            true,
	"strumenti":           true,
	"cavalcature_veicoli": true,
	"servizi":             true,
	"talenti":             true,
	"regole":              true,
	"specie":              true,
}

// NormalizePath converts a request path into a parameterized label
// to prevent high-cardinality Prometheus labels.
func NormalizePath(path string) string {
	if path == "/" || path == "" {
		return "/"
	}

	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")

	switch parts[0] {
	case "static":
		return "/static/*"

	case "srd":
		if len(parts) == 1 {
			return "/srd"
		}
		// /srd/search, /srd/search/dropdown — keep as-is
		if parts[1] == "search" {
			return "/" + strings.Join(parts, "/")
		}
		// /srd/area/{slug}
		if len(parts) == 3 && parts[1] == "area" {
			return "/srd/area/{slug}"
		}
		// /srd/rows/{collection}
		if len(parts) == 3 && parts[1] == "rows" && knownSRDCollections[parts[2]] {
			return "/srd/rows/{collection}"
		}
		// /srd/{collection}
		if len(parts) == 2 && knownSRDCollections[parts[1]] {
			return "/srd/{collection}"
		}
		// /srd/{collection}/{source}/{slug}
		if len(parts) == 4 && knownSRDCollections[parts[1]] {
			return "/srd/{collection}/{source}/{slug}"
		}
		return "/srd/{collection}"

	case "mappe":
		if len(parts) == 1 {
			return "/mappe"
		}
		return "/mappe/{slug}"

	case "generatori":
		if len(parts) == 1 {
			return "/generatori"
		}
		if len(parts) == 3 && parts[2] == "roll" {
			return "/generatori/{slug}/roll"
		}
		return "/generatori/{slug}"

	case "combattimenti":
		return "/" + strings.Join(parts, "/")

	default:
		return "/" + parts[0]
	}
}
