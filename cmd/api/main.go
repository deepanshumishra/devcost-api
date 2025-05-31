package main

import (
	"log"

	"github.com/deepanshumishra/devcost-api/internal/api"
	"github.com/deepanshumishra/devcost-api/internal/config"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Gin router
	r := gin.Default()

	// Set trusted proxies (localhost for testing)
	if err := r.SetTrustedProxies([]string{"127.0.0.1", "::1"}); err != nil {
		log.Fatalf("Failed to set trusted proxies: %v", err)
	}

	// Setup routes
	api.SetupRoutes(r, cfg)

	// Start server
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}