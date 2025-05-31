package handlers

import (
	"github.com/deepanshumishra/devcost-api/internal/aws"
	"github.com/deepanshumishra/devcost-api/internal/config"
	"github.com/deepanshumishra/devcost-api/internal/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetProjectCosts(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try fetching costs from AWS Cost Explorer
		costs, err := aws.GetProjectCosts(cfg)
		if err != nil {
			// Return mock data if AWS call fails (e.g., no credentials)
			if err.Error() == "operation error Cost Explorer: GetCostAndUsage, https response error StatusCode: 400, RequestID: *, api error UnrecognizedClientException: The security token included in the request is invalid." {
				mockCosts := []models.ProjectCost{
					{Project: "dev-cluster", Cost: "100.50", Currency: "USD"},
					{Project: "prod-cluster", Cost: "200.75", Currency: "USD"},
				}
				c.JSON(http.StatusOK, gin.H{
					"projects": mockCosts,
					"warning":  "Using mock data due to invalid AWS credentials",
				})
				return
			}
			// Return error for other failures
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return real AWS data
		c.JSON(http.StatusOK, gin.H{
			"projects": costs,
		})
	}
}