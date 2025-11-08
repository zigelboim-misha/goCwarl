package main

import (
	"log"
	"os"
	"time"

	_ "github.com/mishazigelboim/gocrawl/docs"
	"github.com/mishazigelboim/gocrawl/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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

	// CORS middleware - allow all origins for Swagger UI
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check endpoints
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	r.GET("/ready", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ready"})
	})

	// Swagger documentation at root
	r.GET("/", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	v1 := r.Group("/api/v1")
	{
		v1.POST("/crawl", crawlHandler.CrawlModel)
	}

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
