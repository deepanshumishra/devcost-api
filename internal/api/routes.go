package api

import (
	"github.com/deepanshumishra/devcost-api/internal/api/handlers"
	"github.com/deepanshumishra/devcost-api/internal/config"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, cfg *config.Config) {
	// Health check
	r.GET("/health", handlers.HealthCheck)

	// Cost endpoints
	r.GET("/costs/projects", handlers.GetProjectCosts(cfg))

	// AWS IAM endpoints
	r.GET("/users", handlers.GetIAMUsersList(cfg))

	// AWS Resource endpoints
	r.GET("/getresourcesbytag", handlers.GetResourcesByTags(cfg))
}