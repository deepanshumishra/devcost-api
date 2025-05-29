package api

import (
	"github.com/deepanshumishra/devcost-api/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// Health check endpoint
	r.GET("/health", handlers.HealthCheck)

	// TODO: Add endpoints for DevCost features
	// r.GET("/costs/projects", handlers.GetProjectCosts)
	// r.GET("/resources/unused", handlers.GetUnusedResources)
	// r.POST("/slack/summary", handlers.SendSlackSummary)
}