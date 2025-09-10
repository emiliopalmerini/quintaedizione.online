package web

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Metrics holds application performance metrics
type Metrics struct {
	mu               sync.RWMutex
	RequestCount     int64                    `json:"request_count"`
	ErrorCount       int64                    `json:"error_count"`
	TotalDuration    time.Duration            `json:"total_duration_ms"`
	AverageDuration  time.Duration            `json:"average_duration_ms"`
	EndpointMetrics  map[string]*EndpointStat `json:"endpoint_metrics"`
	CollectionStats  map[string]int64         `json:"collection_stats"`
	SearchStats      *SearchStats             `json:"search_stats"`
	StartTime        time.Time                `json:"start_time"`
	LastRequestTime  time.Time                `json:"last_request_time"`
}

// EndpointStat holds statistics for a specific endpoint
type EndpointStat struct {
	Count       int64         `json:"count"`
	TotalTime   time.Duration `json:"total_time_ms"`
	AverageTime time.Duration `json:"average_time_ms"`
	MinTime     time.Duration `json:"min_time_ms"`
	MaxTime     time.Duration `json:"max_time_ms"`
	ErrorCount  int64         `json:"error_count"`
}

// SearchStats holds search-related statistics
type SearchStats struct {
	TotalSearches    int64   `json:"total_searches"`
	EmptyQueries     int64   `json:"empty_queries"`
	AverageQueryTime time.Duration `json:"average_query_time_ms"`
	PopularTerms     map[string]int64 `json:"popular_terms"`
}

// Global metrics instance
var globalMetrics = &Metrics{
	EndpointMetrics: make(map[string]*EndpointStat),
	CollectionStats: make(map[string]int64),
	SearchStats: &SearchStats{
		PopularTerms: make(map[string]int64),
	},
	StartTime: time.Now(),
}

// GetGlobalMetrics returns the global metrics instance
func GetGlobalMetrics() *Metrics {
	return globalMetrics
}

// MetricsMiddleware tracks request metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		
		// Process request
		c.Next()
		
		// Calculate duration
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		
		// Update metrics
		globalMetrics.recordRequest(method, path, duration, statusCode)
		
		// Track collection-specific stats
		if collection := c.Param("collection"); collection != "" {
			globalMetrics.recordCollectionAccess(collection)
		}
		
		// Track search queries
		if q := c.Query("q"); q != "" {
			globalMetrics.recordSearch(q, duration)
		}
	}
}

// recordRequest updates request metrics
func (m *Metrics) recordRequest(method, path string, duration time.Duration, statusCode int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Overall metrics
	m.RequestCount++
	m.TotalDuration += duration
	m.AverageDuration = m.TotalDuration / time.Duration(m.RequestCount)
	m.LastRequestTime = time.Now()
	
	// Error tracking
	if statusCode >= 400 {
		m.ErrorCount++
	}
	
	// Endpoint-specific metrics
	endpoint := method + " " + path
	if stat, exists := m.EndpointMetrics[endpoint]; exists {
		stat.Count++
		stat.TotalTime += duration
		stat.AverageTime = stat.TotalTime / time.Duration(stat.Count)
		
		if duration < stat.MinTime || stat.MinTime == 0 {
			stat.MinTime = duration
		}
		if duration > stat.MaxTime {
			stat.MaxTime = duration
		}
		
		if statusCode >= 400 {
			stat.ErrorCount++
		}
	} else {
		m.EndpointMetrics[endpoint] = &EndpointStat{
			Count:       1,
			TotalTime:   duration,
			AverageTime: duration,
			MinTime:     duration,
			MaxTime:     duration,
			ErrorCount:  map[bool]int64{true: 1, false: 0}[statusCode >= 400],
		}
	}
}

// recordCollectionAccess tracks which collections are accessed most
func (m *Metrics) recordCollectionAccess(collection string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.CollectionStats[collection]++
}

// recordSearch tracks search queries and performance
func (m *Metrics) recordSearch(query string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.SearchStats.TotalSearches++
	
	if query == "" {
		m.SearchStats.EmptyQueries++
		return
	}
	
	// Update average query time
	if m.SearchStats.TotalSearches > 0 {
		totalTime := m.SearchStats.AverageQueryTime * time.Duration(m.SearchStats.TotalSearches-1)
		m.SearchStats.AverageQueryTime = (totalTime + duration) / time.Duration(m.SearchStats.TotalSearches)
	} else {
		m.SearchStats.AverageQueryTime = duration
	}
	
	// Track popular search terms (limit to top 100)
	if len(m.SearchStats.PopularTerms) < 100 {
		m.SearchStats.PopularTerms[query]++
	} else {
		// Only track if it's already a popular term
		if _, exists := m.SearchStats.PopularTerms[query]; exists {
			m.SearchStats.PopularTerms[query]++
		}
	}
}

// GetMetricsSnapshot returns a thread-safe copy of current metrics
func (m *Metrics) GetMetricsSnapshot() *Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	snapshot := &Metrics{
		RequestCount:    m.RequestCount,
		ErrorCount:      m.ErrorCount,
		TotalDuration:   m.TotalDuration,
		AverageDuration: m.AverageDuration,
		StartTime:       m.StartTime,
		LastRequestTime: m.LastRequestTime,
		EndpointMetrics: make(map[string]*EndpointStat),
		CollectionStats: make(map[string]int64),
		SearchStats: &SearchStats{
			TotalSearches:    m.SearchStats.TotalSearches,
			EmptyQueries:     m.SearchStats.EmptyQueries,
			AverageQueryTime: m.SearchStats.AverageQueryTime,
			PopularTerms:     make(map[string]int64),
		},
	}
	
	// Deep copy endpoint metrics
	for k, v := range m.EndpointMetrics {
		snapshot.EndpointMetrics[k] = &EndpointStat{
			Count:       v.Count,
			TotalTime:   v.TotalTime,
			AverageTime: v.AverageTime,
			MinTime:     v.MinTime,
			MaxTime:     v.MaxTime,
			ErrorCount:  v.ErrorCount,
		}
	}
	
	// Copy collection stats
	for k, v := range m.CollectionStats {
		snapshot.CollectionStats[k] = v
	}
	
	// Copy popular terms
	for k, v := range m.SearchStats.PopularTerms {
		snapshot.SearchStats.PopularTerms[k] = v
	}
	
	return snapshot
}

// ResetMetrics clears all metrics (useful for testing)
func (m *Metrics) ResetMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.RequestCount = 0
	m.ErrorCount = 0
	m.TotalDuration = 0
	m.AverageDuration = 0
	m.EndpointMetrics = make(map[string]*EndpointStat)
	m.CollectionStats = make(map[string]int64)
	m.SearchStats = &SearchStats{
		PopularTerms: make(map[string]int64),
	}
	m.StartTime = time.Now()
	m.LastRequestTime = time.Time{}
}

// ToJSON converts metrics to JSON-serializable format
func (m *Metrics) ToJSON() map[string]any {
	snapshot := m.GetMetricsSnapshot()
	
	uptime := time.Since(snapshot.StartTime)
	
	// Convert durations to milliseconds for JSON
	endpointMetricsJSON := make(map[string]map[string]any)
	for endpoint, stat := range snapshot.EndpointMetrics {
		endpointMetricsJSON[endpoint] = map[string]any{
			"count":        stat.Count,
			"total_time":   stat.TotalTime.Milliseconds(),
			"average_time": stat.AverageTime.Milliseconds(),
			"min_time":     stat.MinTime.Milliseconds(),
			"max_time":     stat.MaxTime.Milliseconds(),
			"error_count":  stat.ErrorCount,
			"error_rate":   float64(stat.ErrorCount) / float64(stat.Count) * 100,
		}
	}
	
	return map[string]any{
		"uptime_seconds":     uptime.Seconds(),
		"request_count":      snapshot.RequestCount,
		"error_count":        snapshot.ErrorCount,
		"error_rate":         float64(snapshot.ErrorCount) / float64(snapshot.RequestCount) * 100,
		"average_duration":   snapshot.AverageDuration.Milliseconds(),
		"requests_per_second": float64(snapshot.RequestCount) / uptime.Seconds(),
		"endpoint_metrics":   endpointMetricsJSON,
		"collection_stats":   snapshot.CollectionStats,
		"search_stats": map[string]any{
			"total_searches":       snapshot.SearchStats.TotalSearches,
			"empty_queries":        snapshot.SearchStats.EmptyQueries,
			"average_query_time":   snapshot.SearchStats.AverageQueryTime.Milliseconds(),
			"popular_terms":        snapshot.SearchStats.PopularTerms,
		},
		"start_time":      snapshot.StartTime,
		"last_request":    snapshot.LastRequestTime,
		"generated_at":    time.Now(),
	}
}