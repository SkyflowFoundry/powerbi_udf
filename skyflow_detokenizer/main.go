// main.go
//go:build server
// +build server

package main

import (
	"fmt"
	"log"
	"skyflow_detokenizer/internal/config"
	"skyflow_detokenizer/internal/handler"
	"skyflow_detokenizer/internal/logging"
	"skyflow_detokenizer/internal/processor"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration at startup
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize the logging system
	if err := logging.InitLogger(cfg.Logging.Level, cfg.Logging.LogToFile, cfg.Logging.FilePath); err != nil {
		log.Fatalf("Failed to initialize logging: %v", err)
	}
	defer logging.Close()

	// Initialize the token processor with our configuration
	p := processor.NewTokenProcessor(
		cfg.Skyflow.VaultURL,
		cfg.Skyflow.BearerToken,
		cfg.Skyflow.BatchSize,
		cfg.Skyflow.ParallelCalls,
		cfg.Skyflow.RequestTimeoutSecs,
	)

	// Configure Gin for production
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Set up CORS middleware to allow Power BI to call this service
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, Authorization")

		// Handle preflight OPTIONS request
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Set up our main endpoint
	r.POST("/detokenize", handler.DetokenizeHandler(p))

	// Add a health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Start the server
	serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	logging.Info("Starting server on %s", serverAddr)
	if err := r.Run(serverAddr); err != nil {
		logging.Error("Failed to start server: %v", err)
	}
}
