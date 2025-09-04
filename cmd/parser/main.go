package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/emiliopalmerini/due-draghi-5e-srd/pkg/mongodb"
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
	
	// Setup Gin router
	r := gin.Default()
	
	r.GET("/healthz", func(c *gin.Context) {
		c.String(200, "ok")
	})
	
	r.GET("/", func(c *gin.Context) {
		// TODO: Render parser form template
		c.JSON(200, gin.H{
			"message": "SRD Parser Web Interface (Go version)",
			"database": mongoClient.DatabaseName(),
		})
	})
	
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