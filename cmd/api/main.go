package main

import (
	"github.com/deepanshumishra/devcost-api/internal/api"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Gin router
	r := gin.Default()

	// Setup routes
	api.SetupRoutes(r)

	// Run server on port 8080
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}