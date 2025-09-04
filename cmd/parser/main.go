package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// TODO: Initialize parser components
	
	r := gin.Default()
	
	r.GET("/healthz", func(c *gin.Context) {
		c.String(200, "ok")
	})
	
	r.GET("/", func(c *gin.Context) {
		// TODO: Render parser form template
		c.JSON(200, gin.H{
			"message": "SRD Parser Web Interface (Go version)",
		})
	})
	
	log.Println("🚀 Starting SRD Parser (Go version) on :8100")
	if err := r.Run(":8100"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}