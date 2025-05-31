package handlers

import (
	"net/http"
	"strings"

	"github.com/deepanshumishra/devcost-api/internal/aws"
	"github.com/deepanshumishra/devcost-api/internal/config"
	"github.com/gin-gonic/gin"
)

// GetIAMUsersList returns a handler function that lists IAM usernames.
func GetIAMUsersList(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Fetch IAM usernames from AWS
		usernames, err := aws.ListIAMUsernames(cfg)
		if err != nil {
			// Return mock data if AWS credentials are invalid
			if strings.Contains(err.Error(), "UnrecognizedClientException") {
				mockUsernames := []string{"mock-user1", "mock-user2"}
				c.JSON(http.StatusOK, gin.H{
					"users":   mockUsernames,
					"warning": "Using mock data due to invalid AWS credentials",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return AWS usernames
		c.JSON(http.StatusOK, gin.H{
			"users": usernames,
		})
	}
}