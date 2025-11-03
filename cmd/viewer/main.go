package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/adapters/repositories"
	web "github.com/emiliopalmerini/due-draghi-5e-srd/internal/adapters/web"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/filters"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/services"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/infrastructure"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/infrastructure/database"
	pkgMongodb "github.com/emiliopalmerini/due-draghi-5e-srd/pkg/mongodb"
	"github.com/emiliopalmerini/due-draghi-5e-srd/pkg/templates"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	config := infrastructure.LoadConfig()

	// Setup logging
	if config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize MongoDB client
	mongoConfig := pkgMongodb.Config{
		URI:         config.MongoURI,
		Database:    config.DatabaseName,
		Timeout:     10 * time.Second,
		MaxPoolSize: 100,
	}

	mongoClient, err := pkgMongodb.NewClient(mongoConfig)
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}
	defer mongoClient.Close()

	// Check database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := mongoClient.Ping(ctx); err != nil {
		log.Fatalf("MongoDB health check failed: %v", err)
	}
	log.Println("✅ MongoDB connection established")

	// Initialize database indexes for optimal query performance
	indexManager := database.NewIndexManager(mongoClient)
	indexCtx, indexCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer indexCancel()

	if err := indexManager.EnsureIndexes(indexCtx); err != nil {
		log.Printf("⚠️  Failed to create indexes: %v", err)
		// Don't fail startup, indexes can be created later
	} else {
		log.Println("✅ Database indexes ensured")
	}

	// Initialize Templ engine
	var templateEngine *templates.TemplEngine
	if config.IsProduction() {
		templateEngine = templates.NewTemplEngine()
	} else {
		templateEngine = templates.NewDevTemplEngine()
	}
	log.Println("✅ Templates loaded")

	// Initialize repository factory
	repositoryFactory := repositories.NewRepositoryFactory(mongoClient)

	// Initialize filter registry
	filterRegistry, err := filters.NewYAMLFilterRegistry("configs/filters.yaml")
	if err != nil {
		log.Fatalf("Failed to initialize filter registry: %v", err)
	}
	log.Println("✅ Filter registry loaded")

	// Initialize filter service
	filterService := services.NewFilterService(filterRegistry)

	// Initialize services
	contentService := services.NewContentService(repositoryFactory.DocumentRepository(), filterService)

	// Initialize web handlers
	webHandlers := web.NewHandlers(contentService, templateEngine)

	// Setup Gin router
	router := gin.Default()

	// Middleware
	router.Use(web.RequestLoggingMiddleware())
	router.Use(web.MetricsMiddleware())
	router.Use(webHandlers.ErrorRecoveryMiddleware())
	router.Use(web.SecurityMiddleware())
	router.Use(web.ValidationMiddleware())
	router.Use(corsMiddleware())

	// Static files
	router.Static("/static", "./web/static")

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		cacheStats := infrastructure.GetGlobalCache().GetStats()
		metrics := web.GetGlobalMetrics()

		c.JSON(http.StatusOK, gin.H{
			"status":         "healthy",
			"version":        "3.0.0-go",
			"architecture":   "hexagonal",
			"database":       mongoClient.DatabaseName(),
			"cache_items":    cacheStats["item_count"],
			"uptime_seconds": time.Since(metrics.StartTime).Seconds(),
			"request_count":  metrics.RequestCount,
			"error_rate":     float64(metrics.ErrorCount) / max(float64(metrics.RequestCount), 1) * 100,
		})
	})

	// Detailed metrics endpoint
	router.GET("/admin/metrics", func(c *gin.Context) {
		metrics := web.GetGlobalMetrics()
		c.JSON(http.StatusOK, metrics.ToJSON())
	})

	// Register routes
	webHandlers.RegisterRoutes(router)

	// Setup HTTP server
	srv := &http.Server{
		Addr:    config.GetAddress(),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Starting D&D 5e SRD Viewer on %s", config.GetAddress())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down D&D 5e SRD Viewer...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server shutdown completed")
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}
