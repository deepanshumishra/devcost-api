package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/deepanshumishra/devcost-api/internal/aws"
	"github.com/deepanshumishra/devcost-api/internal/config"
	"github.com/deepanshumishra/devcost-api/internal/models"
	"github.com/gin-gonic/gin"
)

// GetResourcesByTags returns a handler function that lists AWS resources by tag.
func GetResourcesByTags(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get query parameters
		tagKey := c.Query("tag_key")
		tagValue := c.Query("tag_value")
		if tagKey == "" || tagValue == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tag_key and tag_value are required"})
			return
		}

		// Fetch resources from AWS
		resources, err := aws.ListResourcesByTags(cfg, tagKey, tagValue)
		if err != nil {
			// Return mock data if AWS credentials are invalid
			if strings.Contains(err.Error(), "UnrecognizedClientException") {
				mockResources := []models.Resource{
					{
						ResourceARN:  "arn:aws:ec2:us-east-1:123456789012:instance/i-mock123",
						ResourceType: "ec2:instance",
						Tags:         map[string]string{"project": "dev-cluster"},
					},
					{
						ResourceARN:  "arn:aws:s3:::mock-bucket",
						ResourceType: "s3:bucket",
						Tags:         map[string]string{"project": "dev-cluster"},
					},
				}
				c.JSON(http.StatusOK, gin.H{
					"resources": mockResources,
					"warning":   "Using mock data due to invalid AWS credentials",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return AWS resources
		c.JSON(http.StatusOK, gin.H{
			"resources": resources,
		})
	}
}

// GetUnusedResources returns a handler function that lists unused AWS resources.
func GetUnusedResources(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get query parameters for time range
		startStr := c.Query("start") // e.g., "2025-05-01"
		endStr := c.Query("end")     // e.g., "2025-05-07"
		unusedForDaysStr := c.Query("unusedForDays")
		var start, end time.Time
		var err error

		// Default unusedForDays for Secrets Manager
		unusedForDays := 90
		if unusedForDaysStr != "" {
			unusedForDays, err = strconv.Atoi(unusedForDaysStr)
			if err != nil || unusedForDays < 1 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unusedForDays, must be a positive integer"})
				return
			}
		}

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

		// Fetch unused resources
		resources, err := aws.ListUnusedResources(c.Request.Context(), cfg, start, end, unusedForDays)
		if err != nil {
			// Return mock data if AWS call fails
			if strings.Contains(err.Error(), "UnrecognizedClientException") {
				mockResources := []models.UnusedResource{
					{
						ResourceType: "ec2:instance",
						ResourceID:   "i-mock123",
						Reason:       "CPU utilization <20% for 7 days",
					},
					{
						ResourceType: "ebs:volume",
						ResourceID:   "vol-mock456",
						Reason:       "Unattached",
					},
					{
						ResourceType: "rds:instance",
						ResourceID:   "db-mock789",
						Reason:       "Stopped",
					},
					{
						ResourceType: "ec2:elastic-ip",
						ResourceID:   "eipalloc-mock012",
						Reason:       "Not associated with any resource",
					},
					{
						ResourceType: "bedrock:knowledge-base",
						ResourceID:   "kb-mock345",
						Reason:       "No queries for 90 days",
					},
					{
						ResourceType: "lambda:function",
						ResourceID:   "arn:aws:lambda:us-east-1:123456789012:function:mock-function",
						Reason:       "No invocations for 90 days",
					},
					{
						ResourceType: "dynamodb:table",
						ResourceID:   "mock-table",
						Reason:       "No reads or writes for 90 days",
					},
					{
						ResourceType: "elasticloadbalancing:loadbalancer",
						ResourceID:   "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/mock-lb/...",
						Reason:       "No registered targets",
					},
				}
				c.JSON(http.StatusOK, gin.H{
					"unused_resources": mockResources,
					"warning":          "Using mock data due to invalid AWS credentials",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return unused resources
		c.JSON(http.StatusOK, gin.H{
			"unused_resources": resources,
		})
	}
}

// GetTagCosts returns a handler function that lists costs by a specified tag key.
func GetTagCosts(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get query parameters
		tagKey := c.Query("tag_key")
		startStr := c.Query("start") // e.g., "2025-05-01"
		endStr := c.Query("end")     // e.g., "2025-05-07"
		var start, end time.Time
		var err error

		// Validate tag_key
		if tagKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tag_key is required"})
			return
		}

		// Validate date range
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

		// Fetch tag costs
		costs, err := aws.GetTagCosts(cfg, tagKey, start, end)
		if err != nil {
			if strings.Contains(err.Error(), "UnrecognizedClientException") {
				mockCosts := []models.TagCost{
					{
						TagValue:    "dev-cluster",
						Cost:        100.50,
						Currency:    "USD",
						Resources:   []models.ResourceCost{
							{
								ResourceType: "ec2:instance",
								ResourceID:   "i-mock123",
								Cost:         50.25,
							},
							{
								ResourceType: "rds:instance",
								ResourceID:   "db-mock789",
								Cost:         50.25,
							},
						},
					},
				}
				c.JSON(http.StatusOK, gin.H{
					"tag_costs": mockCosts,
					"warning":   "Using mock data due to invalid AWS credentials",
				})
				return
			}
			if strings.Contains(err.Error(), "not active in cost allocation tags") {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return tag costs
		c.JSON(http.StatusOK, gin.H{
			"tag_costs": costs,
		})
	}
}