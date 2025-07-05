package main

import (
	"log"
	"log/slog"
	"os"
	"github.com/deepanshumishra/devcost-api/internal/api"
	"github.com/deepanshumishra/devcost-api/internal/config"
	"github.com/gin-gonic/gin"
)

func main() {

	// Set up structured JSON logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Gin router
	r := gin.Default()

	// Set trusted proxies (update for production, e.g., ALB)
	if err := r.SetTrustedProxies([]string{"127.0.0.1", "::1"}); err != nil {
		slog.Error("Failed to set trusted proxies", "error", err)
		os.Exit(1)
	}

	

	// Setup routes
	api.SetupRoutes(r, cfg)

	// Start server
	slog.Info("Starting server", "port", 8080)
	if err := r.Run(":8080"); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}