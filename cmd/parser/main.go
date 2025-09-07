package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/adapters"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/adapters/web"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/application/services"
	"github.com/emiliopalmerini/due-draghi-5e-srd/pkg/mongodb"
	"github.com/emiliopalmerini/due-draghi-5e-srd/pkg/templates"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration from environment
	mongoURI := getEnv("MONGO_URI", "mongodb://localhost:27017")
	dbName := getEnv("DB_NAME", "dnd")
	port := getEnv("PORT", "8100")

	// Initialize MongoDB client
	mongoConfig := mongodb.Config{
		URI:         mongoURI,
		Database:    dbName,
		Timeout:     10 * time.Second,
		MaxPoolSize: 100,
	}

	mongoClient, err := mongodb.NewClient(mongoConfig)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer mongoClient.Close()

	// Test MongoDB connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := mongoClient.Ping(ctx); err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	// Initialize template engine
	templateEngine := templates.NewEngine("web/templates")
	if err := templateEngine.LoadTemplates(); err != nil {
		log.Fatal("Failed to load templates:", err)
	}
	log.Println("✅ Templates loaded")

	// Setup Gin router
	r := gin.Default()

	// Static files
	r.Static("/static", "./web/static")

	r.GET("/healthz", func(c *gin.Context) {
		c.String(200, "ok")
	})

	// Setup dependencies
	repo := adapters.NewMongoParserRepository(mongoClient)
	ingestService := services.NewIngestService(repo)
	ingestHandler := web.NewIngestHandler(ingestService, templateEngine, getEnv("INPUT_DIR", "data"))

	// Routes
	r.GET("/", ingestHandler.GetIndex)
	r.POST("/run", ingestHandler.PostRun)

	address := ":" + port
	log.Printf("Starting SRD Parser (Go version) on %s", address)
	if err := r.Run(address); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
