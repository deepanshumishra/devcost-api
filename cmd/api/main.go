package main

import (
	"github.com/deepanshumishra/devcost-api/internal/api"
	"github.com/deepanshumishra/devcost-api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize AWS config
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// Initialize Gin router
	r := gin.Default()

	// Setup routes with config
	api.SetupRoutes(r, cfg)

	// Run server on port 8080
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}