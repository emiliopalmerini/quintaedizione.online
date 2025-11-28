package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/emiliopalmerini/quintaedizione.online/internal/adapters/repositories"
	web "github.com/emiliopalmerini/quintaedizione.online/internal/adapters/web"
	"github.com/emiliopalmerini/quintaedizione.online/internal/application/filters"
	"github.com/emiliopalmerini/quintaedizione.online/internal/application/parsers"
	"github.com/emiliopalmerini/quintaedizione.online/internal/application/services"
	"github.com/emiliopalmerini/quintaedizione.online/internal/infrastructure"
	"github.com/emiliopalmerini/quintaedizione.online/internal/infrastructure/database"
	pkgMongodb "github.com/emiliopalmerini/quintaedizione.online/pkg/mongodb"
	"github.com/emiliopalmerini/quintaedizione.online/pkg/templates"
	"github.com/gin-gonic/gin"
)

func main() {

	config := infrastructure.LoadConfig()

	if config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := mongoClient.Ping(ctx); err != nil {
		log.Fatalf("MongoDB health check failed: %v", err)
	}
	log.Println("MongoDB connection established")

	indexManager := database.NewIndexManager(mongoClient)

	repositoryFactory := repositories.NewRepositoryFactory(mongoClient)

	log.Println("Parsing markdown files...")
	if err := parseMarkdownFiles(repositoryFactory, indexManager); err != nil {
		log.Fatalf("Failed to parse markdown files: %v", err)
	}

	var templateEngine *templates.TemplEngine
	if config.IsProduction() {
		templateEngine = templates.NewTemplEngine()
	} else {
		templateEngine = templates.NewDevTemplEngine()
	}
	log.Println("Templates loaded")

	filterRegistry, err := filters.NewYAMLFilterRegistry("configs/filters.yaml")
	if err != nil {
		log.Fatalf("Failed to initialize filter registry: %v", err)
	}
	log.Println("Filter registry loaded")

	filterService := services.NewFilterService(filterRegistry)

	contentService := services.NewContentService(repositoryFactory.DocumentRepository(), filterService)

	webHandlers := web.NewHandlers(contentService, templateEngine)

	router := gin.Default()

	router.Use(web.RequestLoggingMiddleware())
	router.Use(web.MetricsMiddleware())
	router.Use(webHandlers.ErrorRecoveryMiddleware())
	router.Use(web.SecurityMiddleware())
	router.Use(web.ValidationMiddleware())
	router.Use(corsMiddleware())

	router.Static("/static", "./web/static")

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

	router.GET("/admin/metrics", func(c *gin.Context) {
		metrics := web.GetGlobalMetrics()
		c.JSON(http.StatusOK, metrics.ToJSON())
	})

	webHandlers.RegisterRoutes(router)

	srv := &http.Server{
		Addr:    config.GetAddress(),
		Handler: router,
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

func parseMarkdownFiles(repositoryFactory *repositories.RepositoryFactory, indexManager *database.IndexManager) error {
	ctx := context.Background()

	log.Println("Dropping existing collections for clean parse...")
	collections := []string{
		"incantesimi", "mostri", "classi", "backgrounds", "equipaggiamenti",
		"oggetti_magici", "armi", "armature", "talenti", "servizi",
		"strumenti", "animali", "regole", "cavalcature_veicoli",
	}

	repo := repositoryFactory.DocumentRepository()
	for _, collection := range collections {
		if err := repo.DropCollection(ctx, collection); err != nil {
			log.Printf("Warning: Failed to drop collection %s: %v", collection, err)
		}
	}
	log.Println("Collections dropped")

	indexCtx, indexCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer indexCancel()
	if err := indexManager.EnsureIndexes(indexCtx); err != nil {
		log.Printf("Warning: Failed to recreate indexes: %v", err)

	} else {
		log.Println("✅ Indexes recreated")
	}

	documentRegistry, err := parsers.CreateDocumentRegistry()
	if err != nil {
		return fmt.Errorf("failed to create document registry: %w", err)
	}

	parserService := services.NewParserService(services.ParserServiceConfig{
		DocumentRegistry: documentRegistry,
		DocumentRepo:     repositoryFactory.DocumentRepository(),
		WorkItems:        nil,
		Logger:           parsers.NewConsoleLogger("parser"),
		DryRun:           false,
	})

	result, err := parserService.ParseAllFiles(ctx, "data/ita/lists")
	if err != nil {
		return err
	}

	log.Printf("Parsing completed: %d files, %d documents in %.2fs\n",
		result.SuccessCount, result.TotalDocuments, result.Duration.Seconds())

	return nil
}

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
