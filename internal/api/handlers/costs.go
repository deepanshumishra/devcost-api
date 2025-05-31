package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/deepanshumishra/devcost-api/internal/aws"
	"github.com/deepanshumishra/devcost-api/internal/config"
	"github.com/deepanshumishra/devcost-api/internal/models"
	"github.com/gin-gonic/gin"
)

func GetProjectCosts(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get query parameters for time range
		startStr := c.Query("start") // e.g., "2025-05-01"
		endStr := c.Query("end")     // e.g., "2025-05-07"
		var start, end time.Time
		var err error

		// Validate query parameters
		if startStr != "" && endStr != "" {
			start, err = time.Parse("2006-01-02", startStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format, use YYYY-MM-DD"})
				return
			}
			end, err = time.Parse("2006-01-02", endStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format, use YYYY-MM-DD"})
				return
			}
			if end.Before(start) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "End date must be after start date"})
				return
			}
		} else if startStr == "" && endStr == "" {
			// Default: last 7 days
			end = time.Now()
			start = end.AddDate(0, 0, -7)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Both start and end dates must be provided together"})
			return
		}

		// Fetch costs from AWS Cost Explorer
		costs, err := aws.GetProjectCosts(cfg, start, end)
		if err != nil {
			// Return mock data if AWS call fails
			if strings.Contains(err.Error(), "UnrecognizedClientException") {
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return real AWS data
		c.JSON(http.StatusOK, gin.H{
			"projects": costs,
		})
	}
}