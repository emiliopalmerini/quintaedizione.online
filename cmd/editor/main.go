package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/adapters/web"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/services"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/infrastructure"
	"github.com/emiliopalmerini/due-draghi-5e-srd/pkg/mongodb"
	"github.com/emiliopalmerini/due-draghi-5e-srd/pkg/templates"
)

func main() {
	// Load configuration
	config := infrastructure.LoadConfig()

	// Setup logging and optimizations
	if config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
		infrastructure.OptimizeForProduction()
	}

	// Initialize MongoDB client
	mongoConfig := mongodb.Config{
		URI:         config.MongoURI,
		Database:    config.DatabaseName,
		Timeout:     10 * time.Second,
		MaxPoolSize: 100,
	}

	mongoClient, err := mongodb.NewClient(mongoConfig)
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

	// Initialize template engine
	templateEngine := templates.NewEngine("web/templates")
	if err := templateEngine.LoadTemplates(); err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}
	log.Println("✅ Templates loaded")

	// Initialize services
	contentService := services.NewContentService(mongoClient)

	// Initialize web handlers
	webHandlers := web.NewHandlers(contentService, templateEngine)

	// Setup Gin router
	router := gin.Default()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(corsMiddleware())
	router.Use(infrastructure.PerformanceMiddleware())

	// Static files
	router.Static("/static", "./web/static")

	// Health check endpoint with performance metrics
	router.GET("/health", func(c *gin.Context) {
		metrics := infrastructure.GetGlobalMetricsCollector().GetMetrics()
		cacheStats := infrastructure.GetGlobalCache().GetStats()
		
		c.JSON(http.StatusOK, gin.H{
			"status":       "healthy",
			"version":      "3.0.0-go",
			"architecture": "hexagonal",
			"database":     mongoClient.DatabaseName(),
			"performance": gin.H{
				"request_count":     metrics.RequestCount,
				"average_response":  metrics.AverageResponse.String(),
				"memory_usage_mb":   metrics.MemoryUsage / 1024 / 1024,
				"goroutine_count":   metrics.GoroutineCount,
				"active_connections": metrics.ActiveConns,
				"cache_hit_rate":    metrics.CacheHitRate,
				"cache_items":       cacheStats["item_count"],
			},
		})
	})

	// Register routes
	webHandlers.RegisterRoutes(router)

	// Setup HTTP server
	srv := &http.Server{
		Addr:    config.GetAddress(),
		Handler: router,
	}

	// Start performance monitoring
	monitorCtx, monitorCancel := context.WithCancel(context.Background())
	defer monitorCancel()
	infrastructure.StartPerformanceMonitoring(monitorCtx)

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Starting D&D 5e SRD Editor on %s", config.GetAddress())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down D&D 5e SRD Editor...")

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