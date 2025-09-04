package benchmarks

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/infrastructure"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

// BenchmarkHealthEndpoint benchmarks the health check endpoint
func BenchmarkHealthEndpoint(b *testing.B) {
	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		metrics := infrastructure.GetGlobalMetricsCollector().GetMetrics()
		cacheStats := infrastructure.GetGlobalCache().GetStats()
		
		c.JSON(http.StatusOK, gin.H{
			"status":       "healthy",
			"version":      "3.0.0-go",
			"architecture": "hexagonal",
			"performance": gin.H{
				"request_count":      metrics.RequestCount,
				"average_response":   metrics.AverageResponse.String(),
				"memory_usage_mb":    metrics.MemoryUsage / 1024 / 1024,
				"goroutine_count":    metrics.GoroutineCount,
				"active_connections": metrics.ActiveConns,
				"cache_hit_rate":     metrics.CacheHitRate,
				"cache_items":        cacheStats["item_count"],
			},
		})
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/health", nil)
			router.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				b.Errorf("Expected status 200, got %d", w.Code)
			}
		}
	})
}

// BenchmarkCacheOperations benchmarks cache operations
func BenchmarkCacheOperations(b *testing.B) {
	cache := infrastructure.NewSimpleCache()
	
	// Test data
	testData := map[string]interface{}{
		"nome":        "Test Item",
		"descrizione": "This is a test item for benchmarking",
		"livello":     5,
	}
	
	b.Run("CacheSet", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := "benchmark-key-" + string(rune(i%1000))
				cache.Set(key, testData, 5*time.Minute)
				i++
			}
		})
	})
	
	// Pre-populate cache for get benchmark
	for i := 0; i < 1000; i++ {
		key := "get-benchmark-key-" + string(rune(i))
		cache.Set(key, testData, 5*time.Minute)
	}
	
	b.Run("CacheGet", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := "get-benchmark-key-" + string(rune(i%1000))
				_, _ = cache.Get(key)
				i++
			}
		})
	})
}

// BenchmarkMetricsCollection benchmarks metrics collection
func BenchmarkMetricsCollection(b *testing.B) {
	collector := infrastructure.NewMetricsCollector()
	
	b.Run("RecordRequest", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				collector.RecordRequest(time.Millisecond * 50)
			}
		})
	})
	
	b.Run("GetMetrics", func(b *testing.B) {
		// Pre-record some data
		for i := 0; i < 100; i++ {
			collector.RecordRequest(time.Millisecond * 50)
		}
		
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = collector.GetMetrics()
			}
		})
	})
}

// BenchmarkJSONSerialization benchmarks JSON serialization of typical responses
func BenchmarkJSONSerialization(b *testing.B) {
	// Typical collection response
	items := make([]map[string]interface{}, 20)
	for i := 0; i < 20; i++ {
		items[i] = map[string]interface{}{
			"slug":               "test-item-" + string(rune(i)),
			"nome":               "Test Item " + string(rune(i)),
			"descrizione":        "This is a test item with a longer description to simulate realistic data",
			"livello":            i % 10,
			"fonte":              "SRD",
			"versione":           "1.0",
			"contenuto_markdown": "## Test Item\n\nThis is the full markdown content of the test item.\n\n**Bold text** and *italic text* for formatting.",
		}
	}
	
	response := map[string]interface{}{
		"title":       "Test Collection",
		"collection":  "test",
		"items":       items,
		"page":        1,
		"totalPages":  5,
		"totalCount":  100,
		"hasNext":     true,
		"hasPrev":     false,
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := json.Marshal(response)
			if err != nil {
				b.Errorf("JSON marshal error: %v", err)
			}
		}
	})
}

// BenchmarkPerformanceMiddleware benchmarks the performance middleware
func BenchmarkPerformanceMiddleware(b *testing.B) {
	middleware := infrastructure.PerformanceMiddleware()
	
	handler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
	
	router := gin.New()
	router.Use(middleware)
	router.GET("/test", handler)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				b.Errorf("Expected status 200, got %d", w.Code)
			}
		}
	})
}

// BenchmarkConcurrentRequests simulates concurrent request load
func BenchmarkConcurrentRequests(b *testing.B) {
	router := gin.New()
	router.Use(infrastructure.PerformanceMiddleware())
	
	// Simulate different endpoints
	router.GET("/", func(c *gin.Context) {
		time.Sleep(time.Millisecond * 1) // Simulate minimal processing
		c.JSON(http.StatusOK, gin.H{"page": "home"})
	})
	
	router.GET("/collection/:name", func(c *gin.Context) {
		time.Sleep(time.Millisecond * 5) // Simulate DB query
		c.JSON(http.StatusOK, gin.H{
			"collection": c.Param("name"),
			"items":      []string{"item1", "item2", "item3"},
		})
	})
	
	router.GET("/item/:collection/:slug", func(c *gin.Context) {
		time.Sleep(time.Millisecond * 2) // Simulate cached item retrieval
		c.JSON(http.StatusOK, gin.H{
			"collection": c.Param("collection"),
			"slug":       c.Param("slug"),
			"name":       "Test Item",
		})
	})
	
	endpoints := []string{
		"/",
		"/collection/spells",
		"/collection/monsters",
		"/item/spells/fireball",
		"/item/monsters/dragon",
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			endpoint := endpoints[i%len(endpoints)]
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", endpoint, nil)
			router.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				b.Errorf("Expected status 200, got %d for %s", w.Code, endpoint)
			}
			i++
		}
	})
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	b.Run("MapCreation", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				m := make(map[string]interface{})
				m["slug"] = "test-item"
				m["nome"] = "Test Item"
				m["descrizione"] = "Test description"
				m["livello"] = 5
				_ = m
			}
		})
	})
	
	b.Run("SliceCreation", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				slice := make([]map[string]interface{}, 0, 20)
				for i := 0; i < 20; i++ {
					item := map[string]interface{}{
						"slug": "item-" + string(rune(i)),
						"nome": "Item " + string(rune(i)),
					}
					slice = append(slice, item)
				}
				_ = slice
			}
		})
	})
}

// TestCacheHitRatio tests cache hit ratio under load
func TestCacheHitRatio(t *testing.T) {
	cache := infrastructure.NewSimpleCache()
	collector := infrastructure.NewMetricsCollector()
	
	// Pre-populate cache
	for i := 0; i < 100; i++ {
		key := "test-item-" + string(rune(i))
		cache.Set(key, map[string]interface{}{
			"nome": "Item " + string(rune(i)),
		}, 5*time.Minute)
	}
	
	// Simulate access pattern (80% cache hits)
	for i := 0; i < 1000; i++ {
		var key string
		if i%10 < 8 {
			// 80% chance of hitting existing keys
			key = "test-item-" + string(rune(i%100))
		} else {
			// 20% chance of missing
			key = "missing-item-" + string(rune(i))
		}
		
		_, found := cache.Get(key)
		if found {
			collector.RecordCacheHit()
		} else {
			collector.RecordCacheMiss()
		}
	}
	
	metrics := collector.GetMetrics()
	if metrics.CacheHitRate < 75.0 {
		t.Errorf("Expected cache hit rate >= 75%%, got %.2f%%", metrics.CacheHitRate)
	}
	
	t.Logf("Cache hit rate: %.2f%%", metrics.CacheHitRate)
}

// TestPerformanceMonitoring tests the performance monitoring system
func TestPerformanceMonitoring(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	// Start monitoring
	infrastructure.StartPerformanceMonitoring(ctx)
	
	// Simulate some activity
	collector := infrastructure.GetGlobalMetricsCollector()
	for i := 0; i < 10; i++ {
		collector.RecordRequest(time.Millisecond * 100)
		time.Sleep(time.Millisecond * 10)
	}
	
	metrics := collector.GetMetrics()
	
	if metrics.RequestCount != 10 {
		t.Errorf("Expected 10 requests, got %d", metrics.RequestCount)
	}
	
	if metrics.AverageResponse == 0 {
		t.Error("Expected non-zero average response time")
	}
	
	t.Logf("Metrics: %+v", metrics)
}