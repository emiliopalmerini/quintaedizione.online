package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// TODO: Initialize hexagonal architecture components
	
	r := gin.Default()
	
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":       "healthy",
			"version":      "1.0.0-go",
			"architecture": "hexagonal",
		})
	})
	
	log.Println("🚀 Starting D&D 5e SRD Editor (Go version) on :8000")
	if err := r.Run(":8000"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}