package api

import (
	"github.com/deepanshumishra/devcost-api/internal/api/handlers"
	"github.com/deepanshumishra/devcost-api/internal/config"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, cfg *config.Config) {
	// Health check
	r.GET("/health", handlers.HealthCheck)

	// AWS IAM endpoints
	r.GET("/users", handlers.GetIAMUsersList(cfg))

	// AWS Resource endpoints
	r.GET("/getresourcesbytag", handlers.GetResourcesByTags(cfg))

	// Get Unused Resources
	r.GET("/resources/unused", handlers.GetUnusedResources(cfg))

	// Get Cost by Tag
	r.GET("/costs/tag", handlers.GetTagCosts(cfg))
}