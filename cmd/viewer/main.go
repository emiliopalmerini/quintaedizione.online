package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	jsondata "github.com/emiliopalmerini/quintaedizione.online/data/ita/json"
	"github.com/emiliopalmerini/quintaedizione.online/internal/adapters/repositories/inmemory"
	web "github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web"
	"github.com/emiliopalmerini/quintaedizione.online/internal/application/filters"
	"github.com/emiliopalmerini/quintaedizione.online/internal/application/parsers"
	"github.com/emiliopalmerini/quintaedizione.online/internal/application/search"
	"github.com/emiliopalmerini/quintaedizione.online/internal/application/services"
	"github.com/emiliopalmerini/quintaedizione.online/internal/infrastructure"
	"github.com/emiliopalmerini/quintaedizione.online/internal/infrastructure/datastore"
	"github.com/emiliopalmerini/quintaedizione.online/pkg/templates"
	"github.com/gin-gonic/gin"
)

func main() {
	config := infrastructure.LoadConfig()

	if config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Load JSON data into in-memory store
	log.Println("Loading JSON data...")
	glossaryLinker, err := parsers.NewGlossaryLinker(jsondata.Files)
	if err != nil {
		log.Printf("Warning: Failed to initialize glossary linker: %v", err)
	}
	renderer := parsers.NewMarkdownRenderer(glossaryLinker)
	loader := datastore.NewLoader(jsondata.Files, renderer)
	data, err := loader.LoadAll()
	if err != nil {
		log.Fatalf("Failed to load JSON data: %v", err)
	}
	store := datastore.NewStore(data)
	log.Println("Data loaded into in-memory store")

	// Log collection counts
	for _, name := range store.Collections() {
		log.Printf("  %s: %d documents", name, store.Count(name))
	}

	// Create repositories
	repo := inmemory.NewDocumentRepository(store)
	searchRepo := inmemory.NewSearchRepository(store)

	// Initialize template engine
	var templateEngine *templates.TemplEngine
	if config.IsProduction() {
		templateEngine = templates.NewTemplEngine()
	} else {
		templateEngine = templates.NewDevTemplEngine()
	}
	log.Println("Templates loaded")

	// Initialize filter registry and services
	filterRegistry := filters.NewInMemoryFilterRegistry()
	filters.RegisterDefaultFilters(filterRegistry)
	log.Println("Filter registry loaded")

	filterService := services.NewFilterService(filterRegistry)

	cache := infrastructure.NewSimpleCache()

	contentService := services.NewContentService(repo, filterService, cache)

	searchService := search.NewFuzzySearchService(searchRepo)
	log.Println("Fuzzy search service initialized")

	// Set up web layer
	webHandlers := web.NewHandlers(contentService, searchService, templateEngine)

	router := gin.Default()

	rateLimiter := web.NewRateLimiter()

	router.Use(web.RequestLoggingMiddleware())
	router.Use(web.MetricsMiddleware())
	router.Use(webHandlers.ErrorRecoveryMiddleware())
	router.Use(web.SecurityMiddleware())
	router.Use(web.RateLimitMiddleware(rateLimiter))
	router.Use(web.ValidationMiddleware())
	router.Use(corsMiddleware())

	router.Static("/static", "./web/static")

	router.GET("/health", func(c *gin.Context) {
		cacheStats := cache.GetStats()
		metrics := web.GetGlobalMetrics()

		c.JSON(http.StatusOK, gin.H{
			"status":         "healthy",
			"version":        "4.0.0",
			"architecture":   "hexagonal-inmemory",
			"cache_items":    cacheStats["item_count"],
			"uptime_seconds": time.Since(metrics.StartTime).Seconds(),
			"request_count":  metrics.RequestCount,
			"error_rate":     float64(metrics.ErrorCount) / max(float64(metrics.RequestCount), 1) * 100,
		})
	})

	webHandlers.RegisterRoutes(router)

	srv := &http.Server{
		Addr:              config.GetAddress(),
		Handler:           router,
		ReadHeaderTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("Starting Quintaedizione 5e SRD Viewer on %s", config.GetAddress())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Quintaedizione 5e SRD Viewer...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server shutdown completed")
}

func corsMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		allowedOrigins := map[string]bool{
			"https://quintaedizione.online":     true,
			"https://www.quintaedizione.online": true,
		}

		if !isProduction() {
			allowedOrigins["http://localhost:3000"] = true
			allowedOrigins["http://localhost:8000"] = true
			allowedOrigins["http://127.0.0.1:3000"] = true
			allowedOrigins["http://127.0.0.1:8000"] = true
		}

		if allowedOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Max-Age", "3600")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}

func isProduction() bool {
	env := os.Getenv("ENVIRONMENT")
	return env == "production"
}
