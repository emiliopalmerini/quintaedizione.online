package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/emiliopalmerini/due-draghi-5e-srd/internal/infrastructure"
	"github.com/emiliopalmerini/due-draghi-5e-srd/pkg/mongodb"
)

func main() {
	// Load configuration
	config := infrastructure.LoadConfig()
	
	// Initialize MongoDB client
	mongoConfig := mongodb.Config{
		URI:         config.MongoURI,
		Database:    config.DatabaseName,
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
	
	// Setup Gin router
	if config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	
	r := gin.Default()
	
	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":       "healthy",
			"version":      "1.0.0-go",
			"architecture": "hexagonal",
			"database":     mongoClient.DatabaseName(),
		})
	})
	
	log.Printf("Starting D&D 5e SRD Editor (Go version) on %s", config.GetAddress())
	if err := r.Run(config.GetAddress()); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}