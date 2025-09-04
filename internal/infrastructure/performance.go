package infrastructure

import (
	"context"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// PerformanceMetrics holds system performance data
type PerformanceMetrics struct {
	RequestCount    int64         `json:"request_count"`
	AverageResponse time.Duration `json:"average_response_ms"`
	MemoryUsage     uint64        `json:"memory_usage_bytes"`
	GoroutineCount  int           `json:"goroutine_count"`
	CPUPercent      float64       `json:"cpu_percent"`
	ActiveConns     int           `json:"active_connections"`
	CacheHitRate    float64       `json:"cache_hit_rate"`
}

// MetricsCollector manages performance metrics collection
type MetricsCollector struct {
	mu              sync.RWMutex
	requestCount    int64
	responseTimes   []time.Duration
	activeConns     int64
	cacheHits       int64
	cacheMisses     int64
	lastCPUTime     time.Time
	lastCPUUsage    time.Duration
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		responseTimes: make([]time.Duration, 0, 1000),
		lastCPUTime:   time.Now(),
	}
}

// RecordRequest records a request and its response time
func (mc *MetricsCollector) RecordRequest(responseTime time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.requestCount++
	mc.responseTimes = append(mc.responseTimes, responseTime)
	
	// Keep only last 1000 response times for memory efficiency
	if len(mc.responseTimes) > 1000 {
		mc.responseTimes = mc.responseTimes[len(mc.responseTimes)-1000:]
	}
}

// IncrementActiveConnections increments the active connection counter
func (mc *MetricsCollector) IncrementActiveConnections() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.activeConns++
}

// DecrementActiveConnections decrements the active connection counter
func (mc *MetricsCollector) DecrementActiveConnections() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if mc.activeConns > 0 {
		mc.activeConns--
	}
}

// RecordCacheHit records a cache hit
func (mc *MetricsCollector) RecordCacheHit() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.cacheHits++
}

// RecordCacheMiss records a cache miss
func (mc *MetricsCollector) RecordCacheMiss() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.cacheMisses++
}

// GetMetrics returns current performance metrics
func (mc *MetricsCollector) GetMetrics() PerformanceMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	var avgResponse time.Duration
	if len(mc.responseTimes) > 0 {
		var total time.Duration
		for _, rt := range mc.responseTimes {
			total += rt
		}
		avgResponse = total / time.Duration(len(mc.responseTimes))
	}
	
	// Get memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// Calculate cache hit rate
	var cacheHitRate float64
	totalCacheOps := mc.cacheHits + mc.cacheMisses
	if totalCacheOps > 0 {
		cacheHitRate = float64(mc.cacheHits) / float64(totalCacheOps) * 100
	}
	
	return PerformanceMetrics{
		RequestCount:    mc.requestCount,
		AverageResponse: avgResponse,
		MemoryUsage:     m.Alloc,
		GoroutineCount:  runtime.NumGoroutine(),
		CPUPercent:      mc.calculateCPUPercent(),
		ActiveConns:     int(mc.activeConns),
		CacheHitRate:    cacheHitRate,
	}
}

// calculateCPUPercent calculates CPU usage percentage
func (mc *MetricsCollector) calculateCPUPercent() float64 {
	// This is a simplified CPU calculation
	// In production, you'd want to use a proper CPU monitoring library
	return float64(runtime.NumGoroutine()) / 1000.0 * 100
}

// Global metrics collector instance
var globalMetricsCollector = NewMetricsCollector()

// GetGlobalMetricsCollector returns the global metrics collector
func GetGlobalMetricsCollector() *MetricsCollector {
	return globalMetricsCollector
}

// PerformanceMiddleware is a Gin middleware that tracks performance metrics
func PerformanceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Increment active connections
		globalMetricsCollector.IncrementActiveConnections()
		defer globalMetricsCollector.DecrementActiveConnections()
		
		// Process request
		c.Next()
		
		// Record response time
		responseTime := time.Since(start)
		globalMetricsCollector.RecordRequest(responseTime)
		
		// Add performance headers
		c.Header("X-Response-Time", responseTime.String())
		c.Header("X-Request-ID", generateRequestID())
	}
}

// SimpleCache provides basic in-memory caching with TTL
type SimpleCache struct {
	mu    sync.RWMutex
	items map[string]cacheItem
}

type cacheItem struct {
	value      interface{}
	expiration time.Time
}

// NewSimpleCache creates a new simple cache
func NewSimpleCache() *SimpleCache {
	cache := &SimpleCache{
		items: make(map[string]cacheItem),
	}
	
	// Start cleanup goroutine
	go cache.cleanupExpired()
	
	return cache
}

// Set adds an item to the cache with TTL
func (sc *SimpleCache) Set(key string, value interface{}, ttl time.Duration) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	sc.items[key] = cacheItem{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
}

// Get retrieves an item from the cache
func (sc *SimpleCache) Get(key string) (interface{}, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	
	item, exists := sc.items[key]
	if !exists {
		globalMetricsCollector.RecordCacheMiss()
		return nil, false
	}
	
	if time.Now().After(item.expiration) {
		globalMetricsCollector.RecordCacheMiss()
		return nil, false
	}
	
	globalMetricsCollector.RecordCacheHit()
	return item.value, true
}

// Delete removes an item from the cache
func (sc *SimpleCache) Delete(key string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	delete(sc.items, key)
}

// Clear removes all items from the cache
func (sc *SimpleCache) Clear() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.items = make(map[string]cacheItem)
}

// cleanupExpired removes expired items from cache
func (sc *SimpleCache) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		sc.mu.Lock()
		now := time.Now()
		for key, item := range sc.items {
			if now.After(item.expiration) {
				delete(sc.items, key)
			}
		}
		sc.mu.Unlock()
	}
}

// GetStats returns cache statistics
func (sc *SimpleCache) GetStats() map[string]interface{} {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	
	return map[string]interface{}{
		"item_count": len(sc.items),
	}
}

// Global cache instance
var globalCache = NewSimpleCache()

// GetGlobalCache returns the global cache instance
func GetGlobalCache() *SimpleCache {
	return globalCache
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + time.Now().Format("000")
}

// OptimizeForProduction sets production-ready optimizations
func OptimizeForProduction() {
	// Set GC target percentage for better memory management
	runtime.GOMAXPROCS(runtime.NumCPU())
	
	// Log optimization settings
	log.Printf("Production optimizations enabled - GOMAXPROCS: %d", runtime.GOMAXPROCS(0))
}

// StartPerformanceMonitoring starts background performance monitoring
func StartPerformanceMonitoring(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				metrics := globalMetricsCollector.GetMetrics()
				log.Printf("Performance Metrics - Requests: %d, Avg Response: %s, Memory: %d MB, Goroutines: %d",
					metrics.RequestCount,
					metrics.AverageResponse,
					metrics.MemoryUsage/1024/1024,
					metrics.GoroutineCount)
			}
		}
	}()
}