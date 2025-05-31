package api

import (
	"github.com/deepanshumishra/devcost-api/internal/api/handlers"
	"github.com/deepanshumishra/devcost-api/internal/aws"
	"github.com/deepanshumishra/devcost-api/internal/config"
	"github.com/gin-gonic/gin"
	"time"
)

func SetupRoutes(r *gin.Engine, cfg *config.Config) {
	// Health check
	r.GET("/health", handlers.HealthCheck)

	r.GET("/costs/projects", func(c *gin.Context) {
		// Parse start and end time from query parameters
		startStr := c.Query("start")
		endStr := c.Query("end")
		start, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid or missing 'start' query parameter"})
			return
		}
		end, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid or missing 'end' query parameter"})
			return
		}
		costs, err := aws.GetProjectCosts(cfg, start, end)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, costs)
	})

	// AWS IAM endpoints
	r.GET("/users", handlers.GetIAMUsersList(cfg))

	// AWS Resource endpoints
	r.GET("/getresourcesbytag", handlers.GetResourcesByTags(cfg))

	// Get Unused Resources
	r.GET("/resources/unused", handlers.GetUnusedResources(cfg))

	


}