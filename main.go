package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/mishazigelboim/gocrawl/handlers"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title goCrawl API
// @version 1.0
// @description API for crawling AI model pages using dynamic crawl4ai pods
// @host localhost:8080
// @BasePath /
func main() {
	// Get namespace from environment or use default
	namespace := os.Getenv("POD_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	// Initialize handler
	crawlHandler, err := handlers.NewCrawlHandler(namespace)
	if err != nil {
		log.Fatalf("Failed to create crawl handler: %v", err)
	}

	// Set up Gin router
	r := gin.Default()

	// Health check endpoints
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	r.GET("/ready", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ready"})
	})

	// API routes
	v1 := r.Group("/api/v1")
	{
		v1.POST("/crawl", crawlHandler.CrawlModel)
	}

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting goCrawl server on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
