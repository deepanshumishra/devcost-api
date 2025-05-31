package handlers

import (
	"net/http"
	"strings"

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